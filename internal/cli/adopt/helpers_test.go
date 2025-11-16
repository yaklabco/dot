package adopt

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateDirectorySize_Simple(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create test directory with files
	testDir := "/test/calc"
	require.NoError(t, fs.MkdirAll(ctx, testDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(testDir, "file1.txt"), []byte("hello"), 0644))    // 5 bytes
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(testDir, "file2.txt"), []byte("world!!!"), 0644)) // 8 bytes

	size := calculateDirectorySize(ctx, fs, testDir)
	assert.Equal(t, int64(13), size) // 5 + 8 = 13 bytes
}

func TestCalculateDirectorySize_WithSubdirectories(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create nested structure
	testDir := "/test/nested"
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(testDir, "subdir"), 0755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(testDir, "root.txt"), []byte("12345"), 0644))            // 5 bytes
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(testDir, "subdir", "nested.txt"), []byte("1234"), 0644)) // 4 bytes

	size := calculateDirectorySize(ctx, fs, testDir)
	assert.Equal(t, int64(9), size) // 5 + 4 = 9 bytes
}

func TestCalculateDirectorySize_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	testDir := "/test/empty"
	require.NoError(t, fs.MkdirAll(ctx, testDir, 0755))

	size := calculateDirectorySize(ctx, fs, testDir)
	assert.Equal(t, int64(0), size)
}

func TestCalculateDirectorySize_NonExistentDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	size := calculateDirectorySize(ctx, fs, "/nonexistent")
	assert.Equal(t, int64(0), size) // Returns 0 on error
}

func TestCreateCandidate_BasicFile(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create a test file
	testPath := "/home/user/.bashrc"
	require.NoError(t, fs.MkdirAll(ctx, filepath.Dir(testPath), 0755))
	require.NoError(t, fs.WriteFile(ctx, testPath, []byte("bashrc content"), 0644))

	// Get file info
	info, err := fs.Stat(ctx, testPath)
	require.NoError(t, err)

	// Create candidate
	candidate := createCandidate(".bashrc", testPath, "/home/user", info, false)

	assert.Equal(t, ".bashrc", candidate.RelPath)
	assert.Contains(t, candidate.Path, "bashrc")
	assert.Equal(t, int64(14), candidate.Size) // "bashrc content" is 14 bytes
	assert.False(t, candidate.IsDir)
	assert.Equal(t, "shell", candidate.Category)
	assert.Equal(t, "bash", candidate.SuggestedPkg)
}

func TestCreateCandidate_Directory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	// Create a test directory
	testPath := "/home/user/.ssh"
	require.NoError(t, fs.MkdirAll(ctx, testPath, 0755))

	// Get dir info
	info, err := fs.Stat(ctx, testPath)
	require.NoError(t, err)

	// Create candidate
	candidate := createCandidate(".ssh", testPath, "/home/user", info, true)

	assert.Equal(t, ".ssh", candidate.RelPath)
	assert.True(t, candidate.IsDir)
	assert.Equal(t, "tool", candidate.Category)
	assert.Equal(t, "ssh", candidate.SuggestedPkg)
}
