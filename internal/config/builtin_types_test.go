package config

import (
	"regexp"
	"testing"
	"time"
)

func TestBuiltinLogTypes(t *testing.T) {
	tests := []struct {
		name       string
		logType    string
		sampleLine string
		expectTime string
	}{
		{
			name:       "TSKV format",
			logType:    "tskv",
			sampleLine: "\ttimestamp=2023-12-25T10:30:45\tlevel=info\tmsg=test",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Kernel logs",
			logType:    "kern",
			sampleLine: "2023-12-25T10:30:45.123456+03:00 hostname kernel: message",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Apache Common Log",
			logType:    "apache",
			sampleLine: `127.0.0.1 - - [25/Dec/2023:10:30:45 +0000] "GET / HTTP/1.1" 200 1234`,
			expectTime: "25/Dec/2023:10:30:45",
		},
		{
			name:       "Apache Combined Log",
			logType:    "apache_combined",
			sampleLine: `127.0.0.1 - - [25/Dec/2023:10:30:45 +0000] "GET / HTTP/1.1" 200 1234 "-" "Mozilla/5.0"`,
			expectTime: "25/Dec/2023:10:30:45",
		},
		{
			name:       "Nginx default",
			logType:    "nginx",
			sampleLine: `127.0.0.1 - - [25/Dec/2023:10:30:45 +0000] "GET / HTTP/1.1" 200 1234`,
			expectTime: "25/Dec/2023:10:30:45",
		},
		{
			name:       "Nginx ISO format",
			logType:    "nginx_iso",
			sampleLine: "2023-12-25T10:30:45 [error] 1234#0: message",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Java application logs",
			logType:    "java",
			sampleLine: "2023-12-25 10:30:45 INFO [main] com.example.App - Starting application",
			expectTime: "2023-12-25 10:30:45",
		},
		{
			name:       "Java ISO format",
			logType:    "java_iso",
			sampleLine: "2023-12-25T10:30:45 INFO [main] com.example.App - Starting application",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Python logging",
			logType:    "python",
			sampleLine: "2023-12-25 10:30:45,123 INFO root: Starting application",
			expectTime: "2023-12-25 10:30:45",
		},
		{
			name:       "Go standard log",
			logType:    "go",
			sampleLine: "2023/12/25 10:30:45 Starting application",
			expectTime: "2023/12/25 10:30:45",
		},
		{
			name:       "Docker container logs",
			logType:    "docker",
			sampleLine: "2023-12-25T10:30:45.123456789Z Starting application",
			expectTime: "2023-12-25T10:30:45.123456789Z",
		},
		{
			name:       "Docker local timezone",
			logType:    "docker_local",
			sampleLine: "2023-12-25T10:30:45 Starting application",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Kubernetes pod logs",
			logType:    "kubernetes",
			sampleLine: "2023-12-25T10:30:45.123456789Z Starting application",
			expectTime: "2023-12-25T10:30:45.123456789Z",
		},
		{
			name:       "Traditional syslog",
			logType:    "syslog",
			sampleLine: "Dec 25 10:30:45 hostname daemon: message",
			expectTime: "Dec 25 10:30:45",
		},
		{
			name:       "Modern syslog RFC5424",
			logType:    "syslog_rfc5424",
			sampleLine: "2023-12-25T10:30:45 hostname daemon: message",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "MySQL error log",
			logType:    "mysql",
			sampleLine: "2023-12-25T10:30:45.123456Z 0 [Note] Starting MySQL",
			expectTime: "2023-12-25T10:30:45.123456Z",
		},
		{
			name:       "MySQL general log",
			logType:    "mysql_general",
			sampleLine: "2023-12-25 10:30:45 1 Connect user@localhost",
			expectTime: "2023-12-25 10:30:45",
		},
		{
			name:       "PostgreSQL log",
			logType:    "postgresql",
			sampleLine: "2023-12-25 10:30:45.123 UTC [1234] LOG: starting PostgreSQL",
			expectTime: "2023-12-25 10:30:45.123",
		},
		{
			name:       "Elasticsearch log",
			logType:    "elasticsearch",
			sampleLine: "[2023-12-25T10:30:45,123][INFO ][o.e.n.Node] starting Elasticsearch",
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Logstash JSON",
			logType:    "logstash",
			sampleLine: `{"@timestamp":"2023-12-25T10:30:45.123Z","message":"test"}`,
			expectTime: "2023-12-25T10:30:45.123Z",
		},
		{
			name:       "JSON with timestamp",
			logType:    "json",
			sampleLine: `{"timestamp":"2023-12-25T10:30:45","level":"info","message":"test"}`,
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "JSON with time field",
			logType:    "json_time",
			sampleLine: `{"time":"2023-12-25T10:30:45","level":"info","message":"test"}`,
			expectTime: "2023-12-25T10:30:45",
		},
		{
			name:       "Rails application log",
			logType:    "rails",
			sampleLine: "2023-12-25 10:30:45 INFO Started GET \"/\" for 127.0.0.1",
			expectTime: "2023-12-25 10:30:45",
		},
		{
			name:       "Django application log",
			logType:    "django",
			sampleLine: "2023-12-25 10:30:45,123 INFO django.request: GET /",
			expectTime: "2023-12-25 10:30:45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the log type configuration
			logType, exists := BuiltinLogTypes[tt.logType]
			if !exists {
				t.Fatalf("log type %s not found in builtin types", tt.logType)
			}

			// Test regex compilation
			re, err := regexp.Compile(logType.TimeReStr)
			if err != nil {
				t.Fatalf("failed to compile regex for %s: %v", tt.logType, err)
			}

			// Test regex matching
			matches := re.FindStringSubmatch(tt.sampleLine)
			if len(matches) < 2 {
				t.Fatalf("regex for %s did not match sample line: %q", tt.logType, tt.sampleLine)
			}

			extractedTime := matches[1]
			if extractedTime != tt.expectTime {
				t.Errorf("expected extracted time %q, got %q", tt.expectTime, extractedTime)
			}

			// Test time parsing (for most formats)
			if tt.logType != "syslog" { // syslog needs current year context
				_, err = time.Parse(logType.TimeLayout, extractedTime)
				if err != nil {
					t.Errorf("failed to parse extracted time %q with layout %q: %v",
						extractedTime, logType.TimeLayout, err)
				}
			}
		})
	}
}

