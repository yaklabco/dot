package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/internal/config"
)

func TestConfigUpgrade_WithOldConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create an old-style config with some custom values
	oldConfig := `directories:
  package: "/custom/dotfiles"
  target: "~"
  manifest: "."

logging:
  level: "DEBUG"
  format: "json"

symlinks:
  mode: "absolute"
  folding: true
  overwrite: true

ignore:
  use_defaults: true
  patterns:
    - "*.swp"
    - ".git"
  overrides:
    - "important.conf"
    - "secret.key"
`

	require.NoError(t, os.WriteFile(configPath, []byte(oldConfig), 0600))

	// Run upgrade
	backupPath, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)
	require.NotEmpty(t, backupPath)

	// Verify backup was created
	assert.FileExists(t, backupPath)

	// Load upgraded config
	loader := config.NewLoader("dot", configPath)
	upgraded, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// Verify user values were preserved
	assert.Equal(t, "/custom/dotfiles", upgraded.Directories.Package)
	assert.Equal(t, "~", upgraded.Directories.Target)
	assert.Equal(t, "DEBUG", upgraded.Logging.Level)
	assert.Equal(t, "json", upgraded.Logging.Format)
	assert.Equal(t, "absolute", upgraded.Symlinks.Mode)
	assert.True(t, upgraded.Symlinks.Folding)
	assert.True(t, upgraded.Symlinks.Overwrite)

	// Verify deprecated overrides were migrated
	assert.Contains(t, upgraded.Ignore.Patterns, "!important.conf")
	assert.Contains(t, upgraded.Ignore.Patterns, "!secret.key")

	// Verify existing patterns were preserved
	assert.Contains(t, upgraded.Ignore.Patterns, "*.swp")
	assert.Contains(t, upgraded.Ignore.Patterns, ".git")

	// Verify new fields have defaults
	assert.True(t, upgraded.Ignore.PerPackageIgnore, "new field should have default value")
	assert.Equal(t, int64(0), upgraded.Ignore.MaxFileSize, "new field should have default value")

	// Verify config file contains header
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "# Dot Configuration")
	assert.Contains(t, contentStr, "# Upgraded on")
	assert.Contains(t, contentStr, "# Backup saved to:")
	assert.Contains(t, contentStr, "# Deprecated fields migrated:")
}

func TestConfigUpgrade_NoConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")

	_, err := config.UpgradeConfig(configPath, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config file does not exist")
	assert.Contains(t, err.Error(), "dot config init")
}

func TestConfigUpgrade_CreatesBackup(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/packages"
	writer := config.NewWriter(configPath)
	require.NoError(t, writer.Write(cfg, config.WriteOptions{Format: "yaml"}))

	// Store original content
	originalContent, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Run upgrade
	backupPath, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)

	// Verify backup exists and matches original
	assert.FileExists(t, backupPath)
	backupContent, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, backupContent)

	// Verify backup filename format
	backupName := filepath.Base(backupPath)
	assert.Regexp(t, `^\d{8}-\d{6}-config\.bak$`, backupName)
}

func TestConfigUpgrade_PreservesCustomizations(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with extensive customizations
	customConfig := config.DefaultExtended()
	customConfig.Directories.Package = "/my/dotfiles"
	customConfig.Directories.Target = "/my/home"
	customConfig.Directories.Manifest = ".config/dot"
	customConfig.Logging.Level = "ERROR"
	customConfig.Logging.Format = "text"
	customConfig.Symlinks.Mode = "relative"
	customConfig.Symlinks.Folding = false
	customConfig.Symlinks.Overwrite = true
	customConfig.Symlinks.Backup = true
	customConfig.Symlinks.BackupSuffix = ".backup"
	customConfig.Ignore.UseDefaults = false
	customConfig.Ignore.Patterns = []string{"custom-*.pattern", "test/"}
	customConfig.Output.Verbosity = 3
	customConfig.Output.Color = "always"
	customConfig.Operations.MaxParallel = 8
	customConfig.Dotfile.Translate = false

	writer := config.NewWriter(configPath)
	require.NoError(t, writer.Write(customConfig, config.WriteOptions{Format: "yaml"}))

	// Upgrade
	_, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)

	// Load and verify all customizations preserved
	loader := config.NewLoader("dot", configPath)
	upgraded, err := loader.LoadWithEnv()
	require.NoError(t, err)

	assert.Equal(t, "/my/dotfiles", upgraded.Directories.Package)
	assert.Equal(t, "/my/home", upgraded.Directories.Target)
	assert.Equal(t, ".config/dot", upgraded.Directories.Manifest)
	assert.Equal(t, "ERROR", upgraded.Logging.Level)
	assert.Equal(t, "text", upgraded.Logging.Format)
	assert.Equal(t, "relative", upgraded.Symlinks.Mode)
	assert.False(t, upgraded.Symlinks.Folding)
	assert.True(t, upgraded.Symlinks.Overwrite)
	assert.True(t, upgraded.Symlinks.Backup)
	assert.Equal(t, ".backup", upgraded.Symlinks.BackupSuffix)
	assert.False(t, upgraded.Ignore.UseDefaults)
	assert.ElementsMatch(t, []string{"custom-*.pattern", "test/"}, upgraded.Ignore.Patterns)
	assert.Equal(t, 3, upgraded.Output.Verbosity)
	assert.Equal(t, "always", upgraded.Output.Color)
	assert.Equal(t, 8, upgraded.Operations.MaxParallel)
	assert.False(t, upgraded.Dotfile.Translate)
}

