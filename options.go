package ttail

import (
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
)

// DefaultConfigFile for ttail
var DefaultConfigFile = "/etc/ttail/types.toml"

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

// WithBufSize set buffer size for random reads
func WithBufSize(size int64) TimeFileOptions {
	return func(o *options) {
		o.bufSize = size
	}
}

// WithStepsLimit set number of attempts for lastLineTime
func WithStepsLimit(steps int) TimeFileOptions {
	return func(o *options) {
		o.stepsLimit = steps
	}
}

// WithTimeReAsStr compile string to regexp for time search
func WithTimeReAsStr(timeRe string) TimeFileOptions {
	re := regexp.MustCompile(timeRe)
	return func(o *options) {
		o.timeRe = re
	}
}

// WithTimeLayout set expected time layout for time.Parse
func WithTimeLayout(layout string) TimeFileOptions {
	return func(o *options) {
		o.timeLayout = layout
	}
}

// Config for ttail
type Config map[string]Type

// Type of log
type Type struct {
	BufSize    int64
	StepsLimit int
	TimeReStr  string
	TimeLayout string
}

// OptionsFromConfig convert config to options list
func OptionsFromConfig(logType string) ([]TimeFileOptions, error) {
	if _, err := os.Stat(DefaultConfigFile); os.IsNotExist(err) {
		return nil, errors.New("Config file does not exist")
	} else if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.DecodeFile(DefaultConfigFile, &conf); err != nil {
		return nil, err
	}
	aType, ok := conf[logType]
	if !ok {
		return nil, errors.New("Failed to find options for log type: " + logType)
	}
	var opts []TimeFileOptions
	if aType.BufSize != 0 {
		opts = append(opts, WithBufSize(aType.BufSize))
	}

	if aType.StepsLimit != 0 {
		opts = append(opts, WithStepsLimit(aType.StepsLimit))
	}

	if aType.TimeReStr != "" {
		opts = append(opts, WithTimeReAsStr(aType.TimeReStr))
	}

	if aType.TimeLayout != "" {
		opts = append(opts, WithTimeLayout(aType.TimeLayout))
	}
	return opts, nil
}
