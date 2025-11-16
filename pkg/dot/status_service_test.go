package dot

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/manifest"
)

func TestStatusService_checkPackageHealth_Healthy(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create test files in package
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "vimrc"), []byte("test"), 0644))

	// Create symlink
	linkPath := filepath.Join(targetDir, ".vimrc")
	targetPath := filepath.Join(packageDir, "vimrc")
	require.NoError(t, fs.Symlink(ctx, targetPath, linkPath))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - use relative path from targetDir
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.True(t, isHealthy)
	assert.Empty(t, issueType)
}

func TestStatusService_checkPackageHealth_BrokenLinks(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create symlink pointing to non-existent file
	linkPath := filepath.Join(targetDir, ".vimrc")
	targetPath := filepath.Join(packageDir, "nonexistent")
	require.NoError(t, fs.Symlink(ctx, targetPath, linkPath))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - use relative path from targetDir
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.False(t, isHealthy)
	assert.Equal(t, "broken links", issueType)
}

func TestStatusService_checkPackageHealth_WrongTarget(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	wrongPackageDir := "/test/packages/other"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, wrongPackageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create file in wrong package directory
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(wrongPackageDir, "vimrc"), []byte("test"), 0644))

	// Create symlink pointing to wrong package
	linkPath := filepath.Join(targetDir, ".vimrc")
	targetPath := filepath.Join(wrongPackageDir, "vimrc")
	require.NoError(t, fs.Symlink(ctx, targetPath, linkPath))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - use relative path from targetDir
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.False(t, isHealthy)
	assert.Equal(t, "wrong target", issueType)
}

func TestStatusService_checkPackageHealth_MissingLinks(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test with non-existent link - use relative path
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.False(t, isHealthy)
	assert.Equal(t, "missing links", issueType)
}

func TestStatusService_checkPackageHealth_MultipleIssues(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create one broken link
	link1 := filepath.Join(targetDir, ".vimrc")
	target1 := filepath.Join(packageDir, "nonexistent")
	require.NoError(t, fs.Symlink(ctx, target1, link1))

	// Note: .vim link is missing (not created)

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - should report broken links (highest priority) - use relative paths
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc", ".vim"}, packageDir)

	assert.False(t, isHealthy)
	assert.Equal(t, "broken links", issueType)
}

func TestStatusService_List_WithHealthStatus(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, "vim"), 0755))
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, "tmux"), 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create healthy package (vim)
	vimFile := filepath.Join(packageDir, "vim", "vimrc")
	require.NoError(t, fs.WriteFile(ctx, vimFile, []byte("test"), 0644))
	vimLink := filepath.Join(targetDir, ".vimrc")
	require.NoError(t, fs.Symlink(ctx, vimFile, vimLink))

	// Create unhealthy package (tmux) - broken link
	tmuxLink := filepath.Join(targetDir, ".tmux.conf")
	require.NoError(t, fs.Symlink(ctx, filepath.Join(packageDir, "tmux", "nonexistent"), tmuxLink))

	// Create manifest
	targetPathResult := NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:        "vim",
		Source:      manifest.PackageSource("file:///test/packages/vim"),
		InstalledAt: time.Now(),
		LinkCount:   1,
		Links:       []string{".vimrc"}, // Use relative path from targetDir
		PackageDir:  filepath.Join(packageDir, "vim"),
	})
	m.AddPackage(manifest.PackageInfo{
		Name:        "tmux",
		Source:      manifest.PackageSource("file:///test/packages/tmux"),
		InstalledAt: time.Now(),
		LinkCount:   1,
		Links:       []string{".tmux.conf"}, // Use relative path from targetDir
		PackageDir:  filepath.Join(packageDir, "tmux"),
	})

	// Save manifest
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)
	require.NoError(t, manifestSvc.Save(ctx, targetPath, m))

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test List
	packages, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, packages, 2)

	// Check vim is healthy
	var vimPkg *PackageInfo
	var tmuxPkg *PackageInfo
	for i := range packages {
		if packages[i].Name == "vim" {
			vimPkg = &packages[i]
		} else if packages[i].Name == "tmux" {
			tmuxPkg = &packages[i]
		}
	}

	require.NotNil(t, vimPkg, "vim package should exist")
	assert.True(t, vimPkg.IsHealthy, "vim should be healthy")
	assert.Empty(t, vimPkg.IssueType, "vim should have no issues")

	require.NotNil(t, tmuxPkg, "tmux package should exist")
	assert.False(t, tmuxPkg.IsHealthy, "tmux should be unhealthy")
	assert.Equal(t, "broken links", tmuxPkg.IssueType, "tmux should have broken links")
}

func TestStatusService_checkPackageHealth_RelativeSymlinks(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create test file in package
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "vimrc"), []byte("test"), 0644))

	// Create relative symlink
	linkPath := filepath.Join(targetDir, ".vimrc")
	relativeTarget := "../packages/vim/vimrc"
	require.NoError(t, fs.Symlink(ctx, relativeTarget, linkPath))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - use relative path from targetDir
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.True(t, isHealthy)
	assert.Empty(t, issueType)
}

func TestStatusService_checkPackageHealth_NotSymlink(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure
	packageDir := "/test/packages/vim"
	targetDir := "/test/target"
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create a regular file instead of symlink
	linkPath := filepath.Join(targetDir, ".vimrc")
	require.NoError(t, fs.WriteFile(ctx, linkPath, []byte("test"), 0644))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test - use relative path from targetDir
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, packageDir)

	assert.False(t, isHealthy)
	assert.Equal(t, "wrong target", issueType)
}

func TestStatusService_checkPackageHealth_NoPackageDir(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	// Setup test structure - simulating old adopted package without package_dir
	targetDir := "/test/target"
	otherDir := "/test/other"
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, otherDir, 0755))

	// Create file in arbitrary location (not in package dir)
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(otherDir, "vimrc"), []byte("test"), 0644))

	// Create symlink pointing to file outside any package directory
	linkPath := filepath.Join(targetDir, ".vimrc")
	targetPath := filepath.Join(otherDir, "vimrc")
	require.NoError(t, fs.Symlink(ctx, targetPath, linkPath))

	// Create manifest service
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	// Create status service
	svc := newStatusService(fs, logger, manifestSvc, targetDir)

	// Test with empty package directory (simulating old adopted package)
	isHealthy, issueType := svc.checkPackageHealth(ctx, "vim", []string{".vimrc"}, "")

	// Should be healthy because package_dir validation is skipped when empty
	assert.True(t, isHealthy, "Package without package_dir should be healthy if symlink exists and target exists")
	assert.Empty(t, issueType)
}
