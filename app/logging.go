package app

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// InitLogger installs a slog handler as the process-wide default and routes the
// standard library's log package through it, so the existing log.Printf call
// sites emit structured records without being rewritten.
//
//	LOG_LEVEL   debug|info|warn|error   (default: info)
//	LOG_FORMAT  text|json               (default: json when APP_ENV=prod, else text)
//
// Logs go to w; nothing is written to disk. Rotation and retention are the
// supervisor's job (systemd/journald, Docker, launchd).
func InitLogger(w io.Writer) {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: shortenSource,
	}

	var handler slog.Handler
	if resolveLogFormat() == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	// slog.SetDefault reads log.Flags() to decide whether to capture the caller
	// PC for records arriving through the log package, then resets the flags to
	// zero. Setting Lshortfile first is what preserves file:line on the ~130
	// call sites that still use log.Printf.
	log.SetFlags(log.Lshortfile)

	slog.SetDefault(slog.New(handler))

	// Records from the log package carry no level of their own. Emit them at
	// the configured level so raising LOG_LEVEL never silences them outright.
	slog.SetLogLoggerLevel(level)
}

// shortenSource strips the build machine's directory prefix off the source
// attribute, leaving file:line as the old log.Lshortfile flag rendered it.
func shortenSource(groups []string, a slog.Attr) slog.Attr {
	if a.Key != slog.SourceKey {
		return a
	}

	if source, ok := a.Value.Any().(*slog.Source); ok {
		source.File = filepath.Base(source.File)
	}

	return a
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func resolveLogFormat() string {
	if format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT"))); format != "" {
		return format
	}

	if os.Getenv("APP_ENV") == "prod" {
		return "json"
	}

	return "text"
}
