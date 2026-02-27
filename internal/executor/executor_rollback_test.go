package executor

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
)

// failingFS wraps a MemFS and fails on specific operations.
// It allows configuring which operation (by index) should fail.
type failingFS struct {
	inner          *adapters.MemFS
	failOnOpIndex  int   // Which operation index to fail on (0-based)
	currentOpIndex int   // Tracks the current operation count
	failError      error // The error to return when failing
	failOnSymlink  bool  // Whether failure occurs on Symlink calls
	failOnMkdirAll bool  // Whether failure occurs on MkdirAll calls
	symlinkCount   int   // Count of Symlink calls
	mkdirAllCount  int   // Count of MkdirAll calls
	targetOpIndex  int   // Which specific call index to fail on
}

// newFailingFS creates a failing FS that fails on the nth operation.
// failOnIndex is 0-based: 0 means fail on the first operation.
func newFailingFS(failOnIndex int) *failingFS {
	return &failingFS{
		inner:         adapters.NewMemFS(),
		failOnOpIndex: failOnIndex,
		failError:     errors.New("injected failure"),
		failOnSymlink: true, // Default to failing on Symlink calls
		targetOpIndex: failOnIndex,
	}
}

// newFailingFSOnMkdir creates a failing FS that fails on MkdirAll calls.
func newFailingFSOnMkdir(failOnIndex int) *failingFS {
	return &failingFS{
		inner:          adapters.NewMemFS(),
		failOnOpIndex:  failOnIndex,
		failError:      errors.New("injected mkdir failure"),
		failOnMkdirAll: true,
		targetOpIndex:  failOnIndex,
	}
}

// Delegated methods that don't fail
func (f *failingFS) Stat(ctx context.Context, path string) (domain.FileInfo, error) {
	return f.inner.Stat(ctx, path)
}

func (f *failingFS) Lstat(ctx context.Context, path string) (domain.FileInfo, error) {
	return f.inner.Lstat(ctx, path)
}

func (f *failingFS) ReadDir(ctx context.Context, path string) ([]domain.DirEntry, error) {
	return f.inner.ReadDir(ctx, path)
}

func (f *failingFS) ReadLink(ctx context.Context, path string) (string, error) {
	return f.inner.ReadLink(ctx, path)
}

func (f *failingFS) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return f.inner.ReadFile(ctx, path)
}

func (f *failingFS) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	return f.inner.WriteFile(ctx, path, data, perm)
}

func (f *failingFS) Mkdir(ctx context.Context, path string, perm fs.FileMode) error {
	return f.inner.Mkdir(ctx, path, perm)
}

func (f *failingFS) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	if f.failOnMkdirAll {
		if f.mkdirAllCount == f.targetOpIndex {
			f.mkdirAllCount++
			return f.failError
		}
		f.mkdirAllCount++
	}
	return f.inner.MkdirAll(ctx, path, perm)
}

func (f *failingFS) Remove(ctx context.Context, path string) error {
	return f.inner.Remove(ctx, path)
}

func (f *failingFS) RemoveAll(ctx context.Context, path string) error {
	return f.inner.RemoveAll(ctx, path)
}

func (f *failingFS) Symlink(ctx context.Context, oldname, newname string) error {
	if f.failOnSymlink {
		if f.symlinkCount == f.targetOpIndex {
			f.symlinkCount++
			return f.failError
		}
		f.symlinkCount++
	}
	return f.inner.Symlink(ctx, oldname, newname)
}

func (f *failingFS) Rename(ctx context.Context, oldpath, newpath string) error {
	return f.inner.Rename(ctx, oldpath, newpath)
}

func (f *failingFS) Exists(ctx context.Context, path string) bool {
	return f.inner.Exists(ctx, path)
}

func (f *failingFS) IsDir(ctx context.Context, path string) (bool, error) {
	return f.inner.IsDir(ctx, path)
}

func (f *failingFS) IsSymlink(ctx context.Context, path string) (bool, error) {
	return f.inner.IsSymlink(ctx, path)
}

