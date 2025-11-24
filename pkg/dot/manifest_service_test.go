package dot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
)

func TestManifestService_Load(t *testing.T) {
	t.Run("loads existing manifest", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())
		targetPath := targetPathResult.Unwrap()

		// Create manifest directly
		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:        "test-pkg",
			InstalledAt: time.Now(),
			LinkCount:   1,
			Links:       []string{".vimrc"},
		})

		store := manifest.NewFSManifestStore(fs)
		saveErr := store.Save(ctx, targetPath, m)
		require.NoError(t, saveErr)

		// Load via service
		svc := newManifestService(fs, adapters.NewNoopLogger(), store)
		result := svc.Load(ctx, targetPath)

		require.True(t, result.IsOk())
		loaded := result.Unwrap()
		assert.Len(t, loaded.Packages, 1)
	})
}

func TestManifestService_Update(t *testing.T) {
	t.Run("creates new manifest and adds package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())

		store := manifest.NewFSManifestStore(fs)
		svc := newManifestService(fs, adapters.NewNoopLogger(), store)

		// Create plan
		srcPath := NewFilePath(packageDir + "/test-pkg/dot-vimrc")
		tgtPath := NewTargetPath(targetDir + "/.vimrc")
		require.True(t, srcPath.IsOk())
		require.True(t, tgtPath.IsOk())

		plan := Plan{
			Operations: []Operation{
				NewLinkCreate("link-1", srcPath.Unwrap(), tgtPath.Unwrap()),
			},
			PackageOperations: map[string][]OperationID{
				"test-pkg": {"link-1"},
			},
		}

		err := svc.Update(ctx, targetPathResult.Unwrap(), packageDir, []string{"test-pkg"}, plan)
		require.NoError(t, err)

		// Verify manifest created
		loaded := svc.Load(ctx, targetPathResult.Unwrap())
		require.True(t, loaded.IsOk())

		m := loaded.Unwrap()
		pkg, exists := m.GetPackage("test-pkg")
		require.True(t, exists)
		assert.Equal(t, "test-pkg", pkg.Name)
	})
}

func TestManifestService_RemovePackage(t *testing.T) {
	t.Run("removes package from manifest", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())
		targetPath := targetPathResult.Unwrap()

		// Create manifest
		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:        "test-pkg",
			InstalledAt: time.Now(),
			LinkCount:   1,
		})

		store := manifest.NewFSManifestStore(fs)
		saveErr := store.Save(ctx, targetPath, m)
		require.NoError(t, saveErr)

		svc := newManifestService(fs, adapters.NewNoopLogger(), store)
		removeErr := svc.RemovePackage(ctx, targetPath, "test-pkg")
		require.NoError(t, removeErr)

		// Verify removed
		loaded := svc.Load(ctx, targetPath)
		require.True(t, loaded.IsOk())
		reloaded := loaded.Unwrap()
		_, exists := reloaded.GetPackage("test-pkg")
		assert.False(t, exists)
	})
}
