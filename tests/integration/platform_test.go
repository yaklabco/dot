package integration

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestPlatform_PathSeparators tests handling of platform-specific path separators.
func TestPlatform_PathSeparators(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with nested structure using filepath.Join for portability
	pkgPath := filepath.Join("nested", "deep")
	env.FixtureBuilder().Package(pkgPath).
		WithFile("dot-file", "content").
		Create()

	// Manage should handle platform-specific separators
	err := client.Manage(env.Context(), pkgPath)
	require.NoError(t, err)

	// Verify link created
	linkPath := filepath.Join(env.TargetDir, ".file")
	testutil.AssertLinkContains(t, linkPath, "dot-file")
}

// TestPlatform_SymlinkSupport tests symlink functionality on current platform.
func TestPlatform_SymlinkSupport(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink support on Windows requires special permissions")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create simple package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Manage package
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify symlink created successfully
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestPlatform_CaseSensitivity tests platform-specific case handling.
func TestPlatform_CaseSensitivity(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("test").
		WithFile("dot-File", "content").
		Create()

	err := client.Manage(env.Context(), "test")
	require.NoError(t, err)

	// On case-insensitive filesystems, ".file" and ".File" are the same
	// On case-sensitive filesystems, they're different
	linkPath := filepath.Join(env.TargetDir, ".File")
	testutil.AssertLinkContains(t, linkPath, "dot-File")
}

// TestPlatform_AbsolutePaths tests handling of absolute paths.
func TestPlatform_AbsolutePaths(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Verify paths are absolute
	assert.True(t, filepath.IsAbs(env.PackageDir))
	assert.True(t, filepath.IsAbs(env.TargetDir))

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Verify symlink target is correct
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestPlatform_SpecialCharactersInPaths tests handling of special characters.
func TestPlatform_SpecialCharactersInPaths(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with special characters (avoiding OS-specific forbidden chars)
	env.FixtureBuilder().Package("test-pkg").
		WithFile("dot-file_name", "content").
		Create()

	err := client.Manage(env.Context(), "test-pkg")
	require.NoError(t, err)

	linkPath := filepath.Join(env.TargetDir, ".file_name")
	testutil.AssertLinkContains(t, linkPath, "dot-file_name")
}

// TestPlatform_LongPaths tests handling of long pathnames.
func TestPlatform_LongPaths(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with reasonably long path
	longName := "very-long-package-name-that-is-still-reasonable"
	env.FixtureBuilder().Package(longName).
		WithFile("dot-file", "content").
		Create()

	err := client.Manage(env.Context(), longName)
	require.NoError(t, err)
}

// TestPlatform_CurrentOS tests on the current operating system.
func TestPlatform_CurrentOS(t *testing.T) {
	t.Logf("Running on OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Should work on current OS
	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)
}

// TestPlatform_TempDirectoryHandling tests temp directory usage.
func TestPlatform_TempDirectoryHandling(t *testing.T) {
	env := testutil.NewTestEnvironment(t)

	// Verify temp directories are created properly
	assert.DirExists(t, env.PackageDir)
	assert.DirExists(t, env.TargetDir)

	// Temp dirs should be cleaned up automatically by t.TempDir()
}

// TestPlatform_FilePermissions tests file permission handling.
func TestPlatform_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions work differently on Windows")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Symlinks should be created successfully regardless of permissions
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestPlatform_CleanupBehavior tests cleanup of temporary files.
func TestPlatform_CleanupBehavior(t *testing.T) {
	env := testutil.NewTestEnvironment(t)

	// Add custom cleanup
	cleaned := false
	env.AddCleanup(func() {
		cleaned = true
	})

	// Trigger cleanup
	env.Cleanup()

	assert.True(t, cleaned)
}
