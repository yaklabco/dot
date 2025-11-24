package dot

import (
	"io"
	"log/slog"

	"github.com/yaklabco/dot/internal/adapters"
)

// NewOSFilesystem returns a filesystem implementation that uses the OS filesystem.
func NewOSFilesystem() FS {
	return adapters.NewOSFilesystem()
}

// NewSlogLogger returns a logger backed by slog.
func NewSlogLogger(l *slog.Logger) Logger {
	return adapters.NewSlogLogger(l)
}

// NewNoopLogger returns a logger that discards all output.
func NewNoopLogger() Logger {
	return adapters.NewNoopLogger()
}

// NewJSONLogger returns a configured JSON logger.
func NewJSONLogger(w io.Writer, level slog.Level) Logger {
	return adapters.NewSlogLogger(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	})))
}

// NewTextLogger returns a configured text logger.
func NewTextLogger(w io.Writer, level slog.Level) Logger {
	return adapters.NewSlogLogger(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	})))
}
