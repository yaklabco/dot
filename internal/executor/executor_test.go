package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestNewExecutor(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	tracer := adapters.NewNoopTracer()

	exec := New(Opts{
		FS:     fs,
		Logger: logger,
		Tracer: tracer,
	})

	require.NotNil(t, exec)
}

func TestNewExecutor_DefaultCheckpoint(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	tracer := adapters.NewNoopTracer()

	exec := New(Opts{
		FS:     fs,
		Logger: logger,
		Tracer: tracer,
		// No checkpoint store provided - should use default
	})

	require.NotNil(t, exec)
	require.NotNil(t, exec.checkpoint)
}

func TestExecute_EmptyPlan(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Empty plan
	plan := domain.Plan{
		Operations: []domain.Operation{},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr())
	require.IsType(t, domain.ErrEmptyPlan{}, result.UnwrapErr())
}

func TestExecute_SingleOperation_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create source file with parent directories
	source := domain.MustParsePath("/packages/pkg/file")
	target := domain.MustParseTargetPath("/home/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	// Create operation
	op := domain.NewLinkCreate("link1", source, target)

	plan := domain.Plan{
		Operations: []domain.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsOk(), "execution should succeed")

	// Verify symlink created
	exists := fs.Exists(ctx, target.String())
	require.True(t, exists, "symlink should be created")
}

func TestExecute_OperationFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create operation that will fail (source doesn't exist)
	source := domain.MustParsePath("/nonexistent")
	target := domain.MustParseTargetPath("/home/file")
	op := domain.NewLinkCreate("link1", source, target)

	plan := domain.Plan{
		Operations: []domain.Operation{op},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")
}

func TestExecute_MultipleOperations_PartialFailure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// First operation succeeds
	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/file1")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source1.String(), []byte("content1"), 0644))
	op1 := domain.NewLinkCreate("link1", source1, target1)

	// Second operation fails (source doesn't exist)
	source2 := domain.MustParsePath("/nonexistent")
	target2 := domain.MustParseTargetPath("/home/file2")
	op2 := domain.NewLinkCreate("link2", source2, target2)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail due to second operation")

	// First operation should have been executed (and then rolled back)
	// We'll verify rollback behavior in later tests
}
