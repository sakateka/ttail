package searcher

import (
	"io"
	"os"
	"time"

	"github.com/sakateka/ttail/internal/buffer"
	"github.com/sakateka/ttail/internal/config"
	"github.com/sakateka/ttail/internal/parser"
)

// TimeSearcher provides high-performance time-based file searching
type TimeSearcher struct {
	file     *os.File
	parser   *parser.TimeParser
	buffer   *buffer.LineBuffer
	opts     config.Options
	fromTime time.Time
	offset   int64
	size     int64
}

// NewTimeSearcher creates a new time searcher with optimized defaults
func NewTimeSearcher(file *os.File, opts config.Options) *TimeSearcher {
	// Get file size once for efficiency
	size, _ := file.Seek(0, io.SeekEnd)

	return &TimeSearcher{
		file:     file,
		parser:   parser.NewTimeParser(opts.TimeRe, opts.TimeLayout, opts.Location),
		buffer:   buffer.NewLineBuffer(opts.BufSize),
		opts:     opts,
		fromTime: time.Now(),
		size:     size,
	}
}

// SetFromTime sets the reference time for searching
func (ts *TimeSearcher) SetFromTime(t time.Time) {
	ts.fromTime = t
}

// GetOffset returns the current file offset
func (ts *TimeSearcher) GetOffset() int64 {
	return ts.offset
}

// findLastLineTime efficiently finds the timestamp of the last line
func (ts *TimeSearcher) findLastLineTime() (time.Time, error) {
	offset := ts.size - ts.opts.BufSize
	if offset < 0 {
		offset = 0
	}

	for step := ts.opts.StepsLimit; offset >= 0 && step > 0; step-- {
		line, err := ts.buffer.FindLastLine(ts.file, offset)
		if err != nil && err != io.EOF {
			continue
		}

		if len(line) > 0 {
			if tm, err := ts.parser.ParseTime(line); err == nil && tm != nil {
				ts.offset = offset
				return *tm, nil
			}
		}

		// Move to previous buffer position
		if offset > 0 && offset < ts.opts.BufSize {
			offset = 0
		} else {
			offset -= ts.opts.BufSize
		}
	}

	return time.Time{}, io.EOF
}

// findTimeAtOffset finds the first valid timestamp at or after the given offset
func (ts *TimeSearcher) findTimeAtOffset(offset int64) (*time.Time, error) {
	ts.offset = offset

	for {
		line, err := ts.buffer.ReadLine(ts.file, ts.offset)
		if err != nil {
			return nil, err
		}

		if len(line) == 0 {
			ts.offset += int64(ts.buffer.GetLineEnd())
			continue
		}

		if tm, err := ts.parser.ParseTime(line); err == nil && tm != nil {
			return tm, nil
		}

		// Try next line in buffer
		if nextLine, err := ts.buffer.NextLine(); err == nil && len(nextLine) > 0 {
			if tm, err := ts.parser.ParseTime(nextLine); err == nil && tm != nil {
				return tm, nil
			}
		}

		// Move to next buffer
		ts.offset += int64(ts.buffer.GetLineEnd())
	}
}

// preciseFindTime performs precise time searching within a buffer
func (ts *TimeSearcher) preciseFindTime() error {
	for {
		line, err := ts.buffer.NextLine()
		if err == io.EOF {
			ts.offset += int64(ts.buffer.GetLineEnd())
			line, err = ts.buffer.ReadLine(ts.file, ts.offset)
		}
		if err != nil {
			return err
		}

		if tm, err := ts.parser.ParseTime(line); err == nil && tm != nil {
			if ts.fromTime.Sub(*tm) <= ts.opts.Duration {
				break
			}
		}
	}
	return nil
}

// FindPosition performs binary search to find the optimal starting position
func (ts *TimeSearcher) FindPosition() error {
	// Handle empty files
	if ts.size == 0 {
		ts.offset = 0
		return nil
	}

	// Handle time from last line option
	if ts.opts.TimeFromLastLine {
		if lastTime, err := ts.findLastLineTime(); err == nil && !lastTime.IsZero() {
			ts.fromTime = lastTime.Add(-ts.opts.Duration)
		} else {
			// If no timestamp found, start from beginning
			ts.offset = 0
			return nil
		}
	} else {
		ts.fromTime = time.Now().Add(-ts.opts.Duration)
	}

	// Binary search for optimal position
	var up, down int64 = 0, ts.size

	for (down - up) > ts.opts.BufSize {
		middle := up + (down-up)/2 // Prevent overflow

		tm, err := ts.findTimeAtOffset(middle)
		if err != nil {
			if err == io.EOF {
				// If we hit EOF, try from beginning
				ts.offset = 0
				return nil
			}
			return err
		}

		if tm.Before(ts.fromTime) {
			up = middle
		} else {
			down = middle
		}
	}

	// Fine-tune the position
	ts.offset = up
	ts.buffer.Reset()

	// Try to find precise position, but don't fail if we can't
	if err := ts.preciseFindTime(); err != nil && err != io.EOF {
		return err
	}

	ts.offset += int64(ts.buffer.GetLineStart())
	return nil
}

// CopyTo efficiently copies from the found position to the writer
func (ts *TimeSearcher) CopyTo(w io.Writer) (int64, error) {
	if _, err := ts.file.Seek(ts.offset, io.SeekStart); err != nil {
		return 0, err
	}

	return io.Copy(w, ts.file)
}

// GetReader returns a reader positioned at the found offset
func (ts *TimeSearcher) GetReader() (io.Reader, error) {
	if _, err := ts.file.Seek(ts.offset, io.SeekStart); err != nil {
		return nil, err
	}
	return ts.file, nil
}
