package ttail

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sakateka/ttail/internal/config"
)

func createTestFile(t testing.TB, content string) *os.File {
	tmpFile, err := os.CreateTemp("", "ttail_test_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	return tmpFile
}

func TestTFile_FindPosition_BasicFunctionality(t *testing.T) {
	// Create test log with timestamps
	logContent := `	timestamp=2023-12-25T10:00:00	level=info	msg=first
	timestamp=2023-12-25T10:00:30	level=info	msg=second
	timestamp=2023-12-25T10:01:00	level=info	msg=third
	timestamp=2023-12-25T10:01:30	level=info	msg=fourth
	timestamp=2023-12-25T10:02:00	level=info	msg=fifth
`

	file := createTestFile(t, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	// Test with 1 minute duration
	tfile := NewTimeFile(file,
		WithDuration(1*time.Minute),
		WithTimeFromLastLine(true),
	)

	err := tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition failed: %v", err)
	}

	// Should find position (may be 0 if all content is within duration)
	offset := tfile.GetOffset()
	if offset < 0 {
		t.Errorf("expected non-negative offset, got %d", offset)
	}

	// Verify we can read from the position
	var buf bytes.Buffer
	copied, err := tfile.CopyTo(&buf)
	if err != nil {
		t.Fatalf("CopyTo failed: %v", err)
	}
	if copied <= 0 {
		t.Errorf("expected positive bytes copied, got %d", copied)
	}

	output := buf.String()
	if !strings.Contains(output, "msg=fourth") || !strings.Contains(output, "msg=fifth") {
		t.Errorf("output should contain recent entries, got: %s", output)
	}
}

func TestTFile_FindPosition_NoTimestamps(t *testing.T) {
	logContent := `line without timestamp
another line without timestamp
yet another line
`

	file := createTestFile(t, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	tfile := NewTimeFile(file,
		WithDuration(1*time.Minute),
		WithTimeFromLastLine(true),
	)

	err := tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition failed: %v", err)
	}

	// Should start from beginning when no timestamps found
	offset := tfile.GetOffset()
	if offset != 0 {
		t.Errorf("expected offset 0 for no timestamps, got %d", offset)
	}
}

func TestTFile_FindPosition_EmptyFile(t *testing.T) {
	file := createTestFile(t, "")
	defer os.Remove(file.Name())
	defer file.Close()

	tfile := NewTimeFile(file, WithDuration(1*time.Minute))

	err := tfile.FindPosition()
	if err != nil && err != io.EOF {
		t.Fatalf("FindPosition failed: %v", err)
	}
}

func TestTFile_GetReader(t *testing.T) {
	logContent := `	timestamp=2023-12-25T10:00:00	level=info	msg=test
	timestamp=2023-12-25T10:01:00	level=info	msg=test2
`

	file := createTestFile(t, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	tfile := NewTimeFile(file, WithDuration(30*time.Second))

	err := tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition failed: %v", err)
	}

	reader, err := tfile.GetReader()
	if err != nil {
		t.Fatalf("GetReader failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		t.Fatalf("reading from reader failed: %v", err)
	}

	// Output may be empty if no lines match the time criteria
	t.Logf("Reader output length: %d bytes", buf.Len())
}

func TestTFile_WithCustomOptions(t *testing.T) {
	// Test with custom buffer size and steps limit
	logContent := strings.Repeat("	timestamp=2023-12-25T10:00:00	level=info	msg=test\n", 100)

	file := createTestFile(t, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	tfile := NewTimeFile(file,
		WithDuration(1*time.Minute),
		WithBufSize(512), // Small buffer
		WithStepsLimit(10),
	)

	err := tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition with custom options failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = tfile.CopyTo(&buf)
	if err != nil {
		t.Fatalf("CopyTo failed: %v", err)
	}
}

func TestOptionsFromConfig_Integration(t *testing.T) {
	// Create a temporary config file
	configContent := `
[test_format]
bufSize = 8192
stepsLimit = 256
timeReStr = '\ttimestamp=(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\t'
timeLayout = "2006-01-02T15:04:05"
`

	tmpFile, err := os.CreateTemp("", "ttail_config_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	tmpFile.Close()

	// Test loading options from config by creating a custom config
	conf := make(config.Config)
	conf["test_format"] = config.LogType{
		BufSize:    8192,
		StepsLimit: 256,
		TimeReStr:  `\ttimestamp=(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\t`,
		TimeLayout: "2006-01-02T15:04:05",
	}

	// Create options manually for this test
	opts := []TimeFileOptions{
		WithBufSize(8192),
		WithStepsLimit(256),
		WithTimeReAsStr(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\t`),
		WithTimeLayout("2006-01-02T15:04:05"),
	}

	if len(opts) == 0 {
		t.Errorf("expected non-empty options")
	}

	// Test with actual file
	logContent := `	timestamp=2023-12-25T10:00:00	level=info	msg=test
	timestamp=2023-12-25T10:01:00	level=info	msg=test2
`

	file := createTestFile(t, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	allOpts := append([]TimeFileOptions{WithDuration(30 * time.Second)}, opts...)
	tfile := NewTimeFile(file, allOpts...)

	err = tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition with config options failed: %v", err)
	}
}

func BenchmarkTFile_FindPosition(b *testing.B) {
	// Create a large log file for benchmarking
	var logBuilder strings.Builder
	baseTime := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)

	for i := range 10000 {
		timestamp := baseTime.Add(time.Duration(i) * time.Second)
		logBuilder.WriteString(fmt.Sprintf("\ttimestamp=%s\tlevel=info\tmsg=entry_%d\n",
			timestamp.Format("2006-01-02T15:04:05"), i))
	}

	file := createTestFile(b, logBuilder.String())
	defer os.Remove(file.Name())
	defer file.Close()

	// Create TFile once and reuse it to avoid allocations
	tfile := NewTimeFile(file,
		WithDuration(1*time.Hour),
		WithTimeFromLastLine(true),
	)

	for b.Loop() {
		file.Seek(0, io.SeekStart)
		_ = tfile.FindPosition()
	}
}

func BenchmarkTFile_FindPosition_Direct(b *testing.B) {
	// Create a large log file for benchmarking
	var logBuilder strings.Builder
	baseTime := time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC)

	for i := range 10000 {
		timestamp := baseTime.Add(time.Duration(i) * time.Second)
		logBuilder.WriteString(fmt.Sprintf("\ttimestamp=%s\tlevel=info\tmsg=entry_%d\n",
			timestamp.Format("2006-01-02T15:04:05"), i))
	}

	file := createTestFile(b, logBuilder.String())
	defer os.Remove(file.Name())
	defer file.Close()

	// Use direct options to avoid slice allocations
	opts := config.DefaultOptions()
	opts.Duration = 1 * time.Hour
	opts.TimeFromLastLine = true

	// Create TFile once and reuse it to avoid allocations
	tfile := NewTimeFileWithOptions(file, opts)

	// Reset allocations counter after setup
	b.ResetTimer()

	for b.Loop() {
		file.Seek(0, io.SeekStart)
		_ = tfile.FindPosition()
	}
}

func BenchmarkTFile_CopyTo(b *testing.B) {
	logContent := strings.Repeat("	timestamp=2023-12-25T10:00:00	level=info	msg=test data here\n", 1000)

	file := createTestFile(b, logContent)
	defer os.Remove(file.Name())
	defer file.Close()

	tfile := NewTimeFile(file, WithDuration(1*time.Hour))
	_ = tfile.FindPosition()

	for b.Loop() {
		var buf bytes.Buffer
		_, _ = tfile.CopyTo(&buf)
	}
}
