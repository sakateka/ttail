package buffer

import (
	"io"
	"strings"
	"testing"
)

type mockReaderAt struct {
	data []byte
}

func (m *mockReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, io.EOF
	}

	n = copy(p, m.data[off:])
	if off+int64(n) >= int64(len(m.data)) {
		err = io.EOF
	}
	return n, err
}

func TestLineBuffer_ReadLine(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		bufSize  int64
		offset   int64
		expected string
	}{
		{
			name:     "simple line",
			data:     "first line\nsecond line\nthird line\n",
			bufSize:  1024,
			offset:   0,
			expected: "first line",
		},
		{
			name:     "line from middle",
			data:     "first line\nsecond line\nthird line\n",
			bufSize:  1024,
			offset:   10, // start at newline before "second line"
			expected: "second line",
		},
		{
			name:     "small buffer",
			data:     "very long line that exceeds buffer size\nshort\n",
			bufSize:  50, // Make buffer large enough
			offset:   0,
			expected: "very long line that exceeds buffer size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockReaderAt{data: []byte(tt.data)}
			buffer := NewLineBuffer(tt.bufSize)

			line, err := buffer.ReadLine(reader, tt.offset)
			if err != nil && err != io.EOF {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(line) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(line))
			}
		})
	}
}

func TestLineBuffer_NextLine(t *testing.T) {
	data := "first line\nsecond line\nthird line\n"
	reader := &mockReaderAt{data: []byte(data)}
	buffer := NewLineBuffer(1024)

	// Read first line
	line1, err := buffer.ReadLine(reader, 0)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error reading first line: %v", err)
	}
	if string(line1) != "first line" {
		t.Errorf("expected 'first line', got %q", string(line1))
	}

	// Get next line from buffer
	line2, err := buffer.NextLine()
	if err != nil {
		t.Fatalf("unexpected error getting next line: %v", err)
	}
	if string(line2) != "second line" {
		t.Errorf("expected 'second line', got %q", string(line2))
	}

	// Get third line
	line3, err := buffer.NextLine()
	if err != nil {
		t.Fatalf("unexpected error getting third line: %v", err)
	}
	if string(line3) != "third line" {
		t.Errorf("expected 'third line', got %q", string(line3))
	}

	// Should get EOF for next line
	_, err = buffer.NextLine()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestLineBuffer_FindLastLine(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		offset   int64
		expected string
	}{
		{
			name:     "find last complete line",
			data:     "first line\nsecond line\nthird line\npartial",
			offset:   0,
			expected: "third line",
		},
		{
			name:     "single line",
			data:     "only line\n",
			offset:   0,
			expected: "only line",
		},
		{
			name:     "no complete lines",
			data:     "no newlines here",
			offset:   0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockReaderAt{data: []byte(tt.data)}
			buffer := NewLineBuffer(1024)

			line, err := buffer.FindLastLine(reader, tt.offset)
			if tt.expected == "" {
				if err != io.EOF {
					t.Errorf("expected EOF for no lines case")
				}
				return
			}

			if err != nil && err != io.EOF {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(line) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(line))
			}
		})
	}
}

func TestLineBuffer_Reset(t *testing.T) {
	buffer := NewLineBuffer(1024)

	// Set some state
	buffer.SetLineStart(10)
	buffer.SetLineEnd(20)
	buffer.SetDiscard(false)

	// Reset
	buffer.Reset()

	if buffer.GetLineStart() != -1 {
		t.Errorf("expected lineStart -1, got %d", buffer.GetLineStart())
	}
	if buffer.GetLineEnd() != 0 {
		t.Errorf("expected lineEnd 0, got %d", buffer.GetLineEnd())
	}
	if !buffer.IsDiscarded() {
		t.Errorf("expected discard true")
	}
}

func BenchmarkLineBuffer_ReadLine(b *testing.B) {
	data := strings.Repeat("this is a test line with some data\n", 1000)
	reader := &mockReaderAt{data: []byte(data)}
	buffer := NewLineBuffer(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Reset()
		_, _ = buffer.ReadLine(reader, 0)
	}
}

func BenchmarkLineBuffer_FindLastLine(b *testing.B) {
	data := strings.Repeat("this is a test line with some data\n", 1000)
	reader := &mockReaderAt{data: []byte(data)}
	buffer := NewLineBuffer(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = buffer.FindLastLine(reader, 0)
	}
}
