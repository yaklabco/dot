package adapters_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestSlogLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := adapters.NewSlogLogger(slog.New(handler))

	ctx := context.Background()
	logger.Debug(ctx, "debug message", "key", "value")

	assert.Contains(t, buf.String(), "debug message")
	assert.Contains(t, buf.String(), "key")
	assert.Contains(t, buf.String(), "value")
}

func TestSlogLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := adapters.NewSlogLogger(slog.New(handler))

	ctx := context.Background()
	logger.Info(ctx, "info message", "key", "value")

	assert.Contains(t, buf.String(), "info message")
}

func TestSlogLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := adapters.NewSlogLogger(slog.New(handler))

	ctx := context.Background()
	logger.Warn(ctx, "warn message", "key", "value")

	assert.Contains(t, buf.String(), "warn message")
}

func TestSlogLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := adapters.NewSlogLogger(slog.New(handler))

	ctx := context.Background()
	logger.Error(ctx, "error message", "key", "value")

	assert.Contains(t, buf.String(), "error message")
}

func TestSlogLogger_With(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	baseLogger := adapters.NewSlogLogger(slog.New(handler))

	// Create logger with additional fields
	logger := baseLogger.With("component", "test")

	ctx := context.Background()
	logger.Info(ctx, "test message")

	assert.Contains(t, buf.String(), "component")
	assert.Contains(t, buf.String(), "test")
}

func TestNewConsoleLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := adapters.NewConsoleLogger(&buf, "DEBUG")
	require.NotNil(t, logger)

	ctx := context.Background()
	logger.Info(ctx, "test message", "key", "value")

	// Console logger output should contain the message
	output := buf.String()
	assert.NotEmpty(t, output)
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"invalid", slog.LevelInfo}, // Defaults to INFO
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := adapters.ParseLogLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}
