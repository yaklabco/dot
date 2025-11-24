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

func TestUnmanageWithOptions_AdoptedPackage_Purge(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/ssh", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/ssh/dot-config", []byte("config"), 0644))
	require.NoError(t, fs.Symlink(ctx, packageDir+"/ssh/dot-config", targetDir+"/.config"))

	// Create manifest with adopted source
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()
	store := manifest.NewFSManifestStore(fs)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "ssh",
		LinkCount: 1,
		Links:     []string{".config"},
		Source:    manifest.SourceAdopted,
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

	// Unmanage with purge
	opts := dot.UnmanageOptions{
		Purge:   true,
		Restore: false,
	}
	err = client.UnmanageWithOptions(ctx, opts, "ssh")
	require.NoError(t, err)

	// Symlink should be removed
	assert.False(t, fs.Exists(ctx, targetDir+"/.config"))

	// Package directory should be purged
	assert.False(t, fs.Exists(ctx, packageDir+"/ssh"))
}

func TestUnmanageWithOptions_ManagedPackage(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	// Setup managed package
	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/vim", 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/vim/dot-vimrc", []byte("vimrc"), 0644))
	require.NoError(t, fs.Symlink(ctx, packageDir+"/vim/dot-vimrc", targetDir+"/.vimrc"))

	// Create manifest with managed source
	targetPathResult := dot.NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()
	store := manifest.NewFSManifestStore(fs)
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      "vim",
		LinkCount: 1,
		Links:     []string{".vimrc"},
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

	// Unmanage managed package (should not restore)
	err = client.Unmanage(ctx, "vim")
	require.NoError(t, err)

	// Symlink should be removed
	assert.False(t, fs.Exists(ctx, targetDir+"/.vimrc"))

	// Package files should remain untouched
	assert.True(t, fs.Exists(ctx, packageDir+"/vim/dot-vimrc"))
}

func TestUnmanageOptions_DefaultsAreCorrect(t *testing.T) {
	opts := dot.DefaultUnmanageOptions()

	assert.False(t, opts.Purge)
	assert.True(t, opts.Restore)
}

func TestUnmanage_MultipleAdoptedPackages(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))

	// Create two files to adopt
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.file1", []byte("content1"), 0644))
	require.NoError(t, fs.WriteFile(ctx, targetDir+"/.file2", []byte("content2"), 0644))

	cfg := dot.Config{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Adopt both files to different packages
	err = client.Adopt(ctx, []string{".file1"}, "pkg1")
	require.NoError(t, err)

	err = client.Adopt(ctx, []string{".file2"}, "pkg2")
	require.NoError(t, err)

	// Unmanage both (should restore both)
	err = client.Unmanage(ctx, "pkg1", "pkg2")
	require.NoError(t, err)

	// Both files should be restored
	assert.True(t, fs.Exists(ctx, targetDir+"/.file1"))
	assert.True(t, fs.Exists(ctx, targetDir+"/.file2"))

	// Both should be regular files, not symlinks
	isLink, _ := fs.IsSymlink(ctx, targetDir+"/.file1")
	assert.False(t, isLink)

	isLink, _ = fs.IsSymlink(ctx, targetDir+"/.file2")
	assert.False(t, isLink)
}
