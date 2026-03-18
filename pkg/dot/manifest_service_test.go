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

func TestManifestService_UpdateWithSource_PreservesExistingLinks(t *testing.T) {
	t.Run("remanage preserves links not in current plan", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-bashrc", []byte("bash"), 0644))

		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())

		store := manifest.NewFSManifestStore(fs)
		svc := newManifestService(fs, adapters.NewNoopLogger(), store)

		// Pre-populate manifest with two links (simulating a previous manage)
		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:       "test-pkg",
			LinkCount:  2,
			Links:      []string{".vimrc", ".bashrc"},
			Source:     manifest.SourceManaged,
			TargetDir:  targetDir,
			PackageDir: packageDir + "/test-pkg",
		})
		require.NoError(t, store.Save(ctx, targetPathResult.Unwrap(), m))

		// Simulate a remanage plan that only touches .vimrc (delete + recreate)
		// .bashrc has no operations because it's unchanged
		vimSrc := NewFilePath(packageDir + "/test-pkg/dot-vimrc")
		require.True(t, vimSrc.IsOk())
		vimTgt := NewTargetPath(targetDir + "/.vimrc")
		require.True(t, vimTgt.IsOk())

		plan := Plan{
			Operations: []Operation{
				NewLinkDelete("del-1", vimTgt.Unwrap()),
				NewLinkCreate("create-1", vimSrc.Unwrap(), vimTgt.Unwrap()),
			},
			PackageOperations: map[string][]OperationID{
				"test-pkg": {"del-1", "create-1"},
			},
		}

		err := svc.UpdateWithSource(ctx, targetPathResult.Unwrap(), packageDir, []string{"test-pkg"}, plan, manifest.SourceManaged)
		require.NoError(t, err)

		// Load manifest and verify BOTH links are preserved
		loaded := svc.Load(ctx, targetPathResult.Unwrap())
		require.True(t, loaded.IsOk())
		loadedManifest := loaded.Unwrap()
		pkg, exists := loadedManifest.GetPackage("test-pkg")
		require.True(t, exists)

		assert.Contains(t, pkg.Links, ".vimrc", "touched link should be in manifest")
		assert.Contains(t, pkg.Links, ".bashrc", "untouched link should be preserved in manifest")
		assert.Equal(t, 2, pkg.LinkCount, "link count should reflect all links")
	})

	t.Run("deleted links are removed from manifest", func(t *testing.T) {
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

		// Pre-populate with two links
		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:       "test-pkg",
			LinkCount:  2,
			Links:      []string{".vimrc", ".bashrc"},
			Source:     manifest.SourceManaged,
			TargetDir:  targetDir,
			PackageDir: packageDir + "/test-pkg",
		})
		require.NoError(t, store.Save(ctx, targetPathResult.Unwrap(), m))

		// Plan only deletes .bashrc (no recreate) — file was removed from package
		bashTgt := NewTargetPath(targetDir + "/.bashrc")
		require.True(t, bashTgt.IsOk())

		plan := Plan{
			Operations: []Operation{
				NewLinkDelete("del-1", bashTgt.Unwrap()),
			},
			PackageOperations: map[string][]OperationID{
				"test-pkg": {"del-1"},
			},
		}

		err := svc.UpdateWithSource(ctx, targetPathResult.Unwrap(), packageDir, []string{"test-pkg"}, plan, manifest.SourceManaged)
		require.NoError(t, err)

		loaded := svc.Load(ctx, targetPathResult.Unwrap())
		require.True(t, loaded.IsOk())
		loadedManifest := loaded.Unwrap()
		pkg, exists := loadedManifest.GetPackage("test-pkg")
		require.True(t, exists)

		assert.Contains(t, pkg.Links, ".vimrc", "untouched link should remain")
		assert.NotContains(t, pkg.Links, ".bashrc", "deleted link should be removed")
		assert.Equal(t, 1, pkg.LinkCount)
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
