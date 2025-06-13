package ttail

import (
	"regexp"
	"time"

	"github.com/sakateka/ttail/internal/config"
)

// TimeFileOptions represents a configuration function for backward compatibility
type TimeFileOptions func(*config.Options)

// WithDuration sets the tail time span
func WithDuration(t time.Duration) TimeFileOptions {
	return func(o *config.Options) {
		o.Duration = t
	}
}

// WithTimeFromLastLine determines where to take time for tail time span
func WithTimeFromLastLine(timeFromLastLine bool) TimeFileOptions {
	return func(o *config.Options) {
		o.TimeFromLastLine = timeFromLastLine
	}
}

// WithBufSize sets buffer size for random reads
func WithBufSize(size int64) TimeFileOptions {
	return func(o *config.Options) {
		o.BufSize = size
	}
}

// WithStepsLimit sets number of attempts for lastLineTime
func WithStepsLimit(steps int) TimeFileOptions {
	return func(o *config.Options) {
		o.StepsLimit = steps
	}
}

// WithTimeReAsStr compiles string to regexp for time search
func WithTimeReAsStr(timeRe string) TimeFileOptions {
	re := regexp.MustCompile(timeRe)
	return func(o *config.Options) {
		o.TimeRe = re
	}
}

// WithTimeLayout sets expected time layout for time.Parse
func WithTimeLayout(layout string) TimeFileOptions {
	return func(o *config.Options) {
		o.TimeLayout = layout
	}
}

// OptionsFromConfig converts config to options list for backward compatibility
func OptionsFromConfig(logType string, configPath string) ([]TimeFileOptions, error) {
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	lt, err := conf.GetLogTypeOptions(logType)
	if err != nil {
		return nil, err
	}

	// Pre-allocate slice to avoid reallocations
	options := make([]TimeFileOptions, 0, 4)
	if lt.BufSize != 0 {
		options = append(options, WithBufSize(lt.BufSize))
	}
	if lt.StepsLimit != 0 {
		options = append(options, WithStepsLimit(lt.StepsLimit))
	}
	if lt.TimeReStr != "" {
		options = append(options, WithTimeReAsStr(lt.TimeReStr))
	}
	if lt.TimeLayout != "" {
		options = append(options, WithTimeLayout(lt.TimeLayout))
	}

	return options, nil
}

// OptionsFromConfigDirect applies config directly to options to avoid slice allocations
func OptionsFromConfigDirect(logType string, configPath string, opts *config.Options) error {
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	lt, err := conf.GetLogTypeOptions(logType)
	if err != nil {
		return err
	}

	return lt.ApplyToOptions(opts)
}
