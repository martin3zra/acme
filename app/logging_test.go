package app

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"testing"
)

// restoreLoggerGlobals captures the process-wide logger state InitLogger
// mutates, so one test's configuration cannot leak into the next.
func restoreLoggerGlobals(t *testing.T) {
	t.Helper()

	flags := log.Flags()
	output := log.Writer()
	logger := slog.Default()

	t.Cleanup(func() {
		slog.SetDefault(logger)
		slog.SetLogLoggerLevel(slog.LevelInfo)
		log.SetOutput(output)
		log.SetFlags(flags)
	})
}

func TestInitLoggerBridgesStdlibLog(t *testing.T) {
	restoreLoggerGlobals(t)
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "text")

	var buf bytes.Buffer
	InitLogger(&buf)

	log.Printf("scheduler error: %v", "boom")

	got := buf.String()
	for _, want := range []string{"level=INFO", `msg="scheduler error: boom"`, "logging_test.go:"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
}

// The ~130 surviving log.Printf call sites carry no level of their own. Raising
// LOG_LEVEL must not silence them, or a prod deploy at warn/error goes dark.
func TestInitLoggerKeepsStdlibLogVisibleAtHigherLevels(t *testing.T) {
	restoreLoggerGlobals(t)
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("LOG_FORMAT", "text")

	var buf bytes.Buffer
	InitLogger(&buf)

	log.Println("still here")

	got := buf.String()
	if !strings.Contains(got, "still here") {
		t.Fatalf("stdlib log was silenced at LOG_LEVEL=error: %q", got)
	}
	if !strings.Contains(got, "level=ERROR") {
		t.Errorf("expected stdlib records to adopt the configured level, got %q", got)
	}
}

func TestInitLoggerFiltersBelowConfiguredLevel(t *testing.T) {
	restoreLoggerGlobals(t)
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_FORMAT", "text")

	var buf bytes.Buffer
	InitLogger(&buf)

	slog.Debug("chatty")
	slog.Warn("listen")

	got := buf.String()
	if strings.Contains(got, "chatty") {
		t.Errorf("debug record survived LOG_LEVEL=warn: %q", got)
	}
	if !strings.Contains(got, "listen") {
		t.Errorf("warn record was dropped: %q", got)
	}
}

func TestInitLoggerJSONFormat(t *testing.T) {
	restoreLoggerGlobals(t)
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FORMAT", "json")

	var buf bytes.Buffer
	InitLogger(&buf)

	slog.Info("structured", "import_id", "abc123")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, buf.String())
	}
	if record["msg"] != "structured" {
		t.Errorf("msg = %v, want %q", record["msg"], "structured")
	}
	if record["import_id"] != "abc123" {
		t.Errorf("import_id = %v, want %q", record["import_id"], "abc123")
	}
}

func TestResolveLogFormatDefaultsToJSONInProd(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		format string
		want   string
	}{
		{"prod defaults to json", "prod", "", "json"},
		{"dev defaults to text", "dev", "", "text"},
		{"unset defaults to text", "", "", "text"},
		{"explicit format overrides prod", "prod", "text", "text"},
		{"explicit format is normalized", "dev", " JSON ", "json"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("APP_ENV", test.env)
			t.Setenv("LOG_FORMAT", test.format)

			if got := resolveLogFormat(); got != test.want {
				t.Errorf("resolveLogFormat() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{" warn ", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"", slog.LevelInfo},
		{"nonsense", slog.LevelInfo},
	}

	for _, test := range tests {
		if got := parseLogLevel(test.in); got != test.want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", test.in, got, test.want)
		}
	}
}
