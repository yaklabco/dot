package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

func TestCheckDirCreatePreconditions_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create parent directory
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	dirPath := domain.MustParsePath("/home/subdir")
	op := domain.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckDirCreatePreconditions_ParentNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Parent doesn't exist
	dirPath := domain.MustParsePath("/nonexistent/subdir")
	op := domain.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, domain.ErrParentNotFound{}, err)
}

func TestCheckDirCreatePreconditions_TopLevelDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create root so top-level directory can be created
	require.NoError(t, fs.MkdirAll(ctx, "/", 0755))

	dirPath := domain.MustParsePath("/toplevel")
	op := domain.NewDirCreate("dir1", dirPath)

	err := exec.checkDirCreatePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckFileMovePreconditions_Success(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Set up source and destination parent
	source := domain.MustParseTargetPath("/home/file")
	dest := domain.MustParsePath("/packages/pkg/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	op := domain.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckFileMovePreconditions_SourceNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := domain.MustParseTargetPath("/nonexistent")
	dest := domain.MustParsePath("/packages/pkg/file")
	require.NoError(t, fs.MkdirAll(ctx, "/packages/pkg", 0755))

	op := domain.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, domain.ErrSourceNotFound{}, err)
}

func TestCheckFileMovePreconditions_DestParentNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := domain.MustParseTargetPath("/home/file")
	dest := domain.MustParsePath("/nonexistent/file")
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, fs.WriteFile(ctx, source.String(), []byte("content"), 0644))

	op := domain.NewFileMove("move1", source, dest)

	err := exec.checkFileMovePreconditions(ctx, op)
	require.Error(t, err)
	require.IsType(t, domain.ErrParentNotFound{}, err)
}

func TestCheckPreconditions_LinkDelete(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	target := domain.MustParseTargetPath("/home/file")
	op := domain.NewLinkDelete("link1", target)

	// LinkDelete has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckPreconditions_DirDelete(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	dirPath := domain.MustParsePath("/home/dir")
	op := domain.NewDirDelete("dir1", dirPath)

	// DirDelete has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}

func TestCheckPreconditions_FileBackup(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	exec := New(Opts{
		FS:     fs,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source := domain.MustParsePath("/home/file")
	backup := domain.MustParsePath("/home/file.bak")
	op := domain.NewFileBackup("backup1", source, backup)

	// FileBackup has no preconditions - should return nil
	err := exec.checkPreconditions(ctx, op)
	require.NoError(t, err)
}
