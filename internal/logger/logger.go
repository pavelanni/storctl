package logger

import (
	"log/slog"
	"os"
)

var (
	defaultLogger *slog.Logger
)

func init() {
	// Create a text handler with custom options
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// Replace the default time format with a more concise one
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // Remove timestamp for CLI output
			}
			return a
		},
	})

	defaultLogger = slog.New(handler)
}

// GetLogger returns the default logger instance
func GetLogger() *slog.Logger {
	return defaultLogger
}

// SetLogger allows replacing the default logger
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}

// Convenience methods
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}
