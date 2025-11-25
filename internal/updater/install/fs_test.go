package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSFileSystem_ReadFile(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	require.NoError(t, os.WriteFile(testFile, content, 0o644))

	fs := OSFileSystem{}
	data, err := fs.ReadFile(testFile)

	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestOSFileSystem_ReadFile_NotFound(t *testing.T) {
	fs := OSFileSystem{}
	_, err := fs.ReadFile("/nonexistent/file.txt")

	assert.Error(t, err)
}

func TestOSFileSystem_ReadDir(t *testing.T) {
	dir := t.TempDir()

	// Create some files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o755))

	fs := OSFileSystem{}
	entries, err := fs.ReadDir(dir)

	require.NoError(t, err)
	assert.Len(t, entries, 3)

	// Check entries by name
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	assert.True(t, names["a.txt"])
	assert.True(t, names["b.txt"])
	assert.True(t, names["subdir"])
}

func TestOSFileSystem_Stat(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0o644))

	fs := OSFileSystem{}
	info, err := fs.Stat(testFile)

	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())
	assert.Equal(t, int64(5), info.Size())
}

func TestOSFileSystem_Stat_NotFound(t *testing.T) {
	fs := OSFileSystem{}
	_, err := fs.Stat("/nonexistent/file.txt")

	assert.Error(t, err)
}

// MockFileSystem provides a test implementation of FileSystem.
type MockFileSystem struct {
	Files map[string][]byte
	Dirs  map[string][]os.DirEntry
}

// ReadFile implements FileSystem.
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

// ReadDir implements FileSystem.
func (m *MockFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	if entries, ok := m.Dirs[path]; ok {
		return entries, nil
	}
	return nil, os.ErrNotExist
}

// Stat implements FileSystem.
func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if _, ok := m.Files[path]; ok {
		// Return a minimal file info
		return mockFileInfo{name: filepath.Base(path)}, nil
	}
	return nil, os.ErrNotExist
}

type mockFileInfo struct {
	name string
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }
