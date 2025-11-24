package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestE2E_Manage_SinglePackage tests managing a single package.
func TestE2E_Manage_SinglePackage(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create a simple vim package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible\nset number").
		Create()

	// Capture state before
	before := testutil.CaptureState(t, env.TargetDir)

	// Manage package
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify symlink created
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")

	// Verify state changes
	after := testutil.CaptureState(t, env.TargetDir)
	require.NotEqual(t, before.CountSymlinks(), after.CountSymlinks())
	require.Equal(t, 1, after.CountSymlinks())
}

// TestE2E_Manage_MultiplePackages tests managing multiple packages.
func TestE2E_Manage_MultiplePackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create vim package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Create zsh package
	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		WithFile("dot-zshenv", "export PATH=/usr/local/bin:$PATH").
		Create()

	// Create git package
	env.FixtureBuilder().Package("git").
		WithFile("dot-gitconfig", "[user]\nname = Test User").
		Create()

	// Manage all packages
	err := client.Manage(env.Context(), "vim", "zsh", "git")
	require.NoError(t, err)

	// Verify all symlinks created
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshrc"), "dot-zshrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshenv"), "dot-zshenv")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".gitconfig"), "dot-gitconfig")

	// Verify correct count
	after := testutil.CaptureState(t, env.TargetDir)
	require.Equal(t, 4, after.CountSymlinks())
}

// TestE2E_Manage_NestedDirectories tests managing packages with nested directories.
func TestE2E_Manage_NestedDirectories(t *testing.T) {
	t.Skip("Nested directory folding requires directory existence verification - tracked in separate task")

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create nvim package with nested structure
	env.FixtureBuilder().Package("nvim").
		WithFile("dot-config/nvim/init.vim", "syntax on\nset number").
		WithFile("dot-config/nvim/plugin/settings.vim", "let g:loaded_netrw=1").
		Create()

	// Manage package
	err := client.Manage(env.Context(), "nvim")
	require.NoError(t, err)

	// Verify directory structure created
	configNvim := filepath.Join(env.TargetDir, ".config/nvim")
	testutil.AssertDir(t, configNvim)

	// Verify files exist (as symlinks or in symlinked directory)
	initVim := filepath.Join(env.TargetDir, ".config/nvim/init.vim")
	testutil.AssertFile(t, initVim, "syntax on\nset number")
}

