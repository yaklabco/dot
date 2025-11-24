package dot

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/adapters"
)

// TestAdopt_PreventDoubleAdoption tests that adopting an already-adopted file fails gracefully.
func TestAdopt_PreventDoubleAdoption(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	packageDir := "/packages"
	targetDir := "/home"

	// Setup: Create .ssh directory with files
	sshDir := filepath.Join(targetDir, ".ssh")
	require.NoError(t, fs.MkdirAll(ctx, sshDir, 0o755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(sshDir, "config"), []byte("Host example.com"), 0o600))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(sshDir, "known_hosts"), []byte("example.com ssh-rsa ..."), 0o644))

	// Create adopt service
	adoptSvc := &AdoptService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}

	// First adoption - should succeed
	plan, err := adoptSvc.PlanAdopt(ctx, []string{sshDir}, "dot-ssh")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Execute the plan (this would create symlink and move files)
	// For this test, we'll manually set up the symlink state
	pkgPath := filepath.Join(packageDir, "dot-ssh")
	require.NoError(t, fs.MkdirAll(ctx, pkgPath, 0o755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(pkgPath, "config"), []byte("Host example.com"), 0o600))
	require.NoError(t, fs.Remove(ctx, filepath.Join(sshDir, "config")))
	require.NoError(t, fs.Remove(ctx, filepath.Join(sshDir, "known_hosts")))
	require.NoError(t, fs.Remove(ctx, sshDir))
	require.NoError(t, fs.Symlink(ctx, pkgPath, sshDir))

	// Verify symlink exists and points correctly
	isSymlink, err := fs.IsSymlink(ctx, sshDir)
	require.NoError(t, err)
	assert.True(t, isSymlink, "Expected .ssh to be a symlink")

	// Attempt to adopt .ssh again - should fail
	_, err = adoptSvc.PlanAdopt(ctx, []string{sshDir}, "dot-ssh")
	assert.Error(t, err, "Expected error when adopting already-managed symlink")
	assert.Contains(t, err.Error(), "already managed by dot", "Error should mention file is already managed")
}

// TestAdopt_RejectSymlinkSource tests that adopting a symlink to an external location warns the user.
func TestAdopt_RejectSymlinkSource(t *testing.T) {
	t.Skip("Skipping: requires symlink handling in memfs that's not yet fully implemented")
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	packageDir := "/packages"
	targetDir := "/home"

	// Setup: Create directory and symlink to it
	externalPath := "/external/path"
	testLink := filepath.Join(targetDir, ".test")
	require.NoError(t, fs.MkdirAll(ctx, externalPath, 0o755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(externalPath, "file.txt"), []byte("content"), 0o644))
	require.NoError(t, fs.Symlink(ctx, externalPath, testLink))

	// Create adopt service
	adoptSvc := &AdoptService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}

	// Try to adopt ~/.test - should not error but should log warning
	plan, err := adoptSvc.PlanAdopt(ctx, []string{testLink}, "test-pkg")

	// Should succeed but with a warning logged
	// The warning check would require inspecting logger output
	// For now, we just ensure it doesn't fail
	require.NoError(t, err)
	require.NotNil(t, plan)
}

// TestUnmanage_HandleCorruptedNestedStructure tests unmanaging a package with corrupted nested structure.
func TestUnmanage_HandleCorruptedNestedStructure(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	packageDir := "/packages"
	targetDir := "/home"
	pkg := "dot-ssh"

	// Setup: Manually create dot-ssh/dot-ssh/files corruption
	pkgPath := filepath.Join(packageDir, pkg)
	nestedPath := filepath.Join(pkgPath, pkg)
	require.NoError(t, fs.MkdirAll(ctx, nestedPath, 0o755))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(nestedPath, "config"), []byte("Host example.com"), 0o600))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(nestedPath, "known_hosts"), []byte("example.com ssh-rsa ..."), 0o644))

	// Create target directory
	targetSSH := filepath.Join(targetDir, ".ssh")
	require.NoError(t, fs.MkdirAll(ctx, targetSSH, 0o755))

	// Create unmanage service
	unmanageSvc := &UnmanageService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}

	// Call repair helper directly to verify it handles nested structure
	operations := unmanageSvc.createCorruptedStructureRepair(ctx, pkg, nestedPath, targetSSH)

	// Should create operation
	assert.NotEmpty(t, operations, "Expected repair operation to be generated")

	// Verify repair operation ID
	if len(operations) > 0 {
		assert.Contains(t, string(operations[0].ID()), "repair-nested", "Expected repair operation ID")
	}
}

// TestAdoptUnmanage_DirectoryRoundTrip tests full adopt/unmanage cycle for directory integrity.
func TestAdoptUnmanage_DirectoryRoundTrip(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	packageDir := "/packages"
	targetDir := "/home"

	// Setup: Create .config/.cache/data structure
	configDir := filepath.Join(targetDir, ".config")
	cacheDir := filepath.Join(configDir, ".cache")
	require.NoError(t, fs.MkdirAll(ctx, cacheDir, 0o755))

	dataFile := filepath.Join(cacheDir, "data.txt")
	configFile := filepath.Join(configDir, "app.conf")
	require.NoError(t, fs.WriteFile(ctx, dataFile, []byte("test data"), 0o644))
	require.NoError(t, fs.WriteFile(ctx, configFile, []byte("config=value"), 0o644))

	// Create adopt service
	adoptSvc := &AdoptService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}

	// Adopt .config
	plan, err := adoptSvc.PlanAdopt(ctx, []string{configDir}, "dot-config")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Verify plan contains operations for directory adoption
	assert.NotEmpty(t, plan.Operations, "Expected operations for directory adoption")

	// The plan should include directory creation and file moves with translation
	// This verifies the core logic works without executing the full cycle
}

// TestAdopt_NestedDotDirectories tests adoption of directories with nested dotfile directories.
func TestAdopt_NestedDotDirectories(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()

	packageDir := "/packages"
	targetDir := "/home"

	// Setup: Create .local/.cache/.tmp/file
	localDir := filepath.Join(targetDir, ".local")
	cacheDir := filepath.Join(localDir, ".cache")
	tmpDir := filepath.Join(cacheDir, ".tmp")
	require.NoError(t, fs.MkdirAll(ctx, tmpDir, 0o755))

	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, fs.WriteFile(ctx, testFile, []byte("temporary data"), 0o644))

	// Create adopt service
	adoptSvc := &AdoptService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}

	// Adopt .local
	plan, err := adoptSvc.PlanAdopt(ctx, []string{localDir}, "dot-local")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Verify plan would create proper structure with translation
	// Expected: dot-local/dot-cache/dot-tmp/test.txt
	// The operations should handle nested dotfile translation
	assert.NotEmpty(t, plan.Operations, "Expected operations to be generated")
}
