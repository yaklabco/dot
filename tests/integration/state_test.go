package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestState_ManifestCreation tests that manifest is created on first manage.
func TestState_ManifestCreation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify manifest exists
	manifestPath := filepath.Join(env.TargetDir, ".dot-manifest.json")
	assert.FileExists(t, manifestPath)
}

// TestState_ManifestUpdate tests manifest updates on operations.
func TestState_ManifestUpdate(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage first package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Capture manifest state
	manifestPath := filepath.Join(env.TargetDir, ".dot-manifest.json")
	content1, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	// Manage second package
	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	err = client.Manage(env.Context(), "zsh")
	require.NoError(t, err)

	// Verify manifest updated
	content2, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	assert.NotEqual(t, string(content1), string(content2))
}

// TestState_ManifestPreservation tests manifest preservation on errors.
func TestState_ManifestPreservation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Capture manifest
	manifestPath := filepath.Join(env.TargetDir, ".dot-manifest.json")
	content1, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	// Try to manage non-existent package (should fail)
	err = client.Manage(env.Context(), "nonexistent")
	assert.Error(t, err)

	// Verify manifest unchanged
	content2, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, string(content1), string(content2))
}

// TestState_IncrementalDetection_Unchanged tests unchanged package detection.
func TestState_IncrementalDetection_Unchanged(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Plan remanage (should detect no changes)
	plan, err := client.PlanRemanage(env.Context(), "vim")
	require.NoError(t, err)

	// Should have minimal operations (hash check but no actual changes)
	assert.NotNil(t, plan)
}

// TestState_IncrementalDetection_Modified tests modified package detection.
func TestState_IncrementalDetection_Modified(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	vimPackage := filepath.Join(env.PackageDir, "vim")
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Modify package content
	env.FixtureBuilder().FileTree(vimPackage).
		File("dot-vimrc", "set nocompatible\nset number")

	// Remanage should detect change
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link still works
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestState_IncrementalDetection_AddedFile tests detection of added files.
func TestState_IncrementalDetection_AddedFile(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	vimPackage := filepath.Join(env.PackageDir, "vim")
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Add new file to package
	env.FixtureBuilder().FileTree(vimPackage).
		File("dot-vim-colors", "colorscheme default")

	// Remanage
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify new link created
	colorsLink := filepath.Join(env.TargetDir, ".vim-colors")
	testutil.AssertLinkContains(t, colorsLink, "dot-vim-colors")
}

// TestState_IncrementalDetection_DeletedFile tests detection of deleted files.
func TestState_IncrementalDetection_DeletedFile(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with multiple files
	vimPackage := filepath.Join(env.PackageDir, "vim")
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		WithFile("dot-vim-colors", "colorscheme default").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify both links exist
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vim-colors"), "dot-vim-colors")

	// Remove one file from package
	colorsPath := filepath.Join(vimPackage, "dot-vim-colors")
	require.NoError(t, os.Remove(colorsPath))

	// Remanage
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify removed file's link is gone
	testutil.AssertNotExists(t, filepath.Join(env.TargetDir, ".vim-colors"))

	// Verify other link still exists
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
}

// TestState_MultiplePackages tests state with multiple packages.
func TestState_MultiplePackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create multiple packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	env.FixtureBuilder().Package("git").
		WithFile("dot-gitconfig", "[user]\nname = Test").
		Create()

	// Manage all
	err := client.Manage(env.Context(), "vim", "zsh", "git")
	require.NoError(t, err)

	// Verify manifest tracks all packages
	status, err := client.Status(env.Context())
	require.NoError(t, err)

	assert.Len(t, status.Packages, 3)
}

// TestState_ManifestConsistency tests manifest consistency after operations.
func TestState_ManifestConsistency(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		WithFile("dot-vim-colors", "colorscheme default").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Get status
	status, err := client.Status(env.Context(), "vim")
	require.NoError(t, err)

	// Verify manifest matches reality
	assert.Len(t, status.Packages, 1)
	vimPkg := status.Packages[0]
	assert.Equal(t, "vim", vimPkg.Name)
	assert.Equal(t, 2, len(vimPkg.Links))
}
