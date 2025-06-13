package ttail

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sakateka/ttail/internal/config"
	"github.com/sakateka/ttail/internal/searcher"
)

// FlagDebug enable debug output
var FlagDebug bool

// TFile represents a time-searchable file with optimized performance
type TFile struct {
	searcher *searcher.TimeSearcher
	opts     config.Options
	file     *os.File
}

func debug(format string, args ...interface{}) {
	if FlagDebug {
		fmt.Fprintf(os.Stderr, ">>> "+format+"\n", args...)
	}
}

// NewTimeFile creates a new time-searchable file instance with options
func NewTimeFile(f *os.File, opt ...TimeFileOptions) *TFile {
	opts := config.DefaultOptions()
	for _, o := range opt {
		o(&opts)
	}

	debug("NewTimeFile: with options %+v", opts)

	return &TFile{
		searcher: searcher.NewTimeSearcher(f, opts),
		opts:     opts,
		file:     f,
	}
}

// FindPosition searches for the optimal starting position in the file
func (t *TFile) FindPosition() error {
	return t.searcher.FindPosition()
}

// CopyTo copies content from the found position to the writer
func (t *TFile) CopyTo(w io.Writer) (int64, error) {
	debug("[CopyTo]: Copy file from offset=%d", t.searcher.GetOffset())
	copied, err := t.searcher.CopyTo(w)
	if err != nil {
		debug("[CopyTo]: Copy only %d bytes: %s", copied, err)
	}
	return copied, err
}

// GetReader returns a reader positioned at the found offset
func (t *TFile) GetReader() (io.Reader, error) {
	return t.searcher.GetReader()
}

// GetOffset returns the current file offset
func (t *TFile) GetOffset() int64 {
	return t.searcher.GetOffset()
}

// SetFromTime sets the reference time for searching
func (t *TFile) SetFromTime(tm time.Time) {
	t.searcher.SetFromTime(tm)
}
