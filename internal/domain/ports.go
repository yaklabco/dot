package domain

import (
	"context"
	"io/fs"
	"os"
)

// FSReader provides read-only filesystem operations.
type FSReader interface {
	Stat(ctx context.Context, path string) (FileInfo, error)
	Lstat(ctx context.Context, path string) (FileInfo, error)
	ReadDir(ctx context.Context, path string) ([]DirEntry, error)
	ReadLink(ctx context.Context, path string) (string, error)
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// Queries (read-only checks)
	Exists(ctx context.Context, path string) bool
	IsDir(ctx context.Context, path string) (bool, error)
	IsSymlink(ctx context.Context, path string) (bool, error)
}

// FSWriter provides write filesystem operations.
type FSWriter interface {
	WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error
	Mkdir(ctx context.Context, path string, perm os.FileMode) error
	MkdirAll(ctx context.Context, path string, perm os.FileMode) error
	Remove(ctx context.Context, path string) error
	RemoveAll(ctx context.Context, path string) error
	Symlink(ctx context.Context, oldname, newname string) error
	Rename(ctx context.Context, oldpath, newpath string) error
}

// FS combines all filesystem operations.
// This is the full interface for components that need both read and write access.
type FS interface {
	FSReader
	FSWriter
}

// FileInfo is a type alias for the standard library fs.FileInfo interface.
// Using the stdlib type directly simplifies interoperability and eliminates
// the need for wrapper types when interfacing with standard library functions.
type FileInfo = fs.FileInfo

// DirEntry is a type alias for the standard library fs.DirEntry interface.
// Using the stdlib type directly simplifies interoperability and eliminates
// the need for wrapper types when interfacing with standard library functions.
type DirEntry = fs.DirEntry

// Logger defines the logging abstraction interface.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
	With(fields ...any) Logger
}

// Tracer defines the distributed tracing abstraction interface.
type Tracer interface {
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a single span in a trace.
type Span interface {
	End()
	RecordError(err error)
	SetAttributes(attrs ...Attribute)
}

// SpanOption configures span creation.
type SpanOption func(*SpanConfig)

// SpanConfig holds span configuration.
type SpanConfig struct {
	Attributes []Attribute
}

// Attribute represents a key-value span attribute.
type Attribute struct {
	Key   string
	Value any
}

// Metrics defines the metrics collection abstraction interface.
type Metrics interface {
	Counter(name string, labels ...string) Counter
	Histogram(name string, labels ...string) Histogram
	Gauge(name string, labels ...string) Gauge
}

// Counter represents a monotonically increasing metric.
type Counter interface {
	Inc(labels ...string)
	Add(value float64, labels ...string)
}

// Histogram represents a distribution of values.
type Histogram interface {
	Observe(value float64, labels ...string)
}

// Gauge represents a value that can go up or down.
type Gauge interface {
	Set(value float64, labels ...string)
	Inc(labels ...string)
	Dec(labels ...string)
}

// NewNoopTracer returns a tracer that does nothing.
func NewNoopTracer() Tracer {
	return &noopTracer{}
}

type noopTracer struct{}

func (n *noopTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

type noopSpan struct{}

func (n *noopSpan) End()                             {}
func (n *noopSpan) RecordError(err error)            {}
func (n *noopSpan) SetAttributes(attrs ...Attribute) {}

// NewNoopMetrics returns a metrics implementation that does nothing.
func NewNoopMetrics() Metrics {
	return &noopMetrics{}
}

type noopMetrics struct{}

func (n *noopMetrics) Counter(name string, labels ...string) Counter {
	return &noopCounter{}
}

func (n *noopMetrics) Histogram(name string, labels ...string) Histogram {
	return &noopHistogram{}
}

func (n *noopMetrics) Gauge(name string, labels ...string) Gauge {
	return &noopGauge{}
}

type noopCounter struct{}

func (n *noopCounter) Inc(labels ...string)                {}
func (n *noopCounter) Add(value float64, labels ...string) {}

type noopHistogram struct{}

func (n *noopHistogram) Observe(value float64, labels ...string) {}

type noopGauge struct{}

func (n *noopGauge) Set(value float64, labels ...string) {}
func (n *noopGauge) Inc(labels ...string)                {}
func (n *noopGauge) Dec(labels ...string)                {}