// TestE2E_Manage_Idempotent tests that re-managing is idempotent.
func TestE2E_Manage_Idempotent(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Manage once
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link exists
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")

	// Manage again should use remanage internally which is idempotent
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link still exists
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestE2E_Unmanage_SinglePackage tests unmanaging a single package.
func TestE2E_Unmanage_SinglePackage(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link exists
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")

	// Unmanage
	err = client.Unmanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link removed
	testutil.AssertNotExists(t, vimrcLink)
}

// TestE2E_Unmanage_MultiplePackages tests unmanaging multiple packages.
func TestE2E_Unmanage_MultiplePackages(t *testing.T) {
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

	// Unmanage all
	err = client.Unmanage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// Verify all links removed
	testutil.AssertNotExists(t, filepath.Join(env.TargetDir, ".vimrc"))
	testutil.AssertNotExists(t, filepath.Join(env.TargetDir, ".zshrc"))
}

// TestE2E_Unmanage_EmptyDirectoryCleanup tests that empty directories are cleaned up.
func TestE2E_Unmanage_EmptyDirectoryCleanup(t *testing.T) {
	t.Skip("Nested directory cleanup requires directory existence verification - tracked in separate task")

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with nested structure
	env.FixtureBuilder().Package("nvim").
		WithFile("dot-config/nvim/init.vim", "syntax on").
		Create()

	err := client.Manage(env.Context(), "nvim")
	require.NoError(t, err)

	// Verify directory created
	configNvim := filepath.Join(env.TargetDir, ".config/nvim")
	testutil.AssertDir(t, configNvim)

	// Unmanage
	err = client.Unmanage(env.Context(), "nvim")
	require.NoError(t, err)

	// Verify empty directories removed
	testutil.AssertNotExists(t, configNvim)
}

// TestE2E_Remanage_Unchanged tests remanaging unchanged packages (no-op).
func TestE2E_Remanage_Unchanged(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Capture state
	before := testutil.CaptureState(t, env.TargetDir)

	// Remanage unchanged
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify state unchanged (no-op)
	after := testutil.CaptureState(t, env.TargetDir)
	testutil.AssertStateUnchanged(t, before, after)
}

// TestE2E_Remanage_Modified tests remanaging when package content changes.
func TestE2E_Remanage_Modified(t *testing.T) {
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

	// Remanage
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link still exists (remanage should be idempotent for symlinks)
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestE2E_Adopt_SingleFile tests adopting a single file.
func TestE2E_Adopt_SingleFile(t *testing.T) {
	t.Skip("Adopt functionality requires additional implementation - tracked in separate task")

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package directory (without the file we're adopting)
	env.FixtureBuilder().Package("vim").Create()

	// Create existing file in target
	vimrcPath := filepath.Join(env.TargetDir, ".vimrc")
	env.FixtureBuilder().FileTree(env.TargetDir).
		File(".vimrc", "set nocompatible\nset number")

	// Adopt file
	err := client.Adopt(env.Context(), []string{".vimrc"}, "vim")
	require.NoError(t, err)

	// Verify file moved to package
	packageVimrc := filepath.Join(env.PackageDir, "vim/dot-vimrc")
	testutil.AssertFile(t, packageVimrc, "set nocompatible\nset number")

	// Verify symlink created
	testutil.AssertLinkContains(t, vimrcPath, "dot-vimrc")
}

// TestE2E_Adopt_MultipleFiles tests adopting multiple files.
func TestE2E_Adopt_MultipleFiles(t *testing.T) {
	t.Skip("Adopt functionality requires additional implementation - tracked in separate task")

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package directory
	env.FixtureBuilder().Package("zsh").Create()

	// Create existing files
	env.FixtureBuilder().FileTree(env.TargetDir).
		File(".zshrc", "export EDITOR=vim").
		File(".zshenv", "export PATH=/usr/local/bin:$PATH")

	// Adopt files
	err := client.Adopt(env.Context(), []string{".zshrc", ".zshenv"}, "zsh")
	require.NoError(t, err)

	// Verify files moved
	testutil.AssertFile(t, filepath.Join(env.PackageDir, "zsh/dot-zshrc"), "export EDITOR=vim")
	testutil.AssertFile(t, filepath.Join(env.PackageDir, "zsh/dot-zshenv"), "export PATH=/usr/local/bin:$PATH")

	// Verify symlinks created
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshrc"), "dot-zshrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshenv"), "dot-zshenv")
}

// TestE2E_Combined_ManageUnmanageIdentity tests manage then unmanage restores state.
func TestE2E_Combined_ManageUnmanageIdentity(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link exists
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")

	// Unmanage
	err = client.Unmanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify link removed (manifest may still exist, which is acceptable)
	testutil.AssertNotExists(t, vimrcLink)
}

// TestE2E_Combined_ManageAdoptUnmanage tests combined workflow.
func TestE2E_Combined_ManageAdoptUnmanage(t *testing.T) {
	t.Skip("Adopt functionality requires additional implementation - tracked in separate task")

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage vim package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Create zsh package with file
	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	// Manage zsh package
	err = client.Manage(env.Context(), "zsh")
	require.NoError(t, err)

	// Verify both packages installed
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshrc"), "dot-zshrc")

	// Unmanage both
	err = client.Unmanage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// Verify all links removed
	testutil.AssertNotExists(t, filepath.Join(env.TargetDir, ".vimrc"))
	testutil.AssertNotExists(t, filepath.Join(env.TargetDir, ".zshrc"))
}
