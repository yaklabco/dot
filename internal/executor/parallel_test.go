package executor

import (
	"context"
	"testing"

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
