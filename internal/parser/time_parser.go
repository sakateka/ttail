package parser

import (
	"regexp"
	"time"
)

// TimeParser handles parsing timestamps from log lines
type TimeParser struct {
	timeRe     *regexp.Regexp
	timeLayout string
	location   *time.Location
	// Reusable buffers to avoid allocations
	submatchBuf [][]byte
	stringBuf   []byte
	// Fast path optimization
	isTSKV bool
	isISO  bool
}

// NewTimeParser creates a new time parser with the given regex and layout
func NewTimeParser(timeRe *regexp.Regexp, timeLayout string, location *time.Location) *TimeParser {
	// Detect common patterns for fast path optimization
	pattern := timeRe.String()
	isTSKV := pattern == `\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`
	isISO := timeLayout == "2006-01-02T15:04:05"

	return &TimeParser{
		timeRe:      timeRe,
		timeLayout:  timeLayout,
		location:    location,
		submatchBuf: make([][]byte, 0, 4), // Pre-allocate for typical submatch count
		stringBuf:   make([]byte, 0, 64),  // Pre-allocate for typical timestamp length
		isTSKV:      isTSKV,
		isISO:       isISO,
	}
}

// ParseTime extracts and parses a timestamp from a log line
// Returns the parsed time and a boolean indicating if parsing was successful
func (tp *TimeParser) ParseTime(line []byte) (time.Time, bool) {
	// Fast path for TSKV format (most common case)
	if tp.isTSKV && tp.isISO {
		return tp.parseTSKVFast(line)
	}

	// Use FindSubmatchIndex to avoid allocating submatch slices
	indices := tp.timeRe.FindSubmatchIndex(line)
	if indices == nil || len(indices) < 4 {
		return time.Time{}, false // No timestamp found or no capture group
	}

	// Extract the first capture group without string conversion
	start, end := indices[2], indices[3]
	if start < 0 || end < 0 {
		return time.Time{}, false
	}

	// Reuse string buffer to avoid allocation
	tp.stringBuf = tp.stringBuf[:0]
	tp.stringBuf = append(tp.stringBuf, line[start:end]...)

	tm, err := time.ParseInLocation(tp.timeLayout, string(tp.stringBuf), tp.location)
	if err != nil {
		return time.Time{}, false
	}

	return tm, true
}

// parseTSKVFast provides a fast path for TSKV timestamp parsing without regex
func (tp *TimeParser) parseTSKVFast(line []byte) (time.Time, bool) {
	// Look for "\ttimestamp=" pattern
	const prefix = "\ttimestamp="
	const timestampLen = 19 // "2006-01-02T15:04:05"

	idx := -1
	for i := 0; i <= len(line)-len(prefix); i++ {
		if string(line[i:i+len(prefix)]) == prefix {
			idx = i + len(prefix)
			break
		}
	}

	if idx == -1 || idx+timestampLen > len(line) {
		return time.Time{}, false
	}

	// Check if we have enough characters and next char is tab
	if idx+timestampLen >= len(line) || line[idx+timestampLen] != '\t' {
		return time.Time{}, false
	}

	// Extract timestamp directly without regex
	timestampBytes := line[idx : idx+timestampLen]

	// Reuse string buffer
	tp.stringBuf = tp.stringBuf[:0]
	tp.stringBuf = append(tp.stringBuf, timestampBytes...)

	tm, err := time.ParseInLocation(tp.timeLayout, string(tp.stringBuf), tp.location)
	if err != nil {
		return time.Time{}, false
	}

	return tm, true
}

// ParseTimePtr extracts and parses a timestamp from a log line (legacy interface)
func (tp *TimeParser) ParseTimePtr(line []byte) (*time.Time, error) {
	tm, ok := tp.ParseTime(line)
	if !ok {
		return nil, nil
	}
	return &tm, nil
}

// GetRegex returns the compiled regex pattern
func (tp *TimeParser) GetRegex() *regexp.Regexp {
	return tp.timeRe
}

// GetLayout returns the time layout string
func (tp *TimeParser) GetLayout() string {
	return tp.timeLayout
}

// GetLocation returns the time location
func (tp *TimeParser) GetLocation() *time.Location {
	return tp.location
}
