package config

import (
	"errors"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
	"maps"
)

// DefaultConfigFile for ttail
const DefaultConfigFile = "/etc/ttail/types.toml"

// Config represents the configuration file structure
type Config map[string]LogType

// LogType defines configuration for a specific log type
type LogType struct {
	BufSize    int64  `toml:"bufSize"`
	StepsLimit int    `toml:"stepsLimit"`
	TimeReStr  string `toml:"timeReStr"`
	TimeLayout string `toml:"timeLayout"`
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

// BuiltinLogTypes contains predefined log format configurations
var BuiltinLogTypes = Config{
	"tskv": LogType{
		TimeReStr:  `\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"kern": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"apache": LogType{
		TimeReStr:  `\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s`,
		TimeLayout: "02/Jan/2006:15:04:05",
	},
	"apache_common": LogType{
		TimeReStr:  `\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s`,
		TimeLayout: "02/Jan/2006:15:04:05",
	},
	"apache_combined": LogType{
		TimeReStr:  `\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s`,
		TimeLayout: "02/Jan/2006:15:04:05",
	},
	"nginx": LogType{
		TimeReStr:  `\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s`,
		TimeLayout: "02/Jan/2006:15:04:05",
	},
	"nginx_iso": LogType{
		TimeReStr:  `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"java": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02 15:04:05",
	},
	"java_iso": LogType{
		TimeReStr:  `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"python": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}),\d+`,
		TimeLayout: "2006-01-02 15:04:05",
	},
	"go": LogType{
		TimeReStr:  `^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006/01/02 15:04:05",
	},
	"docker": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)`,
		TimeLayout: "2006-01-02T15:04:05.000000000Z",
	},
	"docker_local": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"kubernetes": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)`,
		TimeLayout: "2006-01-02T15:04:05.000000000Z",
	},
	"syslog": LogType{
		TimeReStr:  `^(\w{3}\s+\d{1,2} \d{2}:\d{2}:\d{2})`,
		TimeLayout: "Jan _2 15:04:05",
	},
	"syslog_rfc5424": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"mysql": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)`,
		TimeLayout: "2006-01-02T15:04:05.000000Z",
	},
	"mysql_general": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02 15:04:05",
	},
	"postgresql": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+)`,
		TimeLayout: "2006-01-02 15:04:05.000",
	},
	"elasticsearch": LogType{
		TimeReStr:  `\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}),\d+\]`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"logstash": LogType{
		TimeReStr:  `"@timestamp":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)"`,
		TimeLayout: "2006-01-02T15:04:05.000Z",
	},
	"json": LogType{
		TimeReStr:  `"timestamp":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})"`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"json_time": LogType{
		TimeReStr:  `"time":"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})"`,
		TimeLayout: "2006-01-02T15:04:05",
	},
	"rails": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`,
		TimeLayout: "2006-01-02 15:04:05",
	},
	"django": LogType{
		TimeReStr:  `^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}),\d+`,
		TimeLayout: "2006-01-02 15:04:05",
	},
}

// LoadConfig loads configuration from file, falls back to builtin types
func LoadConfig(configFile string) (Config, error) {
	if configFile == "" {
		configFile = DefaultConfigFile
	}

	// Try to load from file first
	if _, err := os.Stat(configFile); err == nil {
		var conf Config
		if _, err := toml.DecodeFile(configFile, &conf); err != nil {
			return nil, err
		}

		// Merge with builtin types (file takes precedence)
		merged := make(Config)
		maps.Copy(merged, BuiltinLogTypes)
		maps.Copy(merged, conf)
		return merged, nil
	}

	// Fall back to builtin types if no config file
	return BuiltinLogTypes, nil
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
