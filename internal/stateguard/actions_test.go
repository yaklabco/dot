package stateguard

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/manifest"
)

// writeTestManifest creates a manifest file with the given packages in dir.
func writeTestManifest(t *testing.T, dir string, pkgs map[string]manifest.PackageInfo) {
	t.Helper()
	m := manifest.New()
	for name, pkg := range pkgs {
		pkg.Name = name
		m.AddPackage(pkg)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".dot-manifest.json"), data, 0644))
}

func TestActionContinue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	require.NoError(t, ActionContinue())
	assert.True(t, MarkerExists())
}

func TestActionFresh_RemovesSymlinks(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()

	// Create a real file and a symlink in the target directory
	realFile := filepath.Join(targetDir, "realfile.txt")
	require.NoError(t, os.WriteFile(realFile, []byte("keep me"), 0644))

	symlinkPath := filepath.Join(targetDir, ".bashrc")
	require.NoError(t, os.Symlink("/some/source/.bashrc", symlinkPath))

	// Create manifest referencing both
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc", "realfile.txt"},
			LinkCount: 2,
			TargetDir: targetDir,
		},
	})

	require.NoError(t, ActionFresh(manifestDir, targetDir))

	// Symlink should be removed
	_, err := os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err), "symlink should be removed")

	// Regular file should be preserved
	_, err = os.Stat(realFile)
	assert.NoError(t, err, "regular file should be preserved")

	// Manifest file should be removed
	_, err = os.Stat(filepath.Join(manifestDir, ".dot-manifest.json"))
	assert.True(t, os.IsNotExist(err), "manifest should be removed")

	// Marker should be written
	assert.True(t, MarkerExists())
}

func TestActionFresh_SkipsNonExistentLinks(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc", ".zshrc"},
			LinkCount: 2,
			TargetDir: targetDir,
		},
	})

	// Neither link exists, should not error
	require.NoError(t, ActionFresh(manifestDir, targetDir))
	assert.True(t, MarkerExists())
}

func TestActionBackupAndFresh(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")

	// Create config file
	require.NoError(t, os.WriteFile(configPath, []byte("logging:\n  level: DEBUG\n"), 0644))

	// Create symlink in target
	symlinkPath := filepath.Join(targetDir, ".bashrc")
	require.NoError(t, os.Symlink("/some/source/.bashrc", symlinkPath))

	// Create manifest
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc"},
			LinkCount: 1,
			TargetDir: targetDir,
		},
	})

	homeDir := t.TempDir()
	backupDir, err := ActionBackupAndFresh(manifestDir, targetDir, configPath, homeDir)
	require.NoError(t, err)

	// Backup directory should exist and contain manifest + config
	assert.DirExists(t, backupDir)

	manifestBackup := filepath.Join(backupDir, "manifest", ".dot-manifest.json")
	assert.FileExists(t, manifestBackup)

	configBackup := filepath.Join(backupDir, "config.yaml")
	assert.FileExists(t, configBackup)

	// Original symlink should be removed
	_, err = os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err))

	// Marker should be written
	assert.True(t, MarkerExists())
}

func TestActionBackupAndFresh_CollisionHandling(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	homeDir := t.TempDir()
	targetDir := t.TempDir()
	manifestDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("test"), 0644))

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{},
			LinkCount: 0,
			TargetDir: targetDir,
		},
	})

	// Pre-create the backup dir to force collision
	today := time.Now().Format("20060102")
	existingBackup := filepath.Join(homeDir, ".dot-backup-"+today)
	require.NoError(t, os.MkdirAll(existingBackup, 0755))

	backupDir, err := ActionBackupAndFresh(manifestDir, targetDir, configPath, homeDir)
	require.NoError(t, err)

	// Should have created a suffixed backup directory
	assert.NotEqual(t, existingBackup, backupDir)
	assert.Contains(t, backupDir, ".dot-backup-"+today+"-2")
}

func TestActionFresh_UsesManifestTargetDir(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()

	// Create symlink in actual target dir (not the default)
	customTarget := t.TempDir()
	symlinkPath := filepath.Join(customTarget, ".vimrc")
	require.NoError(t, os.Symlink("/some/source/.vimrc", symlinkPath))

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"vim": {
			Links:     []string{".vimrc"},
			LinkCount: 1,
			TargetDir: customTarget,
		},
	})

	require.NoError(t, ActionFresh(manifestDir, targetDir))

	// Symlink in custom target should be removed
	_, err := os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err))
}

// detectExistingStateHelper loads a manifest and counts packages/links for testing.
func detectExistingStateHelper(t *testing.T, manifestDir, targetDir string) *ExistingState {
	t.Helper()
	ctx := context.Background()
	state, err := DetectExistingState(ctx, manifestDir, targetDir)
	require.NoError(t, err)
	return state
}

func TestDetectExistingState_EmptyManifest(t *testing.T) {
	manifestDir := t.TempDir()
	targetDir := t.TempDir()

	state := detectExistingStateHelper(t, manifestDir, targetDir)
	assert.Nil(t, state, "empty manifest should return nil state")
}

func TestDetectExistingState_WithPackages(t *testing.T) {
	manifestDir := t.TempDir()
	targetDir := t.TempDir()

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc", ".zshrc"},
			LinkCount: 2,
		},
		"vim": {
			Links:     []string{".vimrc"},
			LinkCount: 1,
		},
	})

	state := detectExistingStateHelper(t, manifestDir, targetDir)
	require.NotNil(t, state)
	assert.Equal(t, 2, state.PackageCount)
	assert.Equal(t, 3, state.LinkCount)
	assert.Contains(t, state.ManifestPath, ".dot-manifest.json")
}
