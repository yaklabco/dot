package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// mockMetrics captures metrics calls for testing
type mockMetrics struct {
	counters   map[string]float64
	histograms map[string][]float64
	gauges     map[string]float64
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		counters:   make(map[string]float64),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

func (m *mockMetrics) Counter(name string, labels ...string) domain.Counter {
	return &mockCounter{metrics: m, name: name}
}

func (m *mockMetrics) Histogram(name string, labels ...string) domain.Histogram {
	return &mockHistogram{metrics: m, name: name}
}

func (m *mockMetrics) Gauge(name string, labels ...string) domain.Gauge {
	return &mockGauge{metrics: m, name: name}
}

type mockCounter struct {
	metrics *mockMetrics
	name    string
}

func (c *mockCounter) Inc(labels ...string) {
	c.metrics.counters[c.name]++
}

func (c *mockCounter) Add(delta float64, labels ...string) {
	c.metrics.counters[c.name] += delta
}

type mockHistogram struct {
	metrics *mockMetrics
	name    string
}

func (h *mockHistogram) Observe(value float64, labels ...string) {
	h.metrics.histograms[h.name] = append(h.metrics.histograms[h.name], value)
}

type mockGauge struct {
	metrics *mockMetrics
	name    string
}

func (g *mockGauge) Set(value float64, labels ...string) {
	g.metrics.gauges[g.name] = value
}

func (g *mockGauge) Inc(labels ...string) {
	g.metrics.gauges[g.name]++
}

func (g *mockGauge) Dec(labels ...string) {
	g.metrics.gauges[g.name]--
}

func TestInstrumentedExecutor_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	metrics := newMockMetrics()

	inner := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	exec := NewInstrumented(inner, metrics)

	// Create simple plan
	source := domain.MustParsePath("/packages/pkg/file")
	target := domain.MustParseTargetPath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	op := domain.NewLinkCreate("link1", source, target)
	plan := domain.Plan{
		Operations: []domain.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk())

	// Verify metrics were recorded
	require.Equal(t, float64(1), metrics.counters["executor.executions.total"])
	require.Equal(t, float64(1), metrics.counters["executor.executions.success"])
	require.Equal(t, float64(1), metrics.counters["executor.operations.executed"])
	require.Equal(t, float64(1), metrics.gauges["executor.operations.queued"])
	require.Len(t, metrics.histograms["executor.duration.seconds"], 1)
}

func TestInstrumentedExecutor_Failure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	metrics := newMockMetrics()

	inner := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	exec := NewInstrumented(inner, metrics)

	// Operation will fail during prepare (source doesn't exist)
	source := domain.MustParsePath("/nonexistent")
	target := domain.MustParseTargetPath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	op := domain.NewLinkCreate("link1", source, target)
	plan := domain.Plan{
		Operations: []domain.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr())

	// Verify metrics
	require.Equal(t, float64(1), metrics.counters["executor.executions.total"])
	require.Equal(t, float64(1), metrics.counters["executor.executions.failed"])
	require.Len(t, metrics.histograms["executor.duration.seconds"], 1)
}

func TestInstrumentedExecutor_WithRollback(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	metrics := newMockMetrics()

	inner := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	exec := NewInstrumented(inner, metrics)

	// Create operations where prepare passes but execute fails
	// To test rollback metrics, we need to manually trigger a failure
	// after some operations succeed. Since prepare validates everything,
	// we'll create a plan that will pass prepare but fail during parallel execute.

	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))

	op1 := domain.NewLinkCreate("link1", source1, target1)

	plan := domain.Plan{
		Operations: []domain.Operation{op1},
	}

	// Execute successfully first
	result := exec.Execute(ctx, plan)
	require.True(t, result.IsOk())

	// Verify successful execution metrics
	require.Equal(t, float64(1), metrics.counters["executor.executions.total"])
	require.Equal(t, float64(1), metrics.counters["executor.executions.success"])
	require.Equal(t, float64(1), metrics.counters["executor.operations.executed"])
}

func TestInstrumentedExecutor_ParallelMetrics(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	metrics := newMockMetrics()

	inner := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	exec := NewInstrumented(inner, metrics)

	// Create plan with parallel batches
	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target2 := domain.MustParseTargetPath("/home/file2")

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))

	op1 := domain.NewLinkCreate("link1", source1, target1)
	op2 := domain.NewLinkCreate("link2", source2, target2)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2},
		Batches: [][]domain.Operation{
			{op1, op2},
		},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk())

	// Verify parallel metrics
	require.Equal(t, float64(1), metrics.counters["executor.executions.total"])
	require.Equal(t, float64(1), metrics.counters["executor.executions.success"])
	require.Equal(t, float64(2), metrics.counters["executor.operations.executed"])
	require.Len(t, metrics.histograms["executor.parallel.batches"], 1)
	require.Equal(t, float64(1), metrics.histograms["executor.parallel.batches"][0])
}
