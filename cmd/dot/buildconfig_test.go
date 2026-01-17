package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestFlags sets up cliFlags and cliContext for a test and returns cleanup function.
func setupTestFlags(t *testing.T, flags CLIFlags) {
	t.Helper()

	previousFlags := cliFlags
	previousCtx := cliContext

	cliFlags = flags
	cliContext = WithCLIFlags(context.Background(), &cliFlags)

	t.Cleanup(func() {
		cliFlags = previousFlags
		cliContext = previousCtx
	})
}

func TestBuildConfig_UsesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")

	// Create config file with custom values
	configContent := `directories:
  package: /custom/packages
  target: /custom/target
  manifest: /custom/manifest
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

	err := os.Setenv("DOT_CONFIG", tmpConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Unsetenv("DOT_CONFIG")
	})

	// Set flags to defaults (should use config values)
	setupTestFlags(t, CLIFlags{
		packageDir: ".",
		targetDir:  "",
	})

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should use config file values
	assert.Contains(t, cfg.PackageDir, "/custom/packages")
	assert.Contains(t, cfg.TargetDir, "/custom/target")
	assert.Equal(t, "/custom/manifest", cfg.ManifestDir)
}

func TestBuildConfig_FlagsOverrideConfig(t *testing.T) {
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")
	flagPkgDir := tmpDir + "/flag-packages"
	flagTargetDir := tmpDir + "/flag-target"

	// Create config file
	configContent := `directories:
  package: /config/packages
  target: /config/target
`
	require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

	err := os.Setenv("DOT_CONFIG", tmpConfig)
	if err != nil {
		t.Fatalf("os.Setenv DOT_CONFIG=%s: %v", tmpConfig, err)
	}
	t.Cleanup(func() {
		os.Unsetenv("DOT_CONFIG")
	})

	// Set flags explicitly (not defaults)
	setupTestFlags(t, CLIFlags{
		packageDir: flagPkgDir,
		targetDir:  flagTargetDir,
	})

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should use flag values, not config
	assert.Contains(t, cfg.PackageDir, "flag-packages")
	assert.Contains(t, cfg.TargetDir, "flag-target")
}

func TestBuildConfig_AppliesDefaults(t *testing.T) {
	tmpConfig := filepath.Join(t.TempDir(), "nonexistent.yaml")

	err := os.Setenv("DOT_CONFIG", tmpConfig)
	if err != nil {
		t.Fatalf("os.Setenv DOT_CONFIG=%s: %v", tmpConfig, err)
	}
	t.Cleanup(func() {
		os.Unsetenv("DOT_CONFIG")
	})

	setupTestFlags(t, CLIFlags{
		packageDir: ".",
		targetDir:  "",
	})

	cfg, err := buildConfig()
	require.NoError(t, err)

	// Should have defaults applied
	assert.NotEmpty(t, cfg.PackageDir)
	assert.NotEmpty(t, cfg.TargetDir)
	assert.NotNil(t, cfg.FS)
	assert.NotNil(t, cfg.Logger)
}

func TestBuildConfig_BackupDirFlag(t *testing.T) {
	tmpBackup := t.TempDir() + "/backups"
	setupTestFlags(t, CLIFlags{
		packageDir: ".",
		targetDir:  t.TempDir(),
		backupDir:  tmpBackup,
	})

	cfg, err := buildConfig()
	require.NoError(t, err)

	assert.Contains(t, cfg.BackupDir, "backups")
}
