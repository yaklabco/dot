package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestScenario_NewUser_FirstTimeSetup tests a new user's first-time setup workflow.
func TestScenario_NewUser_FirstTimeSetup(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// User creates their first dotfiles package
	env.FixtureBuilder().Package("dotfiles").
		WithFile("dot-bashrc", "export PS1='\\u@\\h:\\w\\$ '").
		WithFile("dot-vimrc", "set nocompatible").
		WithFile("dot-gitconfig", "[user]\nname = Test User").
		Create()

	// User manages their dotfiles
	err := client.Manage(env.Context(), "dotfiles")
	require.NoError(t, err)

	// Verify all files linked
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".bashrc"), "dot-bashrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".gitconfig"), "dot-gitconfig")

	// User checks status
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 1)
	assert.Equal(t, 3, status.Packages[0].LinkCount)
}

// TestScenario_DevelopmentWorkflow tests iterative development workflow.
func TestScenario_DevelopmentWorkflow(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// User creates initial package
	vimPackage := filepath.Join(env.PackageDir, "vim")
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Install package
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// User modifies package (adds new file)
	env.FixtureBuilder().FileTree(vimPackage).
		File("dot-vim-colors", "colorscheme desert")

	// User remanages to pick up changes
	err = client.Remanage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify new file is linked
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vim-colors"), "dot-vim-colors")
}

// TestScenario_MultiMachine tests managing dotfiles across multiple machines.
func TestScenario_MultiMachine(t *testing.T) {
	// Simulate first machine
	env1 := testutil.NewTestEnvironment(t)
	client1 := testutil.NewTestClient(t, env1)

	// Create dotfiles with portable configuration
	env1.FixtureBuilder().Package("dotfiles").
		WithFile("dot-bashrc", "export EDITOR=vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client1.Manage(env1.Context(), "dotfiles")
	require.NoError(t, err)

	// Simulate second machine (new environment)
	env2 := testutil.NewTestEnvironment(t)
	client2 := testutil.NewTestClient(t, env2)

	// Copy package structure to second machine
	env2.FixtureBuilder().Package("dotfiles").
		WithFile("dot-bashrc", "export EDITOR=vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Install on second machine
	err = client2.Manage(env2.Context(), "dotfiles")
	require.NoError(t, err)

	// Verify both machines have same setup
	status1, err := client1.Status(env1.Context())
	require.NoError(t, err)

	status2, err := client2.Status(env2.Context())
	require.NoError(t, err)

	assert.Equal(t, len(status1.Packages), len(status2.Packages))
}

// TestScenario_SelectiveInstallation tests installing subset of packages.
func TestScenario_SelectiveInstallation(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create multiple packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("emacs").
		WithFile("dot-emacs", "(setq inhibit-startup-message t)").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	// User only installs vim and zsh (not emacs)
	err := client.Manage(env.Context(), "vim", "zsh")
	require.NoError(t, err)

	// Verify only selected packages installed
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 2)

	packageNames := make(map[string]bool)
	for _, pkg := range status.Packages {
		packageNames[pkg.Name] = true
	}
	assert.True(t, packageNames["vim"])
	assert.True(t, packageNames["zsh"])
	assert.False(t, packageNames["emacs"])
}

// TestScenario_LargeRepository tests managing many packages.
func TestScenario_LargeRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large repository test in short mode")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create 50 packages
	packages := make([]string, 50)
	for i := 0; i < 50; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+(i%26))))
		if i >= 26 {
			pkgName += string(rune('0' + (i / 26)))
		}
		packages[i] = pkgName

		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	// Install all packages
	err := client.Manage(env.Context(), packages...)
	require.NoError(t, err)

	// Verify all installed
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 50)
}

// TestScenario_BackupAndRestore tests backup workflow.
func TestScenario_BackupAndRestore(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Install
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// List installed packages (for backup)
	list, err := client.List(env.Context())
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Uninstall
	err = client.Unmanage(env.Context(), "vim")
	require.NoError(t, err)

	// Restore from backup (reinstall)
	for _, pkg := range list {
		err = client.Manage(env.Context(), pkg.Name)
		require.NoError(t, err)
	}

	// Verify restored
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 1)
}

// TestScenario_HealthCheckWorkflow tests regular health check workflow.
func TestScenario_HealthCheckWorkflow(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Install packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Run health check
	report, err := client.Doctor(env.Context())
	require.NoError(t, err)

	// Verify healthy
	assert.Equal(t, 0, report.Statistics.BrokenLinks)
}

// TestScenario_CleanupWorkflow tests removing unused packages.
func TestScenario_CleanupWorkflow(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Install multiple packages with unique files to avoid conflicts
	for i := 0; i < 5; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+i)))
		fileName := "dot-file" + string(rune('a'+i))
		env.FixtureBuilder().Package(pkgName).
			WithFile(fileName, "content").
			Create()
		err := client.Manage(env.Context(), pkgName)
		require.NoError(t, err)
	}

	// User decides to remove some packages
	err := client.Unmanage(env.Context(), "pkg/a", "pkg/b", "pkg/c")
	require.NoError(t, err)

	// Verify only desired packages remain
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 2)
}

// TestScenario_OrganizedPackages tests organizing dotfiles by category.
func TestScenario_OrganizedPackages(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create organized package structure
	env.FixtureBuilder().Package("editors/vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("editors/emacs").
		WithFile("dot-emacs", "(setq inhibit-startup-message t)").
		Create()

	env.FixtureBuilder().Package("shells/zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	// Install organized packages
	err := client.Manage(env.Context(), "editors/vim", "shells/zsh")
	require.NoError(t, err)

	// Verify installed
	status, err := client.Status(env.Context())
	require.NoError(t, err)
	assert.Len(t, status.Packages, 2)
}

// TestScenario_RapidIteration tests rapid development iteration.
func TestScenario_RapidIteration(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	vimPackage := filepath.Join(env.PackageDir, "vim")

	// Initial install
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Iterate multiple times
	for i := 0; i < 5; i++ {
		// Modify package
		env.FixtureBuilder().FileTree(vimPackage).
			File("dot-vimrc", "set nocompatible\n\" iteration "+string(rune('0'+i)))

		// Remanage
		err = client.Remanage(env.Context(), "vim")
		require.NoError(t, err)
	}

	// Verify final state
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
}