// TestExecutor_RollbackOnFailure tests that the executor correctly rolls back
// all previously executed operations when a failure occurs mid-execution.
func TestExecutor_RollbackOnFailure(t *testing.T) {
	ctx := context.Background()

	// Create failing FS that fails on the 3rd symlink operation (index 2)
	failFS := newFailingFS(2)

	// Setup initial filesystem state
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file2", []byte("content2"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file3", []byte("content3"), 0644))

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create a plan with 3 link operations
	// Operations 0 and 1 will succeed, operation 2 will fail
	source1 := domain.MustParsePath("/packages/pkg/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	source3 := domain.MustParsePath("/packages/pkg/file3")
	target1 := domain.MustParseTargetPath("/home/file1")
	target2 := domain.MustParseTargetPath("/home/file2")
	target3 := domain.MustParseTargetPath("/home/file3")

	op1 := domain.NewLinkCreate("link1", source1, target1)
	op2 := domain.NewLinkCreate("link2", source2, target2)
	op3 := domain.NewLinkCreate("link3", source3, target3) // This will fail

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2, op3},
	}

	// Execute the plan
	result := exec.Execute(ctx, plan)

	// Verify: error returned
	require.True(t, result.IsErr(), "execution should fail")

	// Verify: we got an ErrExecutionFailed error
	var execErr domain.ErrExecutionFailed
	require.True(t, errors.As(result.UnwrapErr(), &execErr), "error should be ErrExecutionFailed")

	// Verify: 2 operations were executed before failure
	require.Equal(t, 2, execErr.Executed, "should have executed 2 operations before failure")

	// Verify: 1 operation failed
	require.Equal(t, 1, execErr.Failed, "should have 1 failed operation")

	// Verify: rollback operations executed (2 operations should be rolled back)
	require.Equal(t, 2, execErr.RolledBack, "should have rolled back 2 operations")

	// Verify: filesystem state is restored to original
	// Links created by op1 and op2 should have been removed by rollback
	require.False(t, failFS.Exists(ctx, "/home/file1"), "link1 should be removed after rollback")
	require.False(t, failFS.Exists(ctx, "/home/file2"), "link2 should be removed after rollback")
	require.False(t, failFS.Exists(ctx, "/home/file3"), "link3 should not exist (operation failed)")

	// Verify: original files still exist (they were never modified)
	require.True(t, failFS.Exists(ctx, "/packages/pkg/file1"), "source file1 should still exist")
	require.True(t, failFS.Exists(ctx, "/packages/pkg/file2"), "source file2 should still exist")
	require.True(t, failFS.Exists(ctx, "/packages/pkg/file3"), "source file3 should still exist")
}

// TestExecutor_RollbackOnFirstOperation tests rollback when the very first operation fails.
func TestExecutor_RollbackOnFirstOperation(t *testing.T) {
	ctx := context.Background()

	// Create failing FS that fails on the first symlink operation (index 0)
	failFS := newFailingFS(0)

	// Setup initial filesystem state
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file2", []byte("content2"), 0644))

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source1 := domain.MustParsePath("/packages/pkg/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target1 := domain.MustParseTargetPath("/home/file1")
	target2 := domain.MustParseTargetPath("/home/file2")

	op1 := domain.NewLinkCreate("link1", source1, target1) // This will fail
	op2 := domain.NewLinkCreate("link2", source2, target2)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2},
	}

	result := exec.Execute(ctx, plan)

	// Verify: error returned
	require.True(t, result.IsErr(), "execution should fail")

	var execErr domain.ErrExecutionFailed
	require.True(t, errors.As(result.UnwrapErr(), &execErr), "error should be ErrExecutionFailed")

	// No operations executed before failure, so nothing to rollback
	require.Equal(t, 0, execErr.Executed, "should have executed 0 operations")
	require.Equal(t, 1, execErr.Failed, "should have 1 failed operation")
	require.Equal(t, 0, execErr.RolledBack, "should have rolled back 0 operations")

	// Filesystem should be unchanged
	require.False(t, failFS.Exists(ctx, "/home/file1"), "link1 should not exist")
	require.False(t, failFS.Exists(ctx, "/home/file2"), "link2 should not exist")
}

