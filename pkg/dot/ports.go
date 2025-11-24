package dot

import "github.com/yaklabco/dot/internal/domain"

// Port interfaces re-exported from internal/domain

// FS defines the filesystem abstraction interface.
type FS = domain.FS

// FileInfo provides information about a file.
type FileInfo = domain.FileInfo

// DirEntry provides information about a directory entry.
type DirEntry = domain.DirEntry

// Logger provides structured logging.
type Logger = domain.Logger

// Tracer provides distributed tracing support.
type Tracer = domain.Tracer

// Span represents a trace span.
type Span = domain.Span

// SpanOption configures span creation.
type SpanOption = domain.SpanOption

// Attribute represents a span attribute.
type Attribute = domain.Attribute

// Metrics provides metrics collection.
type Metrics = domain.Metrics

// Counter represents a monotonically increasing counter.
type Counter = domain.Counter

// Histogram represents a distribution of values.
type Histogram = domain.Histogram

// Gauge represents an instantaneous value.
type Gauge = domain.Gauge

// NewNoopTracer returns a tracer that does nothing.
func NewNoopTracer() Tracer {
	return domain.NewNoopTracer()
}

// NewNoopMetrics returns a metrics collector that does nothing.
func NewNoopMetrics() Metrics {
	return domain.NewNoopMetrics()
}
