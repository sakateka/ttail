package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Location != time.Local {
		t.Errorf("expected Local timezone, got %v", opts.Location)
	}
	if opts.BufSize != 1<<14 {
		t.Errorf("expected buffer size 16384, got %d", opts.BufSize)
	}
	if opts.StepsLimit != 1024 {
		t.Errorf("expected steps limit 1024, got %d", opts.StepsLimit)
	}
	if opts.TimeLayout != "2006-01-02T15:04:05" {
		t.Errorf("expected time layout '2006-01-02T15:04:05', got %s", opts.TimeLayout)
	}
}

func TestOptions_Clone(t *testing.T) {
	original := DefaultOptions()
	original.Duration = 30 * time.Second
	original.TimeFromLastLine = true

	cloned := original.Clone()

	// Modify original
	original.Duration = 60 * time.Second
	original.TimeFromLastLine = false

	// Cloned should be unchanged
	if cloned.Duration != 30*time.Second {
		t.Errorf("expected cloned duration 30s, got %v", cloned.Duration)
	}
	if !cloned.TimeFromLastLine {
		t.Errorf("expected cloned TimeFromLastLine true")
	}
}

func TestLogType_ApplyToOptions(t *testing.T) {
	opts := DefaultOptions()

	logType := &LogType{
		BufSize:    8192,
		StepsLimit: 512,
		TimeReStr:  `(\d{4}-\d{2}-\d{2})`,
		TimeLayout: "2006-01-02",
	}

	err := logType.ApplyToOptions(&opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.BufSize != 8192 {
		t.Errorf("expected buffer size 8192, got %d", opts.BufSize)
	}
	if opts.StepsLimit != 512 {
		t.Errorf("expected steps limit 512, got %d", opts.StepsLimit)
	}
	if opts.TimeLayout != "2006-01-02" {
		t.Errorf("expected time layout '2006-01-02', got %s", opts.TimeLayout)
	}

	// Test regex was compiled correctly
	testLine := "2023-12-25 some log data"
	if !opts.TimeRe.MatchString(testLine) {
		t.Errorf("compiled regex should match test line")
	}
}

func TestLogType_ApplyToOptions_InvalidRegex(t *testing.T) {
	opts := DefaultOptions()

	logType := &LogType{
		TimeReStr: `[invalid regex`,
	}

	err := logType.ApplyToOptions(&opts)
	if err == nil {
		t.Errorf("expected error for invalid regex")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	conf, err := LoadConfig("/non/existent/file.toml")
	if err != nil {
		t.Errorf("should not error for non-existent file, should return builtins: %v", err)
	}
	if len(conf) == 0 {
		t.Errorf("should return builtin types when config file doesn't exist")
	}
}

func TestConfig_GetLogTypeOptions(t *testing.T) {
	config := Config{
		"test": LogType{
			BufSize:    4096,
			StepsLimit: 256,
			TimeReStr:  `test-(\d+)`,
			TimeLayout: "test-layout",
		},
	}

	// Test existing log type
	logType, err := config.GetLogTypeOptions("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logType.BufSize != 4096 {
		t.Errorf("expected buffer size 4096, got %d", logType.BufSize)
	}

	// Test non-existent log type
	_, err = config.GetLogTypeOptions("nonexistent")
	if err == nil {
		t.Errorf("expected error for non-existent log type")
	}
}

func TestLoadConfig_WithTempFile(t *testing.T) {
	// Create a temporary config file
	content := `
[apache]
bufSize = 8192
stepsLimit = 512
timeReStr = '\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s'
timeLayout = "02/Jan/2006:15:04:05"

[nginx]
bufSize = 4096
timeReStr = '(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})'
timeLayout = "2006-01-02T15:04:05"
`

	tmpFile, err := os.CreateTemp("", "ttail_test_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading the config
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Test apache config
	apache, err := config.GetLogTypeOptions("apache")
	if err != nil {
		t.Fatalf("failed to get apache config: %v", err)
	}
	if apache.BufSize != 8192 {
		t.Errorf("expected apache buffer size 8192, got %d", apache.BufSize)
	}
	if apache.StepsLimit != 512 {
		t.Errorf("expected apache steps limit 512, got %d", apache.StepsLimit)
	}

	// Test nginx config
	nginx, err := config.GetLogTypeOptions("nginx")
	if err != nil {
		t.Fatalf("failed to get nginx config: %v", err)
	}
	if nginx.BufSize != 4096 {
		t.Errorf("expected nginx buffer size 4096, got %d", nginx.BufSize)
	}
	if nginx.StepsLimit != 0 { // not specified, should be 0
		t.Errorf("expected nginx steps limit 0, got %d", nginx.StepsLimit)
	}
}

func BenchmarkOptions_Clone(b *testing.B) {
	opts := DefaultOptions()
	opts.Duration = 30 * time.Second
	opts.TimeFromLastLine = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opts.Clone()
	}
}

func BenchmarkLogType_ApplyToOptions(b *testing.B) {
	logType := &LogType{
		BufSize:    8192,
		StepsLimit: 512,
		TimeReStr:  `(\d{4}-\d{2}-\d{2})`,
		TimeLayout: "2006-01-02",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := DefaultOptions()
		_ = logType.ApplyToOptions(&opts)
	}
}
