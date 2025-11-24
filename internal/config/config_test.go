package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()

	assert.NotNil(t, cfg)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLoadFromFile_YAML(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
log_level: DEBUG
log_format: text
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadFromFile(configFile)
	require.NoError(t, err)

	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
}

func TestLoadFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configContent := `{
  "log_level": "WARN",
  "log_format": "json"
}`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadFromFile(configFile)
	require.NoError(t, err)

	assert.Equal(t, "WARN", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLoadFromFile_TOML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")

	configContent := `
log_level = "ERROR"
log_format = "text"
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadFromFile(configFile)
	require.NoError(t, err)

	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
}

func TestLoadFromFile_NotFound(t *testing.T) {
	cfg, err := config.LoadFromFile("/nonexistent/config.yaml")

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				LogLevel:  "INFO",
				LogFormat: "json",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &config.Config{
				LogLevel:  "INVALID",
				LogFormat: "json",
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			config: &config.Config{
				LogLevel:  "INFO",
				LogFormat: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigPrecedence(t *testing.T) {
	// Test that environment variables override file config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
log_level: INFO
log_format: json
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("DOT_LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("DOT_LOG_LEVEL")

	cfg, err := config.LoadWithEnv(configFile)
	require.NoError(t, err)

	// Environment variable should override file
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat) // File value for non-overridden
}

func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name        string
		appName     string
		envVars     map[string]string
		shouldExist bool
	}{
		{
			name:    "with XDG_CONFIG_HOME set",
			appName: "testapp",
			envVars: map[string]string{
				"XDG_CONFIG_HOME": "/tmp/config",
			},
			shouldExist: false,
		},
		{
			name:        "without XDG_CONFIG_HOME",
			appName:     "testapp",
			envVars:     map[string]string{},
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			path := config.GetConfigPath(tt.appName)
			assert.NotEmpty(t, path)
			assert.Contains(t, path, tt.appName)
		})
	}
}

func TestContainsFunction(t *testing.T) {
	// Test the validation logic indirectly through Validate
	cfg := &config.Config{
		LogLevel:  "INFO",
		LogFormat: "json",
	}

	err := cfg.Validate()
	assert.NoError(t, err)

	// Test with all valid log levels
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for _, level := range validLevels {
		cfg.LogLevel = level
		err := cfg.Validate()
		assert.NoError(t, err, "Level %s should be valid", level)
	}

	// Test with all valid formats
	validFormats := []string{"json", "text"}
	for _, format := range validFormats {
		cfg.LogFormat = format
		err := cfg.Validate()
		assert.NoError(t, err, "Format %s should be valid", format)
	}
}

func TestLoadWithEnv_FileNotFound(t *testing.T) {
	// Test that missing file is handled gracefully with defaults
	cfg, err := config.LoadWithEnv("/nonexistent/config.yaml")
	require.NoError(t, err)

	// Should return defaults
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLoadWithEnv_EnvOnly(t *testing.T) {
	// Test loading with only environment variables
	os.Setenv("DOT_LOG_LEVEL", "ERROR")
	os.Setenv("DOT_LOG_FORMAT", "text")
	defer os.Unsetenv("DOT_LOG_LEVEL")
	defer os.Unsetenv("DOT_LOG_FORMAT")

	cfg, err := config.LoadWithEnv("/nonexistent/config.yaml")
	require.NoError(t, err)

	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
}
