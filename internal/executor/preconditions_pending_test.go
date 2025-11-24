package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestCheckLinkCreatePreconditions_WithPendingFileMove(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create target directory
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))

	// Create destination directory for the move
	require.NoError(t, fs.MkdirAll(ctx, "/dest", 0755))

	// Track that a file will be moved to /dest/file.txt
	pendingFiles := map[string]struct{}{
		"/dest/file.txt": {},
	}

	// Create a LinkCreate operation that uses the pending file as source
	sourcePath := domain.MustParsePath("/dest/file.txt") // Will exist after move
	targetPathResult := domain.NewTargetPath("/target/link.txt")
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()
	op := domain.NewLinkCreate("link1", sourcePath, targetPath)

	// Should succeed because file will exist after pending move
	err := exec.checkLinkCreatePreconditionsWithPending(ctx, op, nil, pendingFiles)
	require.NoError(t, err)
}

func TestCheckLinkCreatePreconditions_WithoutPendingFileMove(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create target directory
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))

	// Create a LinkCreate operation with non-existent source
	sourcePath := domain.MustParsePath("/dest/file.txt") // Does not exist
	targetPathResult := domain.NewTargetPath("/target/link.txt")
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()
	op := domain.NewLinkCreate("link1", sourcePath, targetPath)

	// Should fail because file doesn't exist and no pending move
	err := exec.checkLinkCreatePreconditionsWithPending(ctx, op, nil, nil)
	require.Error(t, err)
	require.IsType(t, domain.ErrSourceNotFound{}, err)
}

func TestPrepare_WithFileMoveThenLinkCreate(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create source file and directories
	require.NoError(t, fs.MkdirAll(ctx, "/source", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/source/file.txt", []byte("test"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, "/dest", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/target", 0755))

	// Create operations: FileMove followed by LinkCreate
	sourcePathResult := domain.NewTargetPath("/source/file.txt")
	require.True(t, sourcePathResult.IsOk())
	sourcePath := sourcePathResult.Unwrap()

	destPath := domain.MustParsePath("/dest/file.txt")

	targetPathResult := domain.NewTargetPath("/target/link.txt")
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	moveOp := domain.NewFileMove("move1", sourcePath, destPath)
	linkOp := domain.NewLinkCreate("link1", destPath, targetPath)

	plan := domain.Plan{
		Operations: []domain.Operation{moveOp, linkOp},
	}

	// Should succeed - prepare tracks the pending file move
	err := exec.prepare(ctx, plan)
	require.NoError(t, err)
}
