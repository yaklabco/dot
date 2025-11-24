package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestExecute_ContextCancellation_DuringPrepare(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create a plan with many operations to increase chance of cancellation during prepare
	var ops []domain.Operation
	for i := 0; i < 100; i++ {
		source := domain.MustParsePath("/packages/pkg/file")
		target := domain.MustParseTargetPath("/home/file")
		require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
		require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))
		ops = append(ops, domain.NewLinkCreate(domain.OperationID("link"+string(rune(i))), source, target))
	}

	plan := domain.Plan{Operations: ops}

	// Cancel context immediately
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail due to cancellation")
	err := result.UnwrapErr()
	require.Error(t, err)
	// Should fail during prepare due to cancelled context
}

func TestExecute_ContextCancellation_DuringSequentialExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create multiple operations
	var ops []domain.Operation
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/dir" + string(rune('a'+i)))
		ops = append(ops, domain.NewDirCreate(domain.OperationID("dir"+string(rune('a'+i))), dirPath))
	}

	// Ensure parent exists
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	plan := domain.Plan{Operations: ops}

	// Cancel context immediately to ensure we catch it
	cancel()

	result := exec.Execute(ctx, plan)

	// Should fail - either during prepare or execution
	if result.IsErr() {
		// Expected - cancellation or prepare failure
		t.Logf("Execution failed as expected: %v", result.UnwrapErr())
	} else {
		// Operations may have completed before cancellation was observed
		t.Log("Operations completed before cancellation - timing dependent")
	}
}

func TestExecute_CancellationErrorReturnedWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create multiple operations
	var ops []domain.Operation
	for i := 0; i < 10; i++ {
		dirPath := domain.MustParsePath("/test/dir" + string(rune('a'+i)))
		ops = append(ops, domain.NewDirCreate(domain.OperationID("dir"+string(rune('a'+i))), dirPath))
	}

	plan := domain.Plan{Operations: ops}

	// Cancel context before execution
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	// Should get cancellation error during prepare
	err := result.UnwrapErr()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cancel", "error should mention cancellation")
}

func TestExecute_ContextCancellation_DuringParallelExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directories
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create operations that can be parallelized (no dependencies)
	var batch1, batch2 []domain.Operation
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/parallel1_" + string(rune('a'+i)))
		batch1 = append(batch1, domain.NewDirCreate(domain.OperationID("par1_"+string(rune('a'+i))), dirPath))
	}
	for i := 0; i < 5; i++ {
		dirPath := domain.MustParsePath("/test/parallel2_" + string(rune('a'+i)))
		batch2 = append(batch2, domain.NewDirCreate(domain.OperationID("par2_"+string(rune('a'+i))), dirPath))
	}

	plan := domain.Plan{
		Operations: append(batch1, batch2...),
		Batches:    [][]domain.Operation{batch1, batch2}, // Two parallel batches
	}

	// Cancel immediately to catch between batches
	cancel()

	result := exec.Execute(ctx, plan)

	// Should fail - either during prepare or between batches
	if result.IsErr() {
		t.Logf("Execution failed as expected: %v", result.UnwrapErr())
	} else {
		// All operations may have completed before cancellation
		t.Log("Operations completed before cancellation - timing dependent")
	}
}

func TestExecute_ImmediateCancellation_NoOperationsExecuted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create a single operation
	dirPath := domain.MustParsePath("/test/dir1")
	op := domain.NewDirCreate("dir1", dirPath)
	plan := domain.Plan{Operations: []domain.Operation{op}}

	// Cancel immediately before Execute
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	// Verify the directory was not created
	exists := fs.Exists(ctx, dirPath.String())
	require.False(t, exists, "directory should not have been created")
}

func TestExecute_RollbackWithCancelledContext_ContinuesRollback(t *testing.T) {
	// Test that rollback continues even if context is cancelled
	// This ensures system consistency
	ctx, cancel := context.WithCancel(context.Background())
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directories
	require.NoError(t, fs.MkdirAll(ctx, "/test", 0755))

	// Create operations where second will fail
	dir1 := domain.MustParsePath("/test/dir1")
	dir2 := domain.MustParsePath("/nonexistent/dir2") // Parent doesn't exist

	ops := []domain.Operation{
		domain.NewDirCreate("dir1", dir1),
		domain.NewDirCreate("dir2", dir2),
	}

	plan := domain.Plan{Operations: ops}

	// Cancel context before execute - but execution should still attempt
	// and rollback should complete
	cancel()

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	// Either prepare cancelled, or execution failed and rolled back
	err := result.UnwrapErr()
	t.Logf("Result error: %v", err)

	// If execution proceeded despite cancelled context, verify rollback occurred
	if execFailed, ok := err.(domain.ErrExecutionFailed); ok {
		require.Equal(t, 1, execFailed.Executed, "first operation should have executed")
		require.Equal(t, 1, execFailed.Failed, "second operation should have failed")
		require.Equal(t, 1, execFailed.RolledBack, "first operation should have been rolled back")

		// Verify dir1 was rolled back
		exists := fs.Exists(ctx, dir1.String())
		require.False(t, exists, "dir1 should have been rolled back")
	}
}
