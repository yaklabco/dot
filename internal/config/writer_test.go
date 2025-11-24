package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
	"gopkg.in/yaml.v3"
)

func TestWriter_WriteDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{
		Format:          "yaml",
		IncludeComments: true,
	})
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, configPath)

	// Verify file permissions are secure (0600)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content contains comments
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Dot Configuration File")
	assert.Contains(t, string(content), "# Core Directories")
}

func TestWriter_WriteYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/dotfiles"
	cfg.Logging.Level = "DEBUG"

	writer := config.NewWriter(configPath)
	err := writer.Write(cfg, config.WriteOptions{
		Format:          "yaml",
		IncludeComments: false,
	})
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "/test/dotfiles", loaded.Directories.Package)
	assert.Equal(t, "DEBUG", loaded.Logging.Level)
}

func TestWriter_WriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := config.DefaultExtended()
	cfg.Symlinks.Mode = "absolute"
	cfg.Output.Color = "always"

	writer := config.NewWriter(configPath)
	err := writer.Write(cfg, config.WriteOptions{
		Format: "json",
	})
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "absolute", loaded.Symlinks.Mode)
	assert.Equal(t, "always", loaded.Output.Color)
}

func TestWriter_WriteTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := config.DefaultExtended()
	cfg.Operations.DryRun = true
	cfg.Packages.SortBy = "links"

	writer := config.NewWriter(configPath)
	err := writer.Write(cfg, config.WriteOptions{
		Format: "toml",
	})
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.True(t, loaded.Operations.DryRun)
	assert.Equal(t, "links", loaded.Packages.SortBy)
}

func TestWriter_Update(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create initial config
	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{
		Format:          "yaml",
		IncludeComments: false,
	})
	require.NoError(t, err)

	// Update a value
	err = writer.Update("directories.package", "/new/dotfiles")
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "/new/dotfiles", loaded.Directories.Package)
}

func TestWriter_UpdateNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Update non-existent file should create it with default + update
	writer := config.NewWriter(configPath)
	err := writer.Update("logging.level", "DEBUG")
	require.NoError(t, err)

	// Load and verify
	loaded, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "DEBUG", loaded.Logging.Level)
}

func TestWriter_UpdateInvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	// Try to update with invalid key
	err = writer.Update("invalid", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key")
}

func TestWriter_UpdateInvalidValue(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	// Try to update with invalid value
	err = writer.Update("logging.level", "INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestWriter_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Non-existent subdirectory
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{Format: "yaml"})
	require.NoError(t, err)

	// Verify directory was created
	assert.DirExists(t, filepath.Dir(configPath))
	assert.FileExists(t, configPath)
}

func TestWriter_WriteWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{
		Format:          "yaml",
		IncludeComments: true,
	})
	require.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Verify comments are present
	contentStr := string(content)
	assert.Contains(t, contentStr, "# Dot Configuration File")
	assert.Contains(t, contentStr, "# Core Directories")
	assert.Contains(t, contentStr, "# Package directory containing packages")
	assert.Contains(t, contentStr, "# Logging Configuration")
	assert.Contains(t, contentStr, "# Symlink Behavior")
}

func TestWriter_WriteWithoutComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	writer := config.NewWriter(configPath)
	err := writer.WriteDefault(config.WriteOptions{
		Format:          "yaml",
		IncludeComments: false,
	})
	require.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Parse as YAML to verify it's valid
	var parsed map[string]interface{}
	err = yaml.Unmarshal(content, &parsed)
	require.NoError(t, err)

	// Should have key sections
	assert.Contains(t, parsed, "directories")
	assert.Contains(t, parsed, "logging")
	assert.Contains(t, parsed, "symlinks")
}

func TestWriter_DetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"yaml extension", "/path/to/config.yaml", "yaml"},
		{"yml extension", "/path/to/config.yml", "yaml"},
		{"json extension", "/path/to/config.json", "json"},
		{"toml extension", "/path/to/config.toml", "toml"},
		{"no extension defaults to yaml", "/path/to/config", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := config.NewWriter(tt.path)
			format := writer.DetectFormat()
			assert.Equal(t, tt.expected, format)
		})
	}
}
