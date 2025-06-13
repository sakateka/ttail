package parser

import (
	"regexp"
	"testing"
	"time"
)

func TestTimeParser_ParseTime(t *testing.T) {
	tests := []struct {
		name       string
		regex      string
		layout     string
		line       string
		expectTime bool
		expectErr  bool
	}{
		{
			name:       "valid timestamp",
			regex:      `\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`,
			layout:     "2006-01-02T15:04:05",
			line:       "\ttimestamp=2023-12-25T10:30:45\tother data",
			expectTime: true,
			expectErr:  false,
		},
		{
			name:       "no timestamp",
			regex:      `\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`,
			layout:     "2006-01-02T15:04:05",
			line:       "no timestamp here",
			expectTime: false,
			expectErr:  false,
		},
		{
			name:       "invalid timestamp format",
			regex:      `\ttimestamp=([^\t]+)\t`,
			layout:     "2006-01-02T15:04:05",
			line:       "\ttimestamp=invalid-timestamp\tother data",
			expectTime: false,
			expectErr:  true,
		},
		{
			name:       "apache log format",
			regex:      `\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2})\s`,
			layout:     "02/Jan/2006:15:04:05",
			line:       "[25/Dec/2023:10:30:45 +0000] GET /test HTTP/1.1",
			expectTime: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.regex)
			parser := NewTimeParser(re, tt.layout, time.UTC)

			result, ok := parser.ParseTime([]byte(tt.line))

			if tt.expectErr && ok {
				t.Errorf("expected error but got success")
			}
			if !tt.expectErr && !ok && tt.expectTime {
				t.Errorf("unexpected parsing failure")
			}
			if tt.expectTime && !ok {
				t.Errorf("expected timestamp but parsing failed")
			}
			if !tt.expectTime && ok && !tt.expectErr {
				t.Errorf("expected no timestamp but got: %v", result)
			}
		})
	}
}

func TestTimeParser_GetMethods(t *testing.T) {
	re := regexp.MustCompile(`test`)
	layout := "2006-01-02"
	loc := time.UTC

	parser := NewTimeParser(re, layout, loc)

	if parser.GetRegex() != re {
		t.Errorf("GetRegex() returned wrong regex")
	}
	if parser.GetLayout() != layout {
		t.Errorf("GetLayout() returned wrong layout")
	}
	if parser.GetLocation() != loc {
		t.Errorf("GetLocation() returned wrong location")
	}
}

func BenchmarkTimeParser_ParseTime(b *testing.B) {
	re := regexp.MustCompile(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`)
	parser := NewTimeParser(re, "2006-01-02T15:04:05", time.UTC)
	line := []byte("\ttimestamp=2023-12-25T10:30:45\tother data here")

	for b.Loop() {
		_, _ = parser.ParseTime(line)
	}
}
