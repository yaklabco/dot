package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// TestRegression_AdoptUnmanage_PreservesPackageDirectory is a regression test for a critical bug
// where unmanaging an adopted directory would leave the package directory empty.
//
// Bug History:
//   - Date: 2025-10-09
//   - Issue: When adopting a directory (e.g., .ssh), then unmanaging it, the package
//     directory would exist but be empty. Files were not preserved.
//   - Root Cause: MemFS.Rename() only renamed the directory entry, not its children.
//     When FileMove used Rename() to move a directory, child files stayed at the old path.
//   - Fix: Updated MemFS.Rename() to recursively rename all children when renaming a directory.
//   - Impact: Broke adopt/unmanage workflow for directories, leaving users without access
//     to their files.
//
// This test ensures:
// 1. Directory adoption moves the directory AND all its contents
// 2. Unmanaging restores the directory AND preserves it in the package directory
// 3. All file contents are preserved through the full cycle
func TestRegression_AdoptUnmanage_PreservesPackageDirectory(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup: Create a directory with multiple files (simulating .ssh)
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.ssh", 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/config", []byte("ssh config content"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/known_hosts", []byte("host keys"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.ssh/id_ed25519", []byte("private key"), 0600))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Step 1: Adopt the directory
	err = client.Adopt(ctx, []string{".ssh"}, "dot-ssh")
	require.NoError(t, err, "Adopt should succeed")

	// Verify adoption created flat structure at package root
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh"), "Package directory should exist")

	// CRITICAL: Verify all files at package root (flat structure - regression check)
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/config"), "config file should be at package root")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/known_hosts"), "known_hosts should be at package root")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/id_ed25519"), "id_ed25519 should be at package root")

	// Verify contents are intact
	data, err := fs.ReadFile(ctx, packageDir+"/dot-ssh/config")
	require.NoError(t, err)
	assert.Equal(t, []byte("ssh config content"), data, "File content should be preserved during adopt")

	// Verify symlink was created in target
	isLink, err := fs.IsSymlink(ctx, targetDir+"/.ssh")
	require.NoError(t, err)
	assert.True(t, isLink, "Target should be a symlink")

	// Step 2: Unmanage (restore)
	err = client.Unmanage(ctx, "dot-ssh")
	require.NoError(t, err, "Unmanage should succeed")

	// Verify directory was restored to target
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh"), "Directory should be restored to target")

	// Verify it's a real directory, not a symlink
	isLink, err = fs.IsSymlink(ctx, targetDir+"/.ssh")
	require.NoError(t, err)
	assert.False(t, isLink, "Restored directory should not be a symlink")

	// Verify all files are back in target
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh/config"))
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh/known_hosts"))
	assert.True(t, fs.Exists(ctx, targetDir+"/.ssh/id_ed25519"))

	// Verify contents in target
	data, err = fs.ReadFile(ctx, targetDir+"/.ssh/config")
	require.NoError(t, err)
	assert.Equal(t, []byte("ssh config content"), data, "File content should be preserved in target")

	// CRITICAL REGRESSION CHECK: Files should ALSO remain in package directory
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh"), "Package directory should still exist after unmanage")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/config"), "config should remain at package root")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/known_hosts"), "known_hosts should remain at package root")
	assert.True(t, fs.Exists(ctx, packageDir+"/dot-ssh/id_ed25519"), "id_ed25519 should remain at package root")

	// Verify contents in package directory
	data, err = fs.ReadFile(ctx, packageDir+"/dot-ssh/config")
	require.NoError(t, err)
	assert.Equal(t, []byte("ssh config content"), data, "Package directory should have same content as target")
}

// TestRegression_MemFS_RenameDirectory_WithChildren is a unit regression test
// for the MemFS.Rename() bug that only renamed the directory, not its children.
//
// This test is at the MemFS adapter level to directly test the fix.
func TestRegression_MemFS_RenameDirectory_WithChildren(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Create a directory structure
	require.NoError(t, fs.MkdirAll(ctx, "/old/path/dir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/old/path/dir/file1.txt", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, "/old/path/dir/file2.txt", []byte("content2"), 0644))
	require.NoError(t, fs.MkdirAll(ctx, "/old/path/dir/subdir", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/old/path/dir/subdir/nested.txt", []byte("nested"), 0644))

	// Create destination parent
	require.NoError(t, fs.MkdirAll(ctx, "/new/path", 0755))

	// Rename the directory
	err := fs.Rename(ctx, "/old/path/dir", "/new/path/dir")
	require.NoError(t, err)

	// Verify old location is completely gone
	assert.False(t, fs.Exists(ctx, "/old/path/dir"), "Old directory should not exist")
	assert.False(t, fs.Exists(ctx, "/old/path/dir/file1.txt"), "Old files should not exist")
	assert.False(t, fs.Exists(ctx, "/old/path/dir/subdir/nested.txt"), "Old nested files should not exist")

	// Verify new location has EVERYTHING
	assert.True(t, fs.Exists(ctx, "/new/path/dir"), "New directory should exist")
	assert.True(t, fs.Exists(ctx, "/new/path/dir/file1.txt"), "Files should be at new location")
	assert.True(t, fs.Exists(ctx, "/new/path/dir/file2.txt"), "All files should be renamed")
	assert.True(t, fs.Exists(ctx, "/new/path/dir/subdir/nested.txt"), "Nested files should be renamed")

	// Verify contents are intact
	data, err := fs.ReadFile(ctx, "/new/path/dir/file1.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), data)

	data, err = fs.ReadFile(ctx, "/new/path/dir/subdir/nested.txt")
	require.NoError(t, err)
	assert.Equal(t, []byte("nested"), data)
}

// TestRegression_UnmanageCopiesNotMoves verifies that unmanaging adopted files
// uses copy (not move) so files remain in the package directory.
//
// Expected behavior:
// - Individual files: Copied using FileBackup operation
// - Directories: Copied using DirCopy operation (recursive)
// - Result: Files exist in BOTH target and package directories after unmanage
func TestRegression_UnmanageCopiesNotMoves(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Test with a file
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("vim settings"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt file
	err = client.Adopt(ctx, []string{".vimrc"}, "vim")
	require.NoError(t, err)

	// File should be in package
	assert.True(t, fs.Exists(ctx, packageDir+"/vim/dot-vimrc"))

	// Unmanage
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// CRITICAL: File should exist in BOTH locations
	assert.True(t, fs.Exists(ctx, targetDir+"/.vimrc"), "File should be restored to target")
	assert.True(t, fs.Exists(ctx, packageDir+"/vim/dot-vimrc"), "File should ALSO remain in package (copied, not moved)")

	// Verify both have same content
	targetData, _ := fs.ReadFile(ctx, targetDir+"/.vimrc")
	pkgData, _ := fs.ReadFile(ctx, packageDir+"/vim/dot-vimrc")
	assert.Equal(t, targetData, pkgData, "Both copies should have identical content")
}
