package dot

import "github.com/yaklabco/dot/internal/domain"

// Port interfaces re-exported from internal/domain.
//
// These types use type aliases (=) rather than type definitions for several reasons:
//
// 1. Interface contract stability: These are port interfaces defining contracts
//    between the domain and external adapters. The contract IS the interface
//    definition - whether defined here or in internal/domain makes no difference
//    for API stability.
//
// 2. Zero-overhead interoperability: Adapters (OSFilesystem, MemFS, etc.) implement
//    domain.* interfaces and are returned through the public API. Type aliases
//    allow these to satisfy dot.* interfaces without wrapper types.
//
// 3. Go idiom: Type aliases for re-exporting types from internal packages is a
//    standard Go pattern in hexagonal/ports-and-adapters architectures.
//
// 4. No real encapsulation benefit: Defining identical interfaces twice would
//    not provide additional protection - interface changes are breaking changes
//    regardless of the definition location.
//
// If internal interface changes are needed, they represent intentional API
// evolution and should be reflected in the public types accordingly.

// FSReader provides read-only filesystem operations.
type FSReader = domain.FSReader

// FSWriter provides write filesystem operations.
type FSWriter = domain.FSWriter

// FS combines all filesystem operations (read and write).
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
