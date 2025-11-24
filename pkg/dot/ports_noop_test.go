package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestNoopTracer(t *testing.T) {
	tracer := dot.NewNoopTracer()
	assert.NotNil(t, tracer)

	ctx := context.Background()
	newCtx, span := tracer.Start(ctx, "test")

	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)

	span.End()
	span.RecordError(assert.AnError)
	span.SetAttributes(dot.Attribute{Key: "test", Value: "value"})
}

func TestNoopMetrics(t *testing.T) {
	metrics := dot.NewNoopMetrics()
	assert.NotNil(t, metrics)

	counter := metrics.Counter("test")
	assert.NotNil(t, counter)
	counter.Inc()
	counter.Add(1.0)

	histogram := metrics.Histogram("test")
	assert.NotNil(t, histogram)
	histogram.Observe(1.0)

	gauge := metrics.Gauge("test")
	assert.NotNil(t, gauge)
	gauge.Set(1.0)
	gauge.Inc()
	gauge.Dec()
}
