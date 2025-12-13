package executor

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestExecuteBatch_Concurrent(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create independent operations (no dependencies)
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target2 := domain.MustParseTargetPath("/home/file2")
	source3 := domain.MustParsePath("/packages/pkg/file3")
	target3 := domain.MustParseTargetPath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	ops := []domain.Operation{
		domain.NewLinkCreate("link1", source1, target1),
		domain.NewLinkCreate("link2", source2, target2),
		domain.NewLinkCreate("link3", source3, target3),
	}

	checkpoint := exec.checkpoint.Create(ctx)
	result := exec.executeBatch(ctx, ops, checkpoint)

	require.Len(t, result.Executed, 3)
	require.Empty(t, result.Failed)

	// Verify all links created
	require.True(t, fs.Exists(ctx, target1.String()))
	require.True(t, fs.Exists(ctx, target2.String()))
	require.True(t, fs.Exists(ctx, target3.String()))
}

func TestExecuteBatch_PartialFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	// Mix of success and failure
	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	source3 := domain.MustParsePath("/packages/pkg/file3")
	target3 := domain.MustParseTargetPath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	ops := []domain.Operation{
		domain.NewLinkCreate("link1", source1, target1),
		// This will fail because parent directory /nonexistent doesn't exist
		domain.NewLinkCreate("link2", domain.MustParsePath("/packages/pkg/file3"), domain.MustParseTargetPath("/nonexistent/file2")),
		domain.NewLinkCreate("link3", source3, target3),
	}

	checkpoint := exec.checkpoint.Create(ctx)
	result := exec.executeBatch(ctx, ops, checkpoint)

	require.Len(t, result.Executed, 2, "two operations should succeed")
	require.Len(t, result.Failed, 1, "one operation should fail")
	require.Contains(t, result.Failed, domain.OperationID("link2"))
}

func TestExecute_ParallelBatches(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create plan with parallelizable operations
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/dir1", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home/dir2", 0755))

	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/dir1/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target2 := domain.MustParseTargetPath("/home/dir2/file2")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("c1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("c2"), 0644))

	ops := []domain.Operation{
		domain.NewLinkCreate("link1", source1, target1),
		domain.NewLinkCreate("link2", source2, target2),
	}

	// Create plan with parallel batches
	plan := domain.Plan{
		Operations: ops,
		Batches: [][]domain.Operation{
			{ops[0], ops[1]}, // Both in same batch (can run in parallel)
		},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")
	execResult := result.Unwrap()
	require.Len(t, execResult.Executed, 2)
	require.Empty(t, execResult.Failed)
}

func TestExecuteParallel_Internal_MultipleBatches(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target2 := domain.MustParseTargetPath("/home/file2")
	source3 := domain.MustParsePath("/packages/pkg/file3")
	target3 := domain.MustParseTargetPath("/home/file3")

	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source2.String(), []byte("content2"), 0644))
	require.NoError(t, fs.WriteFile(ctx, source3.String(), []byte("content3"), 0644))

	// Batch 1: two operations in parallel
	batch1Op1 := domain.NewLinkCreate("link1", source1, target1)
	batch1Op2 := domain.NewLinkCreate("link2", source2, target2)

	// Batch 2: depends on batch 1 completing
	batch2Op := domain.NewLinkCreate("link3", source3, target3)

	plan := domain.Plan{
		Operations: []domain.Operation{batch1Op1, batch1Op2, batch2Op},
		Batches: [][]domain.Operation{
			{batch1Op1, batch1Op2}, // Batch 1
			{batch2Op},             // Batch 2
		},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")
	execResult := result.Unwrap()

	require.Len(t, execResult.Executed, 3, "all operations should execute")
	require.Empty(t, execResult.Failed)

	// Verify all links created
	require.True(t, fs.Exists(ctx, target1.String()))
	require.True(t, fs.Exists(ctx, target2.String()))
	require.True(t, fs.Exists(ctx, target3.String()))
}

// concurrencyTrackingOp is an operation that tracks concurrent execution count.
type concurrencyTrackingOp struct {
	id            domain.OperationID
	activeCount   *int32
	maxObserved   *int32
	executionTime time.Duration
	mu            *sync.Mutex
}

func (o *concurrencyTrackingOp) ID() domain.OperationID                    { return o.id }
func (o *concurrencyTrackingOp) Kind() domain.OperationKind                { return domain.OpKindLinkCreate }
func (o *concurrencyTrackingOp) Validate() error                           { return nil }
func (o *concurrencyTrackingOp) Dependencies() []domain.Operation          { return nil }
func (o *concurrencyTrackingOp) Rollback(context.Context, domain.FS) error { return nil }
func (o *concurrencyTrackingOp) String() string                            { return string(o.id) }
func (o *concurrencyTrackingOp) Equals(other domain.Operation) bool {
	if other == nil {
		return false
	}
	return o.id == other.ID()
}

func (o *concurrencyTrackingOp) Execute(ctx context.Context, fs domain.FS) error {
	// Increment active count
	current := atomic.AddInt32(o.activeCount, 1)

	// Update max observed (thread-safe)
	o.mu.Lock()
	if current > *o.maxObserved {
		*o.maxObserved = current
	}
	o.mu.Unlock()

	// Simulate work
	time.Sleep(o.executionTime)

	// Decrement active count
	atomic.AddInt32(o.activeCount, -1)

	return nil
}

func TestExecuteBatch_ConcurrencyLimit(t *testing.T) {
	tests := []struct {
		name              string
		batchSize         int
		concurrencyLimit  int
		expectedMaxActive int32
	}{
		{
			name:              "limit 1 serializes execution",
			batchSize:         5,
			concurrencyLimit:  1,
			expectedMaxActive: 1,
		},
		{
			name:              "limit 2 allows max 2 concurrent",
			batchSize:         5,
			concurrencyLimit:  2,
			expectedMaxActive: 2,
		},
		{
			name:             "limit 0 uses runtime.NumCPU",
			batchSize:        20, // Large batch to test NumCPU limit
			concurrencyLimit: 0,
			// expectedMaxActive will be min(runtime.NumCPU(), batchSize)
			// We can't know exact NumCPU, so test is adjusted below
			expectedMaxActive: int32(min(runtime.NumCPU(), 20)),
		},
		{
			name:              "negative limit means no limit",
			batchSize:         4,
			concurrencyLimit:  -1,
			expectedMaxActive: 4,
		},
		{
			name:              "limit greater than batch uses batch size",
			batchSize:         3,
			concurrencyLimit:  10,
			expectedMaxActive: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := adapters.NewMemFS()
			exec := New(Opts{
				FS:          fs,
				Logger:      adapters.NewNoopLogger(),
				Tracer:      adapters.NewNoopTracer(),
				Concurrency: tt.concurrencyLimit,
			})

			var activeCount int32
			var maxObserved int32
			var mu sync.Mutex

			ops := make([]domain.Operation, tt.batchSize)
			for i := 0; i < tt.batchSize; i++ {
				ops[i] = &concurrencyTrackingOp{
					id:            domain.OperationID("op-" + string(rune('a'+i))),
					activeCount:   &activeCount,
					maxObserved:   &maxObserved,
					executionTime: 50 * time.Millisecond,
					mu:            &mu,
				}
			}

			checkpoint := exec.checkpoint.Create(ctx)
			result := exec.executeBatch(ctx, ops, checkpoint)

			require.Len(t, result.Executed, tt.batchSize, "all operations should complete")
			require.Empty(t, result.Failed)
			require.LessOrEqual(t, maxObserved, tt.expectedMaxActive,
				"max concurrent operations should not exceed limit")
		})
	}
}
