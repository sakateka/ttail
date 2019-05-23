package ttail

import (
	"regexp"
	"time"
)

type options struct {
	location         *time.Location
	duration         time.Duration
	bufSize          int64
	stepsLimit       int
	timeRe           *regexp.Regexp
	timeLayout       string
	timeFromLastLine bool
}

// TimeFileOptions set ttail options, duration, time re and layout, bufSize...
type TimeFileOptions func(*options)

var defaultOptions = options{
	location:   time.Local,
	bufSize:    1 << 14, // 16kb
	stepsLimit: 1024,
	timeRe:     regexp.MustCompile(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`),
	timeLayout: "2006-01-02T15:04:05",
}

// WithDuration set tail time span
func WithDuration(t time.Duration) TimeFileOptions {
	return func(o *options) {
		o.duration = t
	}
}

// WithTimeFromLastLine determines where to take time for tail time span
func WithTimeFromLastLine(timeFromLastLine bool) TimeFileOptions {
	return func(o *options) {
		o.timeFromLastLine = timeFromLastLine
	}
}
