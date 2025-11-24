package adapters_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestNoopLogger(t *testing.T) {
	logger := adapters.NewNoopLogger()
	ctx := context.Background()

	// Should not panic
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")

	withLogger := logger.With("key", "value")
	assert.NotNil(t, withLogger)
	withLogger.Info(ctx, "test")
}

func TestNoopTracer(t *testing.T) {
	tracer := adapters.NewNoopTracer()
	ctx := context.Background()

	newCtx, span := tracer.Start(ctx, "test.operation")
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	// Should not panic
	span.SetAttributes()
	span.RecordError(assert.AnError)
	span.End()
}

func TestNoopMetrics(t *testing.T) {
	metrics := adapters.NewNoopMetrics()

	counter := metrics.Counter("test_counter")
	assert.NotNil(t, counter)
	counter.Inc()
	counter.Add(5)

	histogram := metrics.Histogram("test_histogram")
	assert.NotNil(t, histogram)
	histogram.Observe(1.5)

	gauge := metrics.Gauge("test_gauge")
	assert.NotNil(t, gauge)
	gauge.Set(10)
	gauge.Inc()
	gauge.Dec()
}
