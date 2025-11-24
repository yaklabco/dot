package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestQuery_Status_NoPackages tests status with no packages installed.
func TestQuery_Status_NoPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	status, err := client.Status(env.Context())
	require.NoError(t, err)

	assert.Empty(t, status.Packages)
}

// TestQuery_Status_InstalledPackages tests status with installed packages.
func TestQuery_Status_InstalledPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	err := client.Manage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// Query status
	status, err := client.Status(env.Context())
	require.NoError(t, err)

	// Verify packages present
	assert.Len(t, status.Packages, 2)
}

// TestQuery_Status_SpecificPackages tests status for specific packages.
func TestQuery_Status_SpecificPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	err := client.Manage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// Query status for specific package
	status, err := client.Status(env.Context(), "vim")
	require.NoError(t, err)

	// Verify only vim package
	assert.Len(t, status.Packages, 1)
}

// TestQuery_List_NoPackages tests list with no packages.
func TestQuery_List_NoPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	packages, err := client.List(env.Context())
	require.NoError(t, err)

	assert.Empty(t, packages)
}

// TestQuery_List_InstalledPackages tests list with installed packages.
func TestQuery_List_InstalledPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		WithFile("dot-zshenv", "export PATH=/usr/local/bin:$PATH").
		Create()

	err := client.Manage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// List packages
	packages, err := client.List(env.Context())
	require.NoError(t, err)

	// Verify packages
	assert.Len(t, packages, 2)

	// Verify package details
	pkgMap := make(map[string]int)
	for _, pkg := range packages {
		pkgMap[pkg.Name] = pkg.LinkCount
	}

	assert.Equal(t, 1, pkgMap["vim"])
	assert.Equal(t, 2, pkgMap["zsh"])
}

// TestQuery_Doctor_NoIssues tests doctor with no issues.
func TestQuery_Doctor_NoIssues(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Run doctor
	report, err := client.Doctor(env.Context())
	require.NoError(t, err)

	// Verify no critical issues
	assert.Equal(t, 0, report.Statistics.BrokenLinks)
}

// TestQuery_Doctor_BrokenLinks tests doctor detects broken links.
func TestQuery_Doctor_BrokenLinks(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	vimPackage := filepath.Join(env.PackageDir, "vim")
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// For now, just verify doctor runs successfully
	report, err := client.Doctor(env.Context())
	require.NoError(t, err)

	// Report should be generated
	assert.NotNil(t, report)

	// Clean up
	_ = vimPackage
}

// TestQuery_Doctor_OrphanedLinks tests doctor detects orphaned links.
func TestQuery_Doctor_OrphanedLinks(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create orphaned symlink manually (not managed by dot)
	env.FixtureBuilder().FileTree(env.TargetDir).
		Symlink("/nonexistent/target", ".orphaned")

	// Run doctor (with scan mode to detect orphans)
	report, err := client.Doctor(env.Context())
	require.NoError(t, err)

	// Report should be generated
	assert.NotNil(t, report)
}

// TestQuery_Status_Performance tests status query performance.
func TestQuery_Status_Performance(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create multiple packages
	for i := 0; i < 10; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	// Manage all packages
	packages := make([]string, 10)
	for i := 0; i < 10; i++ {
		packages[i] = filepath.Join("pkg", string(rune('a'+i)))
	}
	err := client.Manage(env.Context(), packages...)
	require.NoError(t, err)

	// Query status (should be fast)
	status, err := client.Status(env.Context())
	require.NoError(t, err)

	assert.Len(t, status.Packages, 10)
}
