package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/stateguard"
)

func TestRunStateGuard_NoState(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	cmd := newManageCommand()
	cmd.SetArgs([]string{})
	err := runStateGuard(cmd)
	assert.NoError(t, err)
}

func TestRunStateGuard_MarkerExists(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	require.NoError(t, stateguard.WriteMarker())

	cmd := newManageCommand()
	cmd.SetArgs([]string{})
	err := runStateGuard(cmd)
	assert.NoError(t, err)
}

func TestRunStateGuard_BatchMode(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	// Write manifest into the path resolveGuardPaths will compute
	manifestDir := filepath.Join(dataHome, "dot", "manifest")
	require.NoError(t, os.MkdirAll(manifestDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(manifestDir, ".dot-manifest.json"), []byte(`{
		"version": "1.0",
		"packages": {"shell": {"name": "shell", "links": [".bashrc"], "link_count": 1}}
	}`), 0644))

	// In batch mode, should auto-continue without interactive prompt
	cliFlags.batch = true
	defer func() { cliFlags.batch = false }()

	cmd := newManageCommand()
	cmd.SetArgs([]string{})
	err := runStateGuard(cmd)
	assert.NoError(t, err)
	assert.True(t, stateguard.MarkerExists())
}

func TestResolveGuardPaths(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/test-data")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")

	manifestDir, targetDir, configPath, homeDir := resolveGuardPaths()

	assert.NotEmpty(t, manifestDir)
	assert.NotEmpty(t, targetDir)
	assert.NotEmpty(t, configPath)
	assert.NotEmpty(t, homeDir)
}
