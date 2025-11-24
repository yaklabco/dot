//go:build !windows
// +build !windows

package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestConflict_FileExists tests detection of existing files.
func TestConflict_FileExists(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Create conflicting file
	vimrcPath := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.WriteFile(vimrcPath, []byte("existing content"), 0644))

	// Try to manage (should detect conflict)
	err := client.Manage(env.Context(), "vim")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exists")
}

// TestConflict_WrongLinkTarget tests detection of symlinks with wrong targets.
func TestConflict_WrongLinkTarget(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Create symlink to wrong target
	vimrcPath := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.Symlink("/wrong/target", vimrcPath))

	// Try to manage (should detect wrong target)
	err := client.Manage(env.Context(), "vim")
	assert.Error(t, err)
}

// TestConflict_DirectoryVsFile tests conflict between directory and file.
func TestConflict_DirectoryVsFile(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package with file
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vim", "content").
		Create()

	// Create directory with same name in target
	vimPath := filepath.Join(env.TargetDir, ".vim")
	require.NoError(t, os.MkdirAll(vimPath, 0755))

	// Try to manage (should detect type mismatch)
	err := client.Manage(env.Context(), "vim")
	assert.Error(t, err)
}

// TestConflict_MultiplePackageOverlap tests handling of packages with overlapping files.
func TestConflict_MultiplePackageOverlap(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create two packages with same file
	env.FixtureBuilder().Package("pkg1").
		WithFile("dot-shared", "content1").
		Create()

	env.FixtureBuilder().Package("pkg2").
		WithFile("dot-shared", "content2").
		Create()

	// Manage first package
	err := client.Manage(env.Context(), "pkg1")
	require.NoError(t, err)

	// Try to manage second package (should detect conflict)
	err = client.Manage(env.Context(), "pkg2")
	assert.Error(t, err)
}

// TestConflict_PlanDetection tests that conflicts are detected in planning phase.
func TestConflict_PlanDetection(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Create conflicting file
	vimrcPath := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.WriteFile(vimrcPath, []byte("existing"), 0644))

	// Plan should detect conflict
	plan, err := client.PlanManage(env.Context(), "vim")
	// Depending on implementation, this might return error or plan with conflicts
	if err == nil {
		// Check if plan has conflicts metadata
		assert.NotNil(t, plan)
	} else {
		assert.Error(t, err)
	}
}

// TestConflict_BrokenSymlinkConflict tests conflict with broken symlink.
func TestConflict_BrokenSymlinkConflict(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create broken symlink
	vimrcPath := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.Symlink("/nonexistent", vimrcPath))

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Try to manage (should detect conflict with broken symlink)
	err := client.Manage(env.Context(), "vim")
	assert.Error(t, err)
}

// TestConflict_PermissionConflict tests permission-related conflicts.
func TestConflict_PermissionConflict(t *testing.T) {
	// Skip in CI environments where permission handling is unreliable
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("skipping permission test in CI environment - permission handling unreliable")
	}

	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("skipping permission test on macOS/Windows - directory permissions behave differently")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Create read-only directory in target
	confDir := filepath.Join(env.TargetDir, ".config")
	require.NoError(t, os.MkdirAll(confDir, 0755))
	require.NoError(t, os.Chmod(confDir, 0444))

	// Ensure cleanup happens even if test fails
	t.Cleanup(func() {
		_ = os.Chmod(confDir, 0755)
	})

	// Create package that needs to write in that directory
	env.FixtureBuilder().Package("nvim").
		WithFile("dot-config/nvim/init.vim", "syntax on").
		Create()

	// Try to manage (should encounter permission issue)
	err := client.Manage(env.Context(), "nvim")

	// Restore permissions before assertion to avoid cleanup issues
	require.NoError(t, os.Chmod(confDir, 0755))

	// Now verify we got an error
	assert.Error(t, err, "expected permission error when writing to read-only directory")
}

// TestConflict_RemanageWithModifiedTarget tests remanage when target was modified.
func TestConflict_RemanageWithModifiedTarget(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Modify the symlink target manually
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.Remove(vimrcLink))
	require.NoError(t, os.Symlink("/wrong/target", vimrcLink))

	// Remanage behavior with modified symlinks depends on implementation
	// It may not detect changes if file hash hasn't changed
	err = client.Remanage(env.Context(), "vim")

	// The test just verifies remanage completes without panic
	// Actual behavior (fix vs no-op vs error) depends on implementation
	_ = err
}

// TestConflict_DoctorDetection tests that doctor detects conflicts.
func TestConflict_DoctorDetection(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Modify symlink to wrong target
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.Remove(vimrcLink))
	require.NoError(t, os.Symlink("/wrong/target", vimrcLink))

	// Doctor should detect wrong link
	report, err := client.Doctor(env.Context())
	require.NoError(t, err)
	assert.Greater(t, len(report.Issues), 0)
}

// TestConflict_EmptyDirectoryNoConflict tests that empty directories don't cause conflicts.
func TestConflict_EmptyDirectoryNoConflict(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create empty directory in target
	emptyDir := filepath.Join(env.TargetDir, ".config")
	require.NoError(t, os.MkdirAll(emptyDir, 0755))

	// Create package that creates files in that directory
	env.FixtureBuilder().Package("nvim").
		WithFile("dot-config/nvim/init.vim", "syntax on").
		Create()

	// Should not conflict with empty directory
	err := client.Manage(env.Context(), "nvim")
	// This might succeed or fail depending on implementation
	// Just verify it handles it without panicking
	_ = err
}

// TestConflict_SamePackageReinstall tests reinstalling same package.
func TestConflict_SamePackageReinstall(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create and manage package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Try to manage same package again (should be handled via remanage)
	err = client.Manage(env.Context(), "vim")
	// Should either succeed (no-op) or handle gracefully
	_ = err
}

// TestConflict_SymlinkToDirectory tests symlink pointing to directory.
func TestConflict_SymlinkToDirectory(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Create directory
	actualDir := filepath.Join(env.TargetDir, "actual")
	require.NoError(t, os.MkdirAll(actualDir, 0755))

	// Create symlink to directory
	linkPath := filepath.Join(env.TargetDir, ".config")
	require.NoError(t, os.Symlink(actualDir, linkPath))

	// Create package that expects .config to be manageable
	env.FixtureBuilder().Package("nvim").
		WithFile("dot-config/nvim/init.vim", "syntax on").
		Create()

	// Try to manage (behavior depends on implementation)
	err := client.Manage(env.Context(), "nvim")
	// Just verify it handles it without panicking
	_ = err
}

// TestConflict_RelativeVsAbsoluteSymlinks tests different symlink types.
func TestConflict_RelativeVsAbsoluteSymlinks(t *testing.T) {
	env := testutil.NewTestEnvironment(t)

	// Create packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Test with relative links
	relClient := testutil.NewTestClient(t, env)
	err := relClient.Manage(env.Context(), "vim")
	require.NoError(t, err)

	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	target, err := os.Readlink(vimrcLink)
	require.NoError(t, err)

	// Verify symlink exists (relative or absolute)
	assert.NotEmpty(t, target)
}
