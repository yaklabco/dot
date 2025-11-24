package dot

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
)

// TestDoctorService_IgnoreLink tests the IgnoreLink functionality
func TestDoctorService_IgnoreLink(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, store)

	packageDir := "/packages"
	targetDir := "/home"
	require.NoError(t, fs.MkdirAll(ctx, targetDir+"/.config", 0755))

	// Create a symlink to ignore
	linkPath := filepath.Join(targetDir, ".config/app.conf")
	require.NoError(t, fs.Symlink(ctx, "/source", linkPath))

	svc := newDoctorService(fs, logger, manifestSvc, packageDir, targetDir)

	err := svc.IgnoreLink(ctx, ".config/app.conf", "test reason")
	require.NoError(t, err)

	// Verify ignored
	links, _, err := svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Contains(t, links, ".config/app.conf")
	assert.Equal(t, "test reason", links[".config/app.conf"].Reason)
}

// TestDoctorService_IgnorePattern tests pattern ignoring
func TestDoctorService_IgnorePattern(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", "/home")

	require.NoError(t, fs.MkdirAll(ctx, "/home", 0755))

	err := svc.IgnorePattern(ctx, ".cache/**")
	require.NoError(t, err)

	_, patterns, err := svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Contains(t, patterns, ".cache/**")
}

// TestDoctorService_UnignoreLink tests removing from ignore list
func TestDoctorService_UnignoreLink(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)

	targetDir := "/home"
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Setup manifest with ignored link
	targetPathResult := NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	m := manifest.New()
	m.AddIgnoredLink(".test", "/src", "reason")
	require.NoError(t, store.Save(ctx, targetPathResult.Unwrap(), m))

	svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, "/packages", targetDir)

	err := svc.UnignoreLink(ctx, ".test")
	require.NoError(t, err)

	links, _, _ := svc.ListIgnored(ctx)
	assert.NotContains(t, links, ".test")
}

// TestDoctorService_Fix tests fixing broken links
func TestDoctorService_Fix(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)

	packageDir := "/packages"
	targetDir := "/home"
	pkgName := "config"

	// Create package with source
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, pkgName), 0755))
	sourceFile := filepath.Join(packageDir, pkgName, "dot-bashrc")
	require.NoError(t, fs.WriteFile(ctx, sourceFile, []byte("content"), 0644))

	// Create target directory and broken link
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	brokenLink := filepath.Join(targetDir, ".bashrc")
	require.NoError(t, fs.Symlink(ctx, "/wrong", brokenLink))

	// Setup manifest
	targetPathResult := NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:      pkgName,
		Links:     []string{".bashrc"},
		LinkCount: 1,
	})
	require.NoError(t, store.Save(ctx, targetPathResult.Unwrap(), m))

	svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, packageDir, targetDir)

	result, err := svc.Fix(ctx, DefaultScanConfig(), FixOptions{AutoConfirm: true})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Fixed)
}

// TestDoctorService_ConstructSourcePath tests source path construction
func TestDoctorService_constructSourcePath(t *testing.T) {
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", "/home")

	tests := []struct {
		name     string
		pkg      string
		link     string
		expected string
	}{
		{"dotfile", "config", ".bashrc", "/packages/config/dot-bashrc"},
		{"nested in dotdir", "config", ".config/app.conf", "/packages/config/.config/app.conf"},
		{"nested dotfile", "config", ".config/.gitignore", "/packages/config/.config/dot-gitignore"},
		{"regular file", "scripts", "install.sh", "/packages/scripts/install.sh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.constructSourcePath(tt.pkg, tt.link)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterIssuesByType tests issue filtering
func TestFilterIssuesByType(t *testing.T) {
	issues := []Issue{
		{Type: IssueBrokenLink, Path: ".broken"},
		{Type: IssueOrphanedLink, Path: ".orphan1"},
		{Type: IssueOrphanedLink, Path: ".orphan2"},
	}

	orphans := filterIssuesByType(issues, IssueOrphanedLink)
	assert.Len(t, orphans, 2)

	broken := filterIssuesByType(issues, IssueBrokenLink)
	assert.Len(t, broken, 1)
}

// TestDoctorService_isManagedLink tests managed link detection
func TestDoctorService_isManagedLink(t *testing.T) {
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", "/home")

	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "config",
		Links: []string{".bashrc", ".vimrc"},
	})

	assert.True(t, svc.isManagedLink(".bashrc", &m))
	assert.True(t, svc.isManagedLink(".vimrc", &m))
	assert.False(t, svc.isManagedLink(".orphan", &m))
}

// TestDoctorService_findPackageForLink tests package lookup
func TestDoctorService_findPackageForLink(t *testing.T) {
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", "/home")

	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{Name: "config1", Links: []string{".bashrc"}})
	m.AddPackage(manifest.PackageInfo{Name: "config2", Links: []string{".vimrc"}})

	assert.Equal(t, "config1", svc.findPackageForLink(".bashrc", &m))
	assert.Equal(t, "config2", svc.findPackageForLink(".vimrc", &m))
	assert.Empty(t, svc.findPackageForLink(".orphan", &m))
}

