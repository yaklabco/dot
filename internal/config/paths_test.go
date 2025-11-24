package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/config"
)

func TestGetConfigPath_WithXDGSet(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	path := config.GetConfigPath("dot")
	assert.Contains(t, path, "/tmp/test-config")
	assert.Contains(t, path, "dot")
}

func TestGetConfigPath_WithoutXDG(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")

	path := config.GetConfigPath("dot")
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "dot")
}

func TestGetConfigPath_EmptyApp(t *testing.T) {
	path := config.GetConfigPath("")
	assert.NotEmpty(t, path)
}

func TestDefaultExtended_AllFieldsSet(t *testing.T) {
	cfg := config.DefaultExtended()

	// Verify all sections initialized
	assert.NotEmpty(t, cfg.Directories.Package)
	assert.NotEmpty(t, cfg.Directories.Target)
	assert.NotEmpty(t, cfg.Directories.Manifest)
	assert.NotEmpty(t, cfg.Logging.Level)
	assert.NotEmpty(t, cfg.Symlinks.Mode)
	assert.NotEmpty(t, cfg.Output.Format)
	assert.NotNil(t, cfg.Ignore.UseDefaults)
	assert.NotNil(t, cfg.Dotfile.Translate)
	assert.NotNil(t, cfg.Operations.DryRun)
}
