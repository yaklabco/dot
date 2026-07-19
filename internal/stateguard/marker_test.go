package stateguard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkerPath_Default(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	path := MarkerPath()
	homeDir, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(homeDir, ".local", "state", "dot", ".dot-acknowledged"), path)
}

func TestMarkerPath_WithXDGOverride(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/custom-state")
	path := MarkerPath()
	assert.Equal(t, "/tmp/custom-state/dot/.dot-acknowledged", path)
}

func TestMarkerExists_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	assert.False(t, MarkerExists())
}

func TestMarkerExists_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	markerDir := filepath.Join(dir, "dot")
	require.NoError(t, os.MkdirAll(markerDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(markerDir, ".dot-acknowledged"), nil, 0644))

	assert.True(t, MarkerExists())
}

func TestWriteMarker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	assert.False(t, MarkerExists())
	require.NoError(t, WriteMarker())
	assert.True(t, MarkerExists())
}

func TestWriteMarker_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested")
	t.Setenv("XDG_STATE_HOME", nested)

	require.NoError(t, WriteMarker())
	assert.True(t, MarkerExists())
}

func TestWriteMarker_Idempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	require.NoError(t, WriteMarker())
	require.NoError(t, WriteMarker())
	assert.True(t, MarkerExists())
}
