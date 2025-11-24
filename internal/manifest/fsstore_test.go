package manifest

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestFSManifestStore_Load_MissingFile(t *testing.T) {
	fs := adapters.NewMemFS()
	manifestDir := "/home/user/.local/share/dot/manifest"
	store := NewFSManifestStoreWithDir(fs, manifestDir)
	targetDir := mustTargetPath(t, "/home/user")

	result := store.Load(context.Background(), targetDir)

	require.True(t, result.IsOk())
	m := result.Unwrap()
	assert.Equal(t, "1.0", m.Version)
	assert.Empty(t, m.Packages)
}

func TestFSManifestStore_CustomManifestDir(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Custom manifest directory
	manifestDir := "/custom/manifest/dir"
	require.NoError(t, fs.MkdirAll(ctx, manifestDir, 0755))

	store := NewFSManifestStoreWithDir(fs, manifestDir)
	targetDir := mustTargetPath(t, "/home/user")

	// Save manifest
	m := New()
	m.AddPackage(PackageInfo{
		Name:        "test",
		InstalledAt: time.Now(),
		LinkCount:   1,
		Links:       []string{".test"},
	})

	err := store.Save(ctx, targetDir, m)
	require.NoError(t, err)

	// Verify manifest saved in custom directory, not target directory
	manifestPath := filepath.Join(manifestDir, ".dot-manifest.json")
	exists := fs.Exists(ctx, manifestPath)
	assert.True(t, exists, "Manifest should be in custom directory")

	// Should NOT be in target directory
	targetManifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
	exists = fs.Exists(ctx, targetManifestPath)
	assert.False(t, exists, "Manifest should not be in target directory")

	// Load should work from custom directory
	result := store.Load(ctx, targetDir)
	require.True(t, result.IsOk())
	loaded := result.Unwrap()
	assert.Len(t, loaded.Packages, 1)
}

func TestFSManifestStore_Load_ValidManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Create manifest file
	manifestData := `{
  "version": "1.0",
  "updated_at": "2024-01-15T10:30:00Z",
  "packages": {
    "vim": {
      "name": "vim",
      "installed_at": "2024-01-15T10:00:00Z",
      "link_count": 2,
      "links": [".vimrc", ".vim/colors"]
    }
  },
  "hashes": {
    "vim": "abc123"
  }
}`
	manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
	require.NoError(t, fs.WriteFile(context.Background(), manifestPath, []byte(manifestData), 0644))

	store := NewFSManifestStore(fs)
	result := store.Load(context.Background(), targetDir)

	require.True(t, result.IsOk())
	m := result.Unwrap()
	assert.Equal(t, "1.0", m.Version)
	assert.Len(t, m.Packages, 1)

	vim, exists := m.GetPackage("vim")
	assert.True(t, exists)
	assert.Equal(t, 2, vim.LinkCount)
}

func TestFSManifestStore_Load_CorruptManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	// Write invalid JSON
	manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
	require.NoError(t, fs.WriteFile(context.Background(), manifestPath, []byte("invalid json"), 0644))

	store := NewFSManifestStore(fs)
	result := store.Load(context.Background(), targetDir)

	assert.False(t, result.IsOk())
	assert.Error(t, result.UnwrapErr())
}

func TestFSManifestStore_Load_WithContext(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	store := NewFSManifestStore(fs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := store.Load(ctx, targetDir)

	assert.False(t, result.IsOk())
}

func TestFSManifestStore_Save_NewManifest(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	store := NewFSManifestStore(fs)
	m := New()
	m.AddPackage(PackageInfo{
		Name:      "vim",
		LinkCount: 2,
		Links:     []string{".vimrc", ".vim/colors"},
	})

	err := store.Save(context.Background(), targetDir, m)

	require.NoError(t, err)

	// Verify file exists and is readable
	manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
	exists := fs.Exists(context.Background(), manifestPath)
	assert.True(t, exists)

	// Verify content
	result := store.Load(context.Background(), targetDir)
	require.True(t, result.IsOk())
	loaded := result.Unwrap()
	assert.Len(t, loaded.Packages, 1)
}

func TestFSManifestStore_Save_UpdatesTimestamp(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	store := NewFSManifestStore(fs)
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	err := store.Save(context.Background(), targetDir, m)
	require.NoError(t, err)

	result := store.Load(context.Background(), targetDir)
	require.True(t, result.IsOk())
	loaded := result.Unwrap()
	assert.True(t, loaded.UpdatedAt.After(originalTime))
}

func TestFSManifestStore_Save_AtomicWrite(t *testing.T) {
	// This test verifies atomic write by checking temp file is cleaned up
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	store := NewFSManifestStore(fs)
	m := New()

	err := store.Save(context.Background(), targetDir, m)
	require.NoError(t, err)

	// Verify no temp files left behind
	entries, err := fs.ReadDir(context.Background(), targetDir.String())
	require.NoError(t, err)

	for _, entry := range entries {
		assert.NotContains(t, entry.Name(), ".tmp")
		assert.NotContains(t, entry.Name(), "~")
	}
}

func TestFSManifestStore_Save_WithContext(t *testing.T) {
	fs := adapters.NewMemFS()
	targetDir := mustTargetPath(t, "/home/user")
	require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))

	store := NewFSManifestStore(fs)
	m := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := store.Save(ctx, targetDir, m)

	assert.Error(t, err)
}
