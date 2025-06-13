package buffer

import (
	"bytes"
	"io"
)

// LineBuffer manages buffered reading and line extraction from files
type LineBuffer struct {
	data      []byte
	lineStart int
	lineEnd   int
	discard   bool
	bufSize   int64
}

// NewLineBuffer creates a new line buffer with the specified buffer size
func NewLineBuffer(bufSize int64) *LineBuffer {
	return &LineBuffer{
		data:      make([]byte, bufSize),
		lineStart: -1,
		lineEnd:   0,
		discard:   true,
		bufSize:   bufSize,
	}
}

// Reset resets the buffer state
func (lb *LineBuffer) Reset() {
	lb.lineStart = -1
	lb.lineEnd = 0
	lb.discard = true
}

// IsDiscarded returns true if the buffer is in discard state
func (lb *LineBuffer) IsDiscarded() bool {
	return lb.discard
}

// GetLineStart returns the current line start position
func (lb *LineBuffer) GetLineStart() int {
	return lb.lineStart
}

// GetLineEnd returns the current line end position
func (lb *LineBuffer) GetLineEnd() int {
	return lb.lineEnd
}

// GetData returns the buffer data
func (lb *LineBuffer) GetData() []byte {
	return lb.data
}

// SetLineStart sets the line start position
func (lb *LineBuffer) SetLineStart(pos int) {
	lb.lineStart = pos
}

// SetLineEnd sets the line end position
func (lb *LineBuffer) SetLineEnd(pos int) {
	lb.lineEnd = pos
}

// SetDiscard sets the discard state
func (lb *LineBuffer) SetDiscard(discard bool) {
	lb.discard = discard
}

// ReadLine reads a complete line from the reader at the given offset
func (lb *LineBuffer) ReadLine(reader io.ReaderAt, offset int64) ([]byte, error) {
	lb.data = lb.data[:lb.bufSize]
	lb.lineStart = -1
	lb.lineEnd = 0

	// Read initial buffer
	n, err := reader.ReadAt(lb.data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}

	lb.data = lb.data[:n]
	lb.discard = false

	// Find the start of the first complete line
	if offset == 0 {
		lb.lineStart = 0
	} else {
		// Find first newline to skip partial line at beginning
		firstNewline := bytes.IndexByte(lb.data, '\n')
		if firstNewline >= 0 {
			lb.lineStart = firstNewline + 1
		} else {
			// No complete line in buffer
			return nil, io.EOF
		}
	}

	// Find end of line
	cursor := bytes.IndexByte(lb.data[lb.lineStart:], '\n')
	if cursor >= 0 {
		lb.lineEnd = lb.lineStart + cursor
		return lb.data[lb.lineStart:lb.lineEnd], nil
	}

	// Line extends beyond buffer, try to extend
	if int64(len(lb.data)) < lb.bufSize*4 {
		// Extend buffer and try to read more
		lb.data = append(lb.data, make([]byte, lb.bufSize)...)
		moreN, moreErr := reader.ReadAt(lb.data[n:], offset+int64(n))
		if moreErr != nil && moreErr != io.EOF {
			return nil, moreErr
		}

		lb.data = lb.data[:n+moreN]
		cursor = bytes.IndexByte(lb.data[lb.lineStart:], '\n')
		if cursor >= 0 {
			lb.lineEnd = lb.lineStart + cursor
			return lb.data[lb.lineStart:lb.lineEnd], nil
		}
	}

	// Return what we have (line without newline)
	lb.lineEnd = len(lb.data)
	if lb.lineStart < lb.lineEnd {
		return lb.data[lb.lineStart:lb.lineEnd], nil
	}

	return nil, io.EOF
}

// NextLine returns the next line from the current buffer
func (lb *LineBuffer) NextLine() ([]byte, error) {
	if lb.discard {
		return nil, io.EOF
	}

	lb.lineStart = lb.lineEnd + 1

	// Check bounds before accessing slice
	if lb.lineStart >= len(lb.data) {
		return nil, io.EOF
	}

	cursor := bytes.IndexByte(lb.data[lb.lineStart:], '\n')
	if cursor > 0 {
		lb.lineEnd = lb.lineStart + cursor
		return lb.data[lb.lineStart:lb.lineEnd], nil
	}
	return nil, io.EOF
}

// FindLastLine searches for the last complete line in a buffer read from the given offset
func (lb *LineBuffer) FindLastLine(reader io.ReaderAt, offset int64) ([]byte, error) {
	count, err := reader.ReadAt(lb.data, offset)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if count == 0 {
		return nil, io.EOF
	}

	// Find the last newline in the buffer
	lastNewline := bytes.LastIndexByte(lb.data[:count], '\n')
	if lastNewline == -1 {
		// No newlines found
		return nil, io.EOF
	}

	// Find the second-to-last newline to get the start of the last line
	if lastNewline == 0 {
		// Only one character and it's a newline
		return nil, io.EOF
	}

	secondLastNewline := bytes.LastIndexByte(lb.data[:lastNewline], '\n')
	if secondLastNewline == -1 {
		// Last line starts from beginning of buffer
		lb.lineStart = 0
	} else {
		// Last line starts after the second-to-last newline
		lb.lineStart = secondLastNewline + 1
	}

	lb.lineEnd = lastNewline

	if lb.lineStart < lb.lineEnd {
		return lb.data[lb.lineStart:lb.lineEnd], nil
	}

	return nil, io.EOF
}
