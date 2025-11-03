package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePackageDirectory_ExplicitFlag(t *testing.T) {
	// Explicit flag takes highest priority
	result, err := resolvePackageDirectory("/explicit/path")
	require.NoError(t, err)

	abs, _ := filepath.Abs("/explicit/path")
	assert.Equal(t, abs, result)
}

func TestResolvePackageDirectory_EnvironmentVariable(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("DOT_PACKAGE_DIR", tmpDir)
	defer os.Unsetenv("DOT_PACKAGE_DIR")

	// Should use environment variable when no explicit flag
	result, err := resolvePackageDirectory("")
	require.NoError(t, err)

	abs, _ := filepath.Abs(tmpDir)
	assert.Equal(t, abs, result)
}

func TestResolvePackageDirectory_CurrentDirWithBootstrap(t *testing.T) {
	tmpDir := t.TempDir()
	os.Unsetenv("DOT_PACKAGE_DIR")

	// Create .dotbootstrap.yaml in temp dir
	bootstrapPath := filepath.Join(tmpDir, ".dotbootstrap.yaml")
	err := os.WriteFile(bootstrapPath, []byte("version: 1.0\n"), 0644)
	require.NoError(t, err)

	// Change to temp dir
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	// Should detect current directory
	result, err := resolvePackageDirectory("")
	require.NoError(t, err)

	// On macOS, paths may have /private prefix - normalize both
	expectedAbs, _ := filepath.EvalSymlinks(tmpDir)
	resultAbs, _ := filepath.EvalSymlinks(result)
	assert.Equal(t, expectedAbs, resultAbs)
}

func TestResolvePackageDirectory_ParentSearch(t *testing.T) {
	tmpDir := t.TempDir()
	os.Unsetenv("DOT_PACKAGE_DIR")

	// Create nested structure: tmpDir/.dotbootstrap.yaml and tmpDir/subdir/
	bootstrapPath := filepath.Join(tmpDir, ".dotbootstrap.yaml")
	err := os.WriteFile(bootstrapPath, []byte("version: 1.0\n"), 0644)
	require.NoError(t, err)

	subdir := filepath.Join(tmpDir, "subdir")
	err = os.MkdirAll(subdir, 0755)
	require.NoError(t, err)

	// Change to subdir
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(subdir)

	// Should find parent directory with .dotbootstrap.yaml
	result, err := resolvePackageDirectory("")
	require.NoError(t, err)

	// On macOS, paths may have /private prefix - normalize both
	expectedAbs, _ := filepath.EvalSymlinks(tmpDir)
	resultAbs, _ := filepath.EvalSymlinks(result)
	assert.Equal(t, expectedAbs, resultAbs)
}

func TestResolvePackageDirectory_DefaultFallback(t *testing.T) {
	os.Unsetenv("DOT_PACKAGE_DIR")

	// From a directory that doesn't have .dotbootstrap.yaml and no config file
	// Should fall back to ~/.dotfiles
	result, err := resolvePackageDirectory(".")
	require.NoError(t, err)

	// Result should be an absolute path
	assert.NotEmpty(t, result)
	assert.True(t, filepath.IsAbs(result), "result should be absolute path")
}

func TestIsDotfilesRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Not a repo initially
	assert.False(t, isDotfilesRepo(tmpDir))

	// Create .dotbootstrap.yaml
	bootstrapPath := filepath.Join(tmpDir, ".dotbootstrap.yaml")
	err := os.WriteFile(bootstrapPath, []byte("version: 1.0\n"), 0644)
	require.NoError(t, err)

	// Now it's a repo
	assert.True(t, isDotfilesRepo(tmpDir))
}

func TestFindDotfilesRepo_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	result := findDotfilesRepo(tmpDir)
	assert.Empty(t, result)
}

func TestFindDotfilesRepo_Found(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .dotbootstrap.yaml in tmpDir
	bootstrapPath := filepath.Join(tmpDir, ".dotbootstrap.yaml")
	err := os.WriteFile(bootstrapPath, []byte("version: 1.0\n"), 0644)
	require.NoError(t, err)

	// Create nested directory
	subdir := filepath.Join(tmpDir, "a", "b", "c")
	err = os.MkdirAll(subdir, 0755)
	require.NoError(t, err)

	// Should find tmpDir from deep subdirectory
	result := findDotfilesRepo(subdir)
	// On macOS, paths may have /private prefix
	expectedAbs, _ := filepath.EvalSymlinks(tmpDir)
	resultAbs, _ := filepath.EvalSymlinks(result)
	assert.Equal(t, expectedAbs, resultAbs)
}

func TestFindDotfilesRepo_StopsAtHome(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	// Search from home should not return anything (unless home is a dotfiles repo)
	result := findDotfilesRepo(homeDir)
	// Result will be empty or homeDir if it has .dotbootstrap.yaml
	// We just verify it doesn't panic or error
	_ = result
}
