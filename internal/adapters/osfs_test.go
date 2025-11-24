package adapters_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestOSFilesystem_Stat(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test Stat
	info, err := fsys.Stat(ctx, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())
	assert.False(t, info.IsDir())
}

func TestOSFilesystem_ReadDir(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	entries, err := fsys.ReadDir(ctx, tmpDir)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestOSFilesystem_ReadLink(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("test"), 0644)
	os.Symlink(target, link)

	result, err := fsys.ReadLink(ctx, link)
	require.NoError(t, err)
	assert.Equal(t, target, result)
}

func TestOSFilesystem_ReadFile(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(tmpFile, content, 0644)

	data, err := fsys.ReadFile(ctx, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestOSFilesystem_WriteFile(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err := fsys.WriteFile(ctx, tmpFile, content, 0644)
	require.NoError(t, err)

	// Verify file was written
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestOSFilesystem_Mkdir(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "testdir")

	err := fsys.Mkdir(ctx, newDir, 0755)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(newDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestOSFilesystem_MkdirAll(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")

	err := fsys.MkdirAll(ctx, nestedDir, 0755)
	require.NoError(t, err)

	// Verify nested directory was created
	info, err := os.Stat(nestedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestOSFilesystem_Remove(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	err := fsys.Remove(ctx, tmpFile)
	require.NoError(t, err)

	// Verify file was removed
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))
}

func TestOSFilesystem_RemoveAll(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "file.txt"), []byte("test"), 0644)

	err := fsys.RemoveAll(ctx, filepath.Join(tmpDir, "a"))
	require.NoError(t, err)

	// Verify directory tree was removed
	_, err = os.Stat(filepath.Join(tmpDir, "a"))
	assert.True(t, os.IsNotExist(err))
}

func TestOSFilesystem_Symlink(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("test"), 0644)

	err := fsys.Symlink(ctx, target, link)
	require.NoError(t, err)

	// Verify symlink was created
	linkTarget, err := os.Readlink(link)
	require.NoError(t, err)
	assert.Equal(t, target, linkTarget)
}

func TestOSFilesystem_Rename(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	oldName := filepath.Join(tmpDir, "old.txt")
	newName := filepath.Join(tmpDir, "new.txt")

	os.WriteFile(oldName, []byte("test"), 0644)

	err := fsys.Rename(ctx, oldName, newName)
	require.NoError(t, err)

	// Verify file was renamed
	_, err = os.Stat(oldName)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(newName)
	assert.NoError(t, err)
}

func TestOSFilesystem_Exists(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	existing := filepath.Join(tmpDir, "exists.txt")
	missing := filepath.Join(tmpDir, "missing.txt")

	os.WriteFile(existing, []byte("test"), 0644)

	assert.True(t, fsys.Exists(ctx, existing))
	assert.False(t, fsys.Exists(ctx, missing))
}

func TestOSFilesystem_IsDir(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "dir")
	filePath := filepath.Join(tmpDir, "file.txt")

	os.Mkdir(dirPath, 0755)
	os.WriteFile(filePath, []byte("test"), 0644)

	isDir, err := fsys.IsDir(ctx, dirPath)
	require.NoError(t, err)
	assert.True(t, isDir)

	isDir, err = fsys.IsDir(ctx, filePath)
	require.NoError(t, err)
	assert.False(t, isDir)
}

func TestOSFilesystem_IsSymlink(t *testing.T) {
	ctx := context.Background()
	fsys := adapters.NewOSFilesystem()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link.txt")

	os.WriteFile(target, []byte("test"), 0644)
	os.Symlink(target, link)

	isLink, err := fsys.IsSymlink(ctx, link)
	require.NoError(t, err)
	assert.True(t, isLink)

	isLink, err = fsys.IsSymlink(ctx, target)
	require.NoError(t, err)
	assert.False(t, isLink)
}

func TestOSFilesystem_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	fsys := adapters.NewOSFilesystem()

	// Operations should respect cancellation
	tmpDir := t.TempDir()
	_, err := fsys.Stat(ctx, tmpDir)

	// Should return context error
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestOSFileInfoWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test content"), 0644)

	osInfo, err := os.Stat(tmpFile)
	require.NoError(t, err)

	wrapped := adapters.WrapFileInfo(osInfo)

	assert.Equal(t, osInfo.Name(), wrapped.Name())
	assert.Equal(t, osInfo.Size(), wrapped.Size())
	assert.Equal(t, osInfo.Mode(), wrapped.Mode())
	assert.Equal(t, osInfo.IsDir(), wrapped.IsDir())
	assert.Equal(t, osInfo.ModTime(), wrapped.ModTime())
	assert.Equal(t, osInfo.Sys(), wrapped.Sys())
}

func TestOSDirEntryWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	wrapped := adapters.WrapDirEntry(entries[0])

	assert.Equal(t, "test.txt", wrapped.Name())
	assert.False(t, wrapped.IsDir())
	assert.Equal(t, fs.FileMode(0), wrapped.Type())

	// Test Info method
	info, err := wrapped.Info()
	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())
	assert.False(t, info.IsDir())
}
