package config

import (
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
)

// DefaultConfigFile for ttail
const DefaultConfigFile = "/etc/ttail/types.toml"

// Config represents the configuration file structure
type Config map[string]LogType

// LogType defines configuration for a specific log type
type LogType struct {
	BufSize    int64  `toml:"buf_size"`
	StepsLimit int    `toml:"steps_limit"`
	TimeReStr  string `toml:"time_regex"`
	TimeLayout string `toml:"time_layout"`
}

// Options holds all configuration options for ttail
type Options struct {
	Location         *time.Location
	Duration         time.Duration
	BufSize          int64
	StepsLimit       int
	TimeRe           *regexp.Regexp
	TimeLayout       string
	TimeFromLastLine bool
}

// DefaultOptions returns the default configuration
func DefaultOptions() Options {
	return Options{
		Location:   time.Local,
		BufSize:    1 << 14, // 16KB - optimized for most systems
		StepsLimit: 1024,
		TimeRe:     regexp.MustCompile(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`),
		TimeLayout: "2006-01-02T15:04:05",
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configFile string) (Config, error) {
	if configFile == "" {
		configFile = DefaultConfigFile
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.New("config file does not exist: " + configFile)
	} else if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// GetLogTypeOptions returns options for a specific log type
func (c Config) GetLogTypeOptions(logType string) (*LogType, error) {
	lt, ok := c[logType]
	if !ok {
		return nil, errors.New("failed to find options for log type: " + logType)
	}
	return &lt, nil
}

// ApplyToOptions applies log type configuration to base options
func (lt *LogType) ApplyToOptions(opts *Options) error {
	if lt.BufSize != 0 {
		opts.BufSize = lt.BufSize
	}

	if lt.StepsLimit != 0 {
		opts.StepsLimit = lt.StepsLimit
	}

	if lt.TimeReStr != "" {
		re, err := regexp.Compile(lt.TimeReStr)
		if err != nil {
			return err
		}
		opts.TimeRe = re
	}

	if lt.TimeLayout != "" {
		opts.TimeLayout = lt.TimeLayout
	}

	return nil
}

// Clone creates a deep copy of options (for thread safety)
func (opts *Options) Clone() Options {
	return Options{
		Location:         opts.Location,
		Duration:         opts.Duration,
		BufSize:          opts.BufSize,
		StepsLimit:       opts.StepsLimit,
		TimeRe:           opts.TimeRe, // regexp.Regexp is safe for concurrent use
		TimeLayout:       opts.TimeLayout,
		TimeFromLastLine: opts.TimeFromLastLine,
	}
}