func TestBuiltinLogTypesCount(t *testing.T) {
	expectedCount := 25 // Update this when adding new types
	if len(BuiltinLogTypes) != expectedCount {
		t.Errorf("expected %d builtin log types, got %d", expectedCount, len(BuiltinLogTypes))
	}
}

func TestLoadConfigWithBuiltins(t *testing.T) {
	// Test loading with non-existent config file (should return builtins)
	conf, err := LoadConfig("/non/existent/file.toml")
	if err != nil {
		t.Fatalf("LoadConfig should not fail when falling back to builtins: %v", err)
	}

	// Should contain all builtin types
	if len(conf) != len(BuiltinLogTypes) {
		t.Errorf("expected %d types from builtins, got %d", len(BuiltinLogTypes), len(conf))
	}

	// Test a few key types
	testTypes := []string{"tskv", "kern", "apache", "java", "docker"}
	for _, logType := range testTypes {
		if _, exists := conf[logType]; !exists {
			t.Errorf("builtin log type %s not found in loaded config", logType)
		}
	}
}

func TestGetLogTypeOptionsWithBuiltins(t *testing.T) {
	conf := BuiltinLogTypes

	// Test getting existing type
	logType, err := conf.GetLogTypeOptions("java")
	if err != nil {
		t.Fatalf("failed to get java log type: %v", err)
	}

	if logType.TimeReStr == "" {
		t.Errorf("java log type should have TimeReStr")
	}
	if logType.TimeLayout == "" {
		t.Errorf("java log type should have TimeLayout")
	}

	// Test getting non-existent type
	_, err = conf.GetLogTypeOptions("nonexistent")
	if err == nil {
		t.Errorf("should fail for non-existent log type")
	}
}

func BenchmarkBuiltinLogTypeRegex(b *testing.B) {
	logType := BuiltinLogTypes["java"]
	re, _ := regexp.Compile(logType.TimeReStr)
	sampleLine := "2023-12-25 10:30:45 INFO [main] com.example.App - Starting application"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.FindStringSubmatch(sampleLine)
	}
}
