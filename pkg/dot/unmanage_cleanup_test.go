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

func TestUnmanage_Cleanup_RemovesOrphanedLinks(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create manifest with orphaned entry (links don't exist)
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	store := manifest.NewFSManifestStore(fs)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "orphaned",
		LinkCount: 2,
		Links:     []string{".vimrc", ".vim"},
		Source:    manifest.SourceManaged,
	})
	require.NoError(t, store.Save(ctx, targetPath, m))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Run cleanup
	opts := dot.UnmanageOptions{
		Cleanup: true,
	}
	err = client.UnmanageWithOptions(ctx, opts, "orphaned")
	require.NoError(t, err)

	// Verify package was removed from manifest
	status, err := client.Status(ctx)
	require.NoError(t, err)
	assert.Empty(t, status.Packages, "Orphaned package should be removed from manifest")
}

func TestUnmanage_Cleanup_RemovesPackageWithMissingDirectory(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create manifest with entry but package directory doesn't exist
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	store := manifest.NewFSManifestStore(fs)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "missing-pkg",
		LinkCount: 1,
		Links:     []string{".config"},
		Source:    manifest.SourceManaged,
	})
	require.NoError(t, store.Save(ctx, targetPath, m))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Run cleanup
	opts := dot.UnmanageOptions{
		Cleanup: true,
	}
	err = client.UnmanageWithOptions(ctx, opts, "missing-pkg")
	require.NoError(t, err)

	// Verify package was removed from manifest
	status, err := client.Status(ctx, "missing-pkg")
	require.NoError(t, err)
	assert.Empty(t, status.Packages)
}

func TestUnmanage_Cleanup_KeepsValidPackages(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup valid package
	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/valid", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/valid/file", []byte("data"), 0644))
	require.NoError(t, fs.Symlink(ctx, packageDir+"/valid/file", targetDir+"/link"))

	// Create manifest
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	store := manifest.NewFSManifestStore(fs)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "valid",
		LinkCount: 1,
		Links:     []string{"link"},
		Source:    manifest.SourceManaged,
	})
	require.NoError(t, store.Save(ctx, targetPath, m))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Run cleanup on valid package (should not error, should be no-op)
	opts := dot.UnmanageOptions{
		Cleanup: true,
	}
	err = client.UnmanageWithOptions(ctx, opts, "valid")
	require.NoError(t, err)

	// Valid package should still exist in manifest
	status, err := client.Status(ctx, "valid")
	require.NoError(t, err)
	assert.Len(t, status.Packages, 1)
}
