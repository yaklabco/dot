package adapters

import (
	"context"

	"github.com/yaklabco/dot/internal/domain"
)

// NoopLogger is a logger that does nothing.
// Useful for testing and when logging is disabled.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (l *NoopLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (l *NoopLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (l *NoopLogger) Error(ctx context.Context, msg string, args ...any) {}

func (l *NoopLogger) With(args ...any) domain.Logger {
	return l
}

// NoopTracer is a tracer that does nothing.
// Useful for testing and when tracing is disabled.
type NoopTracer struct{}

// NewNoopTracer creates a new no-op tracer.
func NewNoopTracer() *NoopTracer {
	return &NoopTracer{}
}

func (t *NoopTracer) Start(ctx context.Context, name string, opts ...domain.SpanOption) (context.Context, domain.Span) {
	return ctx, &NoopSpan{}
}

// NoopSpan is a span that does nothing.
type NoopSpan struct{}

func (s *NoopSpan) End()                                    {}
func (s *NoopSpan) RecordError(err error)                   {}
func (s *NoopSpan) SetAttributes(attrs ...domain.Attribute) {}

// NoopMetrics is a metrics collector that does nothing.
// Useful for testing and when metrics are disabled.
type NoopMetrics struct{}

// NewNoopMetrics creates a new no-op metrics collector.
func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

func (m *NoopMetrics) Counter(name string, labels ...string) domain.Counter {
	return &NoopCounter{}
}

func (m *NoopMetrics) Histogram(name string, labels ...string) domain.Histogram {
	return &NoopHistogram{}
}

func (m *NoopMetrics) Gauge(name string, labels ...string) domain.Gauge {
	return &NoopGauge{}
}

// NoopCounter is a counter that does nothing.
type NoopCounter struct{}

func (c *NoopCounter) Inc(labels ...string)                {}
func (c *NoopCounter) Add(delta float64, labels ...string) {}

// NoopHistogram is a histogram that does nothing.
type NoopHistogram struct{}

func (h *NoopHistogram) Observe(value float64, labels ...string) {}

// NoopGauge is a gauge that does nothing.
type NoopGauge struct{}

func (g *NoopGauge) Set(value float64, labels ...string) {}
func (g *NoopGauge) Inc(labels ...string)                {}
func (g *NoopGauge) Dec(labels ...string)                {}
