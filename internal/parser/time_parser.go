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
}

// NewTimeParser creates a new time parser with the given regex and layout
func NewTimeParser(timeRe *regexp.Regexp, timeLayout string, location *time.Location) *TimeParser {
	return &TimeParser{
		timeRe:     timeRe,
		timeLayout: timeLayout,
		location:   location,
	}
}

// ParseTime extracts and parses a timestamp from a log line
func (tp *TimeParser) ParseTime(line []byte) (*time.Time, error) {
	subm := tp.timeRe.FindSubmatch(line)
	if subm == nil {
		return nil, nil // No timestamp found
	}

	tm, err := time.ParseInLocation(tp.timeLayout, string(subm[1]), tp.location)
	if err != nil {
		return nil, err
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