// TestDoctorService_fixBrokenUnmanagedLink tests removing unmanaged broken links
func TestDoctorService_fixBrokenUnmanagedLink(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)

	targetDir := "/home"
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	// Create broken unmanaged symlink
	brokenLink := filepath.Join(targetDir, ".orphan")
	require.NoError(t, fs.Symlink(ctx, "/nowhere", brokenLink))

	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", targetDir)

	err := svc.fixBrokenUnmanagedLink(ctx, ".orphan")
	require.NoError(t, err)

	// Verify link was removed
	assert.False(t, fs.Exists(ctx, brokenLink))
}

// TestDoctorService_groupIssuesForFix tests grouping issues
func TestDoctorService_groupIssuesForFix(t *testing.T) {
	fs := adapters.NewMemFS()
	store := manifest.NewFSManifestStore(fs)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), newManifestService(fs, adapters.NewNoopLogger(), store), "/packages", "/home")

	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:  "config",
		Links: []string{".bashrc", ".vimrc"},
	})

	issues := []Issue{
		{Type: IssueBrokenLink, Path: ".bashrc"}, // managed
		{Type: IssueBrokenLink, Path: ".orphan"}, // unmanaged
		{Type: IssueBrokenLink, Path: ".vimrc"},  // managed
		{Type: IssueWrongTarget, Path: ".other"}, // not broken link
	}

	groups := svc.groupIssuesForFix(issues, &m)

	// Should have 2 groups: managed and unmanaged broken links
	assert.Len(t, groups, 2)

	var managedCount, unmanagedCount int
	for _, group := range groups {
		if group.Category == "Managed broken links" {
			managedCount = len(group.Issues)
		} else if group.Category == "Unmanaged broken links" {
			unmanagedCount = len(group.Issues)
		}
	}

	assert.Equal(t, 2, managedCount, "should have 2 managed broken links")
	assert.Equal(t, 1, unmanagedCount, "should have 1 unmanaged broken link")
}

// TestDoctorService_fixBrokenManagedLink tests fixing managed broken links
func TestDoctorService_fixBrokenManagedLink(t *testing.T) {
	ctx := context.Background()

	t.Run("recreates symlink when source exists", func(t *testing.T) {
		fs := adapters.NewMemFS()
		store := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)

		packageDir := "/packages"
		targetDir := "/home"
		pkgName := "myconfig"

		// Create package source
		require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, pkgName), 0755))
		sourceFile := filepath.Join(packageDir, pkgName, "dot-bashrc")
		require.NoError(t, fs.WriteFile(ctx, sourceFile, []byte("content"), 0644))

		// Create target directory and manifest
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())

		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:  pkgName,
			Links: []string{".bashrc"},
		})
		require.NoError(t, store.Save(ctx, targetPathResult.Unwrap(), m))

		// Create broken symlink
		brokenLink := filepath.Join(targetDir, ".bashrc")
		require.NoError(t, fs.Symlink(ctx, "/wrong/path", brokenLink))

		svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, packageDir, targetDir)

		// Fix the link
		err := svc.fixBrokenManagedLink(ctx, pkgName, ".bashrc", &m)
		require.NoError(t, err)

		// Verify link was recreated correctly
		target, err := fs.ReadLink(ctx, brokenLink)
		require.NoError(t, err)
		assert.Equal(t, sourceFile, target)
	})

	t.Run("removes link and updates manifest when source missing", func(t *testing.T) {
		fs := adapters.NewMemFS()
		store := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)

		packageDir := "/packages"
		targetDir := "/home"
		pkgName := "myconfig"

		// Create directories but no source file
		require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, pkgName), 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		// Create manifest with two links
		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())

		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:  pkgName,
			Links: []string{".bashrc", ".vimrc"},
		})

		// Create broken symlink
		brokenLink := filepath.Join(targetDir, ".bashrc")
		require.NoError(t, fs.Symlink(ctx, "/wrong/path", brokenLink))

		svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, packageDir, targetDir)

		// Fix should remove link and update manifest
		err := svc.fixBrokenManagedLink(ctx, pkgName, ".bashrc", &m)
		require.NoError(t, err)

		// Verify link was removed
		assert.False(t, fs.Exists(ctx, brokenLink))

		// Verify manifest was updated (link removed)
		pkg, exists := m.GetPackage(pkgName)
		require.True(t, exists, "package should still exist (has other links)")
		assert.Len(t, pkg.Links, 1, "should have only 1 link left")
		assert.Equal(t, ".vimrc", pkg.Links[0])
	})

	t.Run("removes package when last link removed", func(t *testing.T) {
		fs := adapters.NewMemFS()
		store := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)

		packageDir := "/packages"
		targetDir := "/home"
		pkgName := "myconfig"

		// Create directories but no source file
		require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, pkgName), 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		// Create manifest with single link
		m := manifest.New()
		m.AddPackage(manifest.PackageInfo{
			Name:  pkgName,
			Links: []string{".bashrc"},
		})

		// Create broken symlink
		brokenLink := filepath.Join(targetDir, ".bashrc")
		require.NoError(t, fs.Symlink(ctx, "/wrong/path", brokenLink))

		svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, packageDir, targetDir)

		// Fix should remove link and package
		err := svc.fixBrokenManagedLink(ctx, pkgName, ".bashrc", &m)
		require.NoError(t, err)

		// Verify package was removed from manifest
		_, exists := m.GetPackage(pkgName)
		assert.False(t, exists, "package should be removed when no links remain")
	})
}
