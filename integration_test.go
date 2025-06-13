package ttail

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestBasicIntegration tests the core functionality end-to-end
func TestBasicIntegration(t *testing.T) {
	// Create a test log file with timestamps
	logContent := generateTestLog(100)

	tmpFile, err := os.CreateTemp("", "ttail_integration_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(logContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	// Test basic functionality
	tfile := NewTimeFile(tmpFile,
		WithDuration(30*time.Second),
		WithTimeFromLastLine(true),
	)

	err = tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition failed: %v", err)
	}

	var buf bytes.Buffer
	copied, err := tfile.CopyTo(&buf)
	if err != nil {
		t.Fatalf("CopyTo failed: %v", err)
	}

	if copied <= 0 {
		t.Errorf("expected positive bytes copied, got %d", copied)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("expected non-empty output")
	}

	// Verify we got recent entries
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 10 {
		t.Errorf("expected at least 10 lines, got %d", len(lines))
	}

	t.Logf("Successfully processed %d bytes, %d lines", copied, len(lines))
}

// generateTestLog creates a test log with timestamps
func generateTestLog(numEntries int) string {
	var builder strings.Builder
	baseTime := time.Now().Add(-time.Duration(numEntries) * time.Second)

	for i := 0; i < numEntries; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Second)
		line := fmt.Sprintf("\ttimestamp=%s\tlevel=info\tmsg=entry_%d\n",
			timestamp.Format("2006-01-02T15:04:05"), i)
		builder.WriteString(line)
	}

	return builder.String()
}

// TestPerformanceBaseline provides a performance baseline
func TestPerformanceBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Create a larger test file
	logContent := generateTestLog(10000)

	tmpFile, err := os.CreateTemp("", "ttail_perf_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(logContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	start := time.Now()

	tfile := NewTimeFile(tmpFile,
		WithDuration(1*time.Hour),
		WithTimeFromLastLine(true),
	)

	err = tfile.FindPosition()
	if err != nil {
		t.Fatalf("FindPosition failed: %v", err)
	}

	var buf bytes.Buffer
	copied, err := tfile.CopyTo(&buf)
	if err != nil {
		t.Fatalf("CopyTo failed: %v", err)
	}

	elapsed := time.Since(start)

	t.Logf("Performance baseline: processed %d bytes in %v (%.2f MB/s)",
		copied, elapsed, float64(copied)/elapsed.Seconds()/1024/1024)

	// Basic performance check - should process at least 1MB/s
	throughput := float64(copied) / elapsed.Seconds()
	if throughput < 1024*1024 {
		t.Errorf("performance below baseline: %.2f bytes/s", throughput)
	}
}
