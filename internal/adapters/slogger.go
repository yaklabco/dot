package adapters

import (
	"context"
	"io"
	"log/slog"
	"strings"

	console "github.com/phsym/console-slog"

	"github.com/yaklabco/dot/internal/domain"
)

// SlogLogger implements the Logger interface using log/slog.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new slog logger adapter.
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{
		logger: logger,
	}
}

// NewConsoleLogger creates a logger with console-slog for human-readable output.
func NewConsoleLogger(w io.Writer, level string) *SlogLogger {
	logLevel := ParseLogLevel(level)

	handler := console.NewHandler(w, &console.HandlerOptions{
		Level: logLevel,
	})

	return &SlogLogger{
		logger: slog.New(handler),
	}
}

// Debug logs a debug-level message.
func (l *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// Info logs an info-level message.
func (l *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning-level message.
func (l *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

// Error logs an error-level message.
func (l *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// With returns a new logger with additional context fields.
func (l *SlogLogger) With(args ...any) domain.Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}

// ParseLogLevel converts a string log level to slog.Level.
func ParseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
