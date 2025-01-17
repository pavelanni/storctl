package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	globalLogger *slog.Logger
	once         sync.Once
)

// Initialize sets up the global logger with the specified level
func Initialize(level slog.Level) {
	once.Do(func() {
		opts := &slog.HandlerOptions{
			Level: level,
		}
		handler := slog.NewTextHandler(os.Stdout, opts)
		globalLogger = slog.New(handler)
	})
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if globalLogger == nil {
		// Default to INFO level if not initialized
		Initialize(slog.LevelInfo)
	}
	return globalLogger
}

func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to INFO if invalid level
	}
}
