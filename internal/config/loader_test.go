package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
)

func TestLoadFromFile_WithYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /test/dotfiles
  target: /test/home

logging:
  level: DEBUG
  format: json
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadExtendedFromFile(configPath)
	require.NoError(t, err)

	assert.Equal(t, "/test/dotfiles", cfg.Directories.Package)
	assert.Equal(t, "/test/home", cfg.Directories.Target)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestNewLoader(t *testing.T) {
	loader := config.NewLoader("dot", "/path/to/config.yaml")
	assert.NotNil(t, loader)
}

func TestLoader_LoadWithMissingFile(t *testing.T) {
	loader := config.NewLoader("dot", "/nonexistent/config.yaml")
	cfg, err := loader.Load()
	require.NoError(t, err)

	// Should return defaults
	assert.NotNil(t, cfg)
	assert.Equal(t, "INFO", cfg.Logging.Level)
}

func TestLoader_LoadWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
logging:
  level: INFO
  format: text
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("DOT_LOGGING_LEVEL", "DEBUG")
	os.Setenv("DOT_LOGGING_FORMAT", "json")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")
	defer os.Unsetenv("DOT_LOGGING_FORMAT")

	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	require.NoError(t, err)

	// Environment should override file
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoader_LoadWithFlags(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /file/dotfiles
  target: /file/home

output:
  verbosity: 1
  color: auto
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	loader := config.NewLoader("dot", configPath)
	flags := map[string]interface{}{
		"dir":     "/flag/dotfiles",
		"verbose": 2,
		"color":   "always",
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Flags should override file
	assert.Equal(t, "/flag/dotfiles", cfg.Directories.Package)
	assert.Equal(t, 2, cfg.Output.Verbosity)
	assert.Equal(t, "always", cfg.Output.Color)
	// File value for non-overridden
	assert.Equal(t, "/file/home", cfg.Directories.Target)
}

func TestLoader_Precedence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := `
directories:
  package: /file/dotfiles

logging:
  level: INFO
  format: text

output:
  verbosity: 1
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("DOT_LOGGING_LEVEL", "WARN")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")

	loader := config.NewLoader("dot", configPath)
	flags := map[string]interface{}{
		"verbose": 2,
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Verify precedence: flags > env > file > default
	assert.Equal(t, "/file/dotfiles", cfg.Directories.Package) // from file
	assert.Equal(t, "WARN", cfg.Logging.Level)                 // from env (overrides file)
	assert.Equal(t, 2, cfg.Output.Verbosity)                   // from flags (highest)
	assert.Equal(t, "text", cfg.Logging.Format)                // from file (no override)
}

func TestLoader_ValidateOnLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create invalid config file
	configContent := `
logging:
  level: INVALID_LEVEL
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	loader := config.NewLoader("dot", configPath)
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestLoader_FlagMapping(t *testing.T) {
	loader := config.NewLoader("dot", "/nonexistent/config.yaml")

	tests := []struct {
		name     string
		flags    map[string]interface{}
		validate func(*testing.T, *config.ExtendedConfig)
	}{
		{
			name: "dir flag",
			flags: map[string]interface{}{
				"dir": "/custom/dotfiles",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "/custom/dotfiles", cfg.Directories.Package)
			},
		},
		{
			name: "target flag",
			flags: map[string]interface{}{
				"target": "/custom/target",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "/custom/target", cfg.Directories.Target)
			},
		},
		{
			name: "dry-run flag",
			flags: map[string]interface{}{
				"dry-run": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.True(t, cfg.Operations.DryRun)
			},
		},
		{
			name: "verbose flag",
			flags: map[string]interface{}{
				"verbose": 2,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, 2, cfg.Output.Verbosity)
			},
		},
		{
			name: "quiet flag",
			flags: map[string]interface{}{
				"quiet": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, 0, cfg.Output.Verbosity)
			},
		},
		{
			name: "log-json flag",
			flags: map[string]interface{}{
				"log-json": true,
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "json", cfg.Logging.Format)
			},
		},
		{
			name: "color flag",
			flags: map[string]interface{}{
				"color": "never",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "never", cfg.Output.Color)
			},
		},
		{
			name: "format flag",
			flags: map[string]interface{}{
				"format": "json",
			},
			validate: func(t *testing.T, cfg *config.ExtendedConfig) {
				assert.Equal(t, "json", cfg.Output.Format)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loader.LoadWithFlags(tt.flags)
			require.NoError(t, err)
			tt.validate(t, cfg)
		})
	}
}

func TestLoader_MultipleSourcesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file with baseline values
	configContent := `
directories:
  package: /file/dotfiles
  target: /file/home

logging:
  level: INFO
  format: text

symlinks:
  mode: relative
  folding: true

output:
  verbosity: 1
  color: auto
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variables to override some file values
	os.Setenv("DOT_LOGGING_LEVEL", "WARN")
	os.Setenv("DOT_SYMLINKS_MODE", "absolute")
	defer os.Unsetenv("DOT_LOGGING_LEVEL")
	defer os.Unsetenv("DOT_SYMLINKS_MODE")

	loader := config.NewLoader("dot", configPath)

	// Load with flags to override env and file
	flags := map[string]interface{}{
		"verbose": 3,
		"color":   "never",
	}

	cfg, err := loader.LoadWithFlags(flags)
	require.NoError(t, err)

	// Verify precedence: flags > env > file > defaults
	assert.Equal(t, "/file/dotfiles", cfg.Directories.Package) // from file (no override)
	assert.Equal(t, "/file/home", cfg.Directories.Target)      // from file (no override)
	assert.Equal(t, "WARN", cfg.Logging.Level)                 // from env (overrides file)
	assert.Equal(t, "text", cfg.Logging.Format)                // from file (no override)
	assert.Equal(t, "absolute", cfg.Symlinks.Mode)             // from env (overrides file)
	assert.True(t, cfg.Symlinks.Folding)                       // from file (no override)
	assert.Equal(t, 3, cfg.Output.Verbosity)                   // from flags (highest priority)
	assert.Equal(t, "never", cfg.Output.Color)                 // from flags (highest priority)
}

func TestLoader_AutoDetectFormat(t *testing.T) {
	tmpDir := t.TempDir()

	formats := []struct {
		ext     string
		content string
	}{
		{
			ext: ".yaml",
			content: `
directories:
  package: /test/dotfiles
`,
		},
		{
			ext: ".json",
			content: `{
  "directories": {
    "package": "/test/dotfiles"
  }
}`,
		},
		{
			ext: ".toml",
			content: `[directories]
package = "/test/dotfiles"
`,
		},
	}

	for _, fmt := range formats {
		t.Run("load from "+fmt.ext, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, "config"+fmt.ext)
			err := os.WriteFile(configPath, []byte(fmt.content), 0600)
			require.NoError(t, err)

			loader := config.NewLoader("dot", configPath)
			cfg, err := loader.Load()
			require.NoError(t, err)
			assert.Equal(t, "/test/dotfiles", cfg.Directories.Package)
		})
	}
}