// TestExecutor_RollbackMultipleBatches tests rollback with mixed operation types
// including directory creation and link creation.
func TestExecutor_RollbackMixedOperations(t *testing.T) {
	ctx := context.Background()

	// Create failing FS that fails on the 2nd symlink operation (index 1)
	failFS := newFailingFS(1)

	// Setup initial filesystem state
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file2", []byte("content2"), 0644))

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create a plan with directory creation followed by link operations
	dirPath := domain.MustParsePath("/home/subdir")
	source1 := domain.MustParsePath("/packages/pkg/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target1 := domain.MustParseTargetPath("/home/file1")
	target2 := domain.MustParseTargetPath("/home/file2")

	op1 := domain.NewDirCreate("dir1", dirPath)
	op2 := domain.NewLinkCreate("link1", source1, target1)
	op3 := domain.NewLinkCreate("link2", source2, target2) // This will fail (2nd symlink call)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2, op3},
	}

	result := exec.Execute(ctx, plan)

	// Verify: error returned
	require.True(t, result.IsErr(), "execution should fail")

	var execErr domain.ErrExecutionFailed
	require.True(t, errors.As(result.UnwrapErr(), &execErr), "error should be ErrExecutionFailed")

	// 2 operations executed before failure (dir create + first link)
	require.Equal(t, 2, execErr.Executed, "should have executed 2 operations")
	require.Equal(t, 1, execErr.Failed, "should have 1 failed operation")
	require.Equal(t, 2, execErr.RolledBack, "should have rolled back 2 operations")

	// Verify rollback: directory and link should be removed
	require.False(t, failFS.Exists(ctx, "/home/subdir"), "directory should be removed after rollback")
	require.False(t, failFS.Exists(ctx, "/home/file1"), "link1 should be removed after rollback")
	require.False(t, failFS.Exists(ctx, "/home/file2"), "link2 should not exist (operation failed)")
}

// TestExecutor_RollbackWithDirCreate tests that directory creation is properly
// rolled back when a subsequent operation fails.
func TestExecutor_RollbackWithDirCreate(t *testing.T) {
	ctx := context.Background()

	// Create failing FS that fails on the first symlink operation
	failFS := newFailingFS(0)

	// Setup initial filesystem state
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create directory first, then a link that will fail
	dirPath := domain.MustParsePath("/home/config")
	source1 := domain.MustParsePath("/packages/pkg/file1")
	target1 := domain.MustParseTargetPath("/home/config/file1")

	op1 := domain.NewDirCreate("dir1", dirPath)
	op2 := domain.NewLinkCreate("link1", source1, target1) // This will fail (first symlink)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr(), "execution should fail")

	var execErr domain.ErrExecutionFailed
	require.True(t, errors.As(result.UnwrapErr(), &execErr))

	// 1 operation executed (dir create), then link failed
	require.Equal(t, 1, execErr.Executed)
	require.Equal(t, 1, execErr.Failed)
	require.Equal(t, 1, execErr.RolledBack)

	// Directory should be rolled back
	require.False(t, failFS.Exists(ctx, "/home/config"), "directory should be removed after rollback")
}

