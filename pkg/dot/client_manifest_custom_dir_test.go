package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_WithCustomManifestDir(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"
	manifestDir := "/custom/manifest"

	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, manifestDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg/dot-file", []byte("data"), 0644))

	cfg := dot.Config{
		PackageDir:  packageDir,
		TargetDir:   targetDir,
		ManifestDir: manifestDir, // Custom manifest directory
		FS:          fs,
		Logger:      adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "pkg")
	require.NoError(t, err)

	// Verify manifest was created in custom directory
	manifestPath := manifestDir + "/.dot-manifest.json"
	assert.True(t, fs.Exists(ctx, manifestPath), "Manifest should be in custom directory")

	// Should NOT be in target directory
	targetManifestPath := targetDir + "/.dot-manifest.json"
	assert.False(t, fs.Exists(ctx, targetManifestPath), "Manifest should not be in target directory")
}

func TestClient_WithEmptyManifestDir(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg/file", []byte("data"), 0644))

	cfg := dot.Config{
		PackageDir:  packageDir,
		TargetDir:   targetDir,
		ManifestDir: "", // Empty means use target directory (backward compatible)
		FS:          fs,
		Logger:      adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "pkg")
	require.NoError(t, err)

	// Manifest should be in target directory (backward compatible behavior)
	manifestPath := targetDir + "/.dot-manifest.json"
	assert.True(t, fs.Exists(ctx, manifestPath), "Manifest should be in target directory when ManifestDir is empty")
}

func TestManifestStore_CreatesDirectory(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	targetDir := "/test/target"
	manifestDir := "/custom/manifest/deep/nested"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	// Don't create manifest dir - let the store create it

	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())

	store := manifest.NewFSManifestStoreWithDir(fs, manifestDir)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "test",
		LinkCount: 0,
		Links:     []string{},
	})

	err := store.Save(ctx, targetPathResult.Unwrap(), m)
	require.NoError(t, err)

	// Verify manifest directory was created
	assert.True(t, fs.Exists(ctx, manifestDir), "Manifest directory should be auto-created")
	assert.True(t, fs.Exists(ctx, manifestDir+"/.dot-manifest.json"), "Manifest file should exist")
}