func TestConfigUpgrade_MultipleUpgrades(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create initial config
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/v1"
	writer := config.NewWriter(configPath)
	require.NoError(t, writer.Write(cfg, config.WriteOptions{Format: "yaml"}))

	// First upgrade
	backup1, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)
	assert.FileExists(t, backup1)

	// Sleep to ensure different timestamp
	time.Sleep(2 * time.Second)

	// Modify and upgrade again
	loader := config.NewLoader("dot", configPath)
	cfg2, err := loader.LoadWithEnv()
	require.NoError(t, err)
	cfg2.Directories.Package = "/test/v2"
	require.NoError(t, writer.Write(cfg2, config.WriteOptions{Format: "yaml"}))

	backup2, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)
	assert.FileExists(t, backup2)

	// Verify both backups exist and are different
	assert.NotEqual(t, backup1, backup2, "backup paths should be different")

	// Verify final config has latest value
	cfg3, err := loader.LoadWithEnv()
	require.NoError(t, err)
	assert.Equal(t, "/test/v2", cfg3.Directories.Package)
}

func TestConfigUpgrade_EmptySlicesGetDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with empty slices
	minimalConfig := `directories:
  package: "/custom/packages"

ignore:
  use_defaults: false
  patterns: []
`

	require.NoError(t, os.WriteFile(configPath, []byte(minimalConfig), 0600))

	// Upgrade
	_, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)

	// Load upgraded config
	loader := config.NewLoader("dot", configPath)
	upgraded, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// User's custom package directory preserved
	assert.Equal(t, "/custom/packages", upgraded.Directories.Package)

	// User's false value preserved (not replaced with default true)
	assert.False(t, upgraded.Ignore.UseDefaults)

	// Empty patterns should remain empty (user explicitly set to empty)
	assert.Empty(t, upgraded.Ignore.Patterns)
}

func TestConfigUpgrade_HeaderContainsUpgradeInfo(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with overrides
	configWithOverrides := `directories:
  package: "."

ignore:
  use_defaults: true
  overrides:
    - "keep-me.txt"
`

	require.NoError(t, os.WriteFile(configPath, []byte(configWithOverrides), 0600))

	// Upgrade
	backupPath, err := config.UpgradeConfig(configPath, true)
	require.NoError(t, err)

	// Read upgraded file
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Verify header components
	lines := strings.Split(contentStr, "\n")
	var headerLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			headerLines = append(headerLines, line)
		} else if line == "" || line == "---" {
			continue
		} else {
			break // End of header
		}
	}

	header := strings.Join(headerLines, "\n")

	assert.Contains(t, header, "# Dot Configuration")
	assert.Contains(t, header, "# Upgraded on")
	assert.Contains(t, header, "# Backup saved to:")
	assert.Contains(t, header, backupPath)
	assert.Contains(t, header, "# Deprecated fields migrated:")
	assert.Contains(t, header, "ignore.overrides → ignore.patterns")
	assert.Contains(t, header, "# See https://github.com/jamesainslie/dot")
}

func TestRunConfigUpgrade_Integration(t *testing.T) {
	// This is a minimal integration test that verifies the command wiring
	// Full CLI testing would require mocking stdin for the confirmation prompt

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create test config
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/integration"
	writer := config.NewWriter(configPath)
	require.NoError(t, writer.Write(cfg, config.WriteOptions{Format: "yaml"}))

	// Set environment to use our temp config
	originalEnv := os.Getenv("DOT_CONFIG")
	os.Setenv("DOT_CONFIG", configPath)
	defer os.Setenv("DOT_CONFIG", originalEnv)

	// Run upgrade with force flag (to skip prompt)
	cmd := newConfigUpgradeCommand()
	cmd.SetArgs([]string{"--force"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify config was upgraded
	loader := config.NewLoader("dot", configPath)
	upgraded, err := loader.LoadWithEnv()
	require.NoError(t, err)
	assert.Equal(t, "/test/integration", upgraded.Directories.Package)
}