// TestExecutor_RollbackOrderIsReverse tests that rollback occurs in reverse order.
func TestExecutor_RollbackOrderIsReverse(t *testing.T) {
	ctx := context.Background()

	// This test verifies that operations are rolled back in reverse order
	// by checking that parent directories can be removed (which requires
	// child links to be removed first).

	failFS := newFailingFS(2) // Fail on 3rd symlink

	// Setup
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file2", []byte("content2"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file3", []byte("content3"), 0644))

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	// Create operations: dir create, two links in that dir, then a failing link
	dirPath := domain.MustParsePath("/home/config")
	source1 := domain.MustParsePath("/packages/pkg/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	source3 := domain.MustParsePath("/packages/pkg/file3")
	target1 := domain.MustParseTargetPath("/home/file1")
	target2 := domain.MustParseTargetPath("/home/file2")
	target3 := domain.MustParseTargetPath("/home/file3")

	op1 := domain.NewDirCreate("dir1", dirPath)
	op2 := domain.NewLinkCreate("link1", source1, target1)
	op3 := domain.NewLinkCreate("link2", source2, target2)
	op4 := domain.NewLinkCreate("link3", source3, target3) // This will fail (3rd symlink)

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2, op3, op4},
	}

	result := exec.Execute(ctx, plan)

	require.True(t, result.IsErr())

	var execErr domain.ErrExecutionFailed
	require.True(t, errors.As(result.UnwrapErr(), &execErr))

	// 3 operations executed, 1 failed, all 3 should be rolled back
	require.Equal(t, 3, execErr.Executed)
	require.Equal(t, 1, execErr.Failed)
	require.Equal(t, 3, execErr.RolledBack)

	// All artifacts should be cleaned up
	require.False(t, failFS.Exists(ctx, "/home/config"), "directory should be removed")
	require.False(t, failFS.Exists(ctx, "/home/file1"), "link1 should be removed")
	require.False(t, failFS.Exists(ctx, "/home/file2"), "link2 should be removed")
}

// TestExecutor_RollbackPreservesInitialState tests that after rollback,
// the filesystem state matches the initial state before execution.
func TestExecutor_RollbackPreservesInitialState(t *testing.T) {
	ctx := context.Background()

	failFS := newFailingFS(1)

	// Setup initial state with some pre-existing content
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/packages/pkg", 0755))
	require.NoError(t, failFS.inner.MkdirAll(ctx, "/home/existing", 0755))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file1", []byte("content1"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/packages/pkg/file2", []byte("content2"), 0644))
	require.NoError(t, failFS.inner.WriteFile(ctx, "/home/existing/old-file", []byte("old"), 0644))

	// Record initial state
	initialDirs := []string{"/packages", "/packages/pkg", "/home", "/home/existing"}
	initialFiles := map[string][]byte{
		"/packages/pkg/file1":     []byte("content1"),
		"/packages/pkg/file2":     []byte("content2"),
		"/home/existing/old-file": []byte("old"),
	}

	exec := New(Opts{
		FS:     failFS,
		Logger: adapters.NewNoopLogger(),
		Tracer: adapters.NewNoopTracer(),
	})

	source1 := domain.MustParsePath("/packages/pkg/file1")
	source2 := domain.MustParsePath("/packages/pkg/file2")
	target1 := domain.MustParseTargetPath("/home/file1")
	target2 := domain.MustParseTargetPath("/home/file2")

	op1 := domain.NewLinkCreate("link1", source1, target1)
	op2 := domain.NewLinkCreate("link2", source2, target2) // This will fail

	plan := domain.Plan{
		Operations: []domain.Operation{op1, op2},
	}

	_ = exec.Execute(ctx, plan)

	// Verify all initial directories still exist
	for _, dir := range initialDirs {
		require.True(t, failFS.Exists(ctx, dir), "directory %s should still exist", dir)
	}

	// Verify all initial files still exist with correct content
	for path, expectedContent := range initialFiles {
		require.True(t, failFS.Exists(ctx, path), "file %s should still exist", path)
		content, err := failFS.ReadFile(ctx, path)
		require.NoError(t, err)
		require.Equal(t, expectedContent, content, "file %s should have original content", path)
	}

	// Verify no new files were left behind
	require.False(t, failFS.Exists(ctx, "/home/file1"), "new link should not exist after rollback")
}

// stubFileInfo implements domain.FileInfo for testing.
type stubFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (s *stubFileInfo) Name() string       { return s.name }
func (s *stubFileInfo) Size() int64        { return s.size }
func (s *stubFileInfo) Mode() os.FileMode  { return s.mode }
func (s *stubFileInfo) ModTime() time.Time { return s.modTime }
func (s *stubFileInfo) IsDir() bool        { return s.isDir }
func (s *stubFileInfo) Sys() any           { return nil }
