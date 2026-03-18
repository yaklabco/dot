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

func TestBuildConfig_PackageNameMappingFromConfig(t *testing.T) {
	t.Run("reads package_name_mapping=false from config", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpConfig := filepath.Join(tmpDir, "config.yaml")

		configContent := `dotfile:
  translate: true
  prefix: "dot-"
  package_name_mapping: false
`
		require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

		os.Setenv("DOT_CONFIG", tmpConfig)
		t.Cleanup(func() { os.Unsetenv("DOT_CONFIG") })

		setupTestFlags(t, CLIFlags{
			packageDir: ".",
			targetDir:  tmpDir,
		})

		cfg, err := buildConfig()
		require.NoError(t, err)
		assert.False(t, cfg.PackageNameMapping, "should read package_name_mapping=false from config")
	})

	t.Run("defaults to true when config has no package_name_mapping", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpConfig := filepath.Join(tmpDir, "config.yaml")

		configContent := `dotfile:
  translate: true
  prefix: "dot-"
`
		require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

		os.Setenv("DOT_CONFIG", tmpConfig)
		t.Cleanup(func() { os.Unsetenv("DOT_CONFIG") })

		setupTestFlags(t, CLIFlags{
			packageDir: ".",
			targetDir:  tmpDir,
		})

		cfg, err := buildConfig()
		require.NoError(t, err)
		assert.True(t, cfg.PackageNameMapping, "should default to true when not set in config")
	})
}

func TestBuildConfig_TranslateFromConfig(t *testing.T) {
	t.Run("reads translate=false from config", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpConfig := filepath.Join(tmpDir, "config.yaml")

		configContent := `dotfile:
  translate: false
  prefix: "dot-"
  package_name_mapping: true
`
		require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

		t.Setenv("DOT_CONFIG", tmpConfig)

		setupTestFlags(t, CLIFlags{
			packageDir: ".",
			targetDir:  tmpDir,
		})

		cfg, err := buildConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg.Translate, "Translate should be set from config")
		assert.False(t, *cfg.Translate, "should read translate=false from config")
	})

	t.Run("reads translate=true from config", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpConfig := filepath.Join(tmpDir, "config.yaml")

		configContent := `dotfile:
  translate: true
  prefix: "dot-"
`
		require.NoError(t, os.WriteFile(tmpConfig, []byte(configContent), 0644))

		t.Setenv("DOT_CONFIG", tmpConfig)

		setupTestFlags(t, CLIFlags{
			packageDir: ".",
			targetDir:  tmpDir,
		})

		cfg, err := buildConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg.Translate, "Translate should be set from config")
		assert.True(t, *cfg.Translate, "should read translate=true from config")
	})

	t.Run("defaults to true when config file missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("DOT_CONFIG", filepath.Join(tmpDir, "nonexistent.yaml"))

		setupTestFlags(t, CLIFlags{
			packageDir: ".",
			targetDir:  tmpDir,
		})

		cfg, err := buildConfig()
		require.NoError(t, err)
		// When config file doesn't exist, default ExtendedConfig is used with Translate=true
		require.NotNil(t, cfg.Translate, "Translate should be set even without config file")
		assert.True(t, *cfg.Translate, "should default to true")
	})
}
