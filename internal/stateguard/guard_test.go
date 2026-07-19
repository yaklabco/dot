package stateguard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/manifest"
)

func TestCheck_MarkerExists(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)
	require.NoError(t, WriteMarker())

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader(""),
		Out:         &bytes.Buffer{},
		ManifestDir: t.TempDir(),
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultAlreadyAcknowledged, result)
}

func TestCheck_NoManifest(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader(""),
		Out:         &bytes.Buffer{},
		ManifestDir: t.TempDir(),
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultNoop, result)
}

func TestCheck_EmptyManifest(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	manifestDir := t.TempDir()
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{})

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader(""),
		Out:         &bytes.Buffer{},
		ManifestDir: manifestDir,
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultNoop, result)
}

func TestCheck_Continue(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	manifestDir := t.TempDir()
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {Links: []string{".bashrc"}, LinkCount: 1},
	})

	out := &bytes.Buffer{}
	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader("1\n"),
		Out:         out,
		ManifestDir: manifestDir,
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultContinue, result)
	assert.True(t, MarkerExists())
	assert.Contains(t, out.String(), "Existing dot installation detected")
}

func TestCheck_Fresh(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()

	// Create a symlink to be cleaned up
	symlinkPath := filepath.Join(targetDir, ".bashrc")
	require.NoError(t, os.Symlink("/some/source", symlinkPath))

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc"},
			LinkCount: 1,
			TargetDir: targetDir,
		},
	})

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader("2\n"),
		Out:         &bytes.Buffer{},
		ManifestDir: manifestDir,
		TargetDir:   targetDir,
	})
	require.NoError(t, err)
	assert.Equal(t, ResultFresh, result)
	assert.True(t, MarkerExists())

	// Symlink should be removed
	_, err = os.Lstat(symlinkPath)
	assert.True(t, os.IsNotExist(err))
}

func TestCheck_BackupAndFresh(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	targetDir := t.TempDir()
	manifestDir := t.TempDir()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("test"), 0644))

	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {
			Links:     []string{".bashrc"},
			LinkCount: 1,
			TargetDir: targetDir,
		},
	})

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader("3\n"),
		Out:         &bytes.Buffer{},
		ManifestDir: manifestDir,
		TargetDir:   targetDir,
		ConfigPath:  configPath,
		HomeDir:     t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultBackupAndFresh, result)
	assert.True(t, MarkerExists())
}

func TestCheck_SkipAutosContinue(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	manifestDir := t.TempDir()
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {Links: []string{".bashrc"}, LinkCount: 1},
	})

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader(""),
		Out:         &bytes.Buffer{},
		Skip:        true,
		ManifestDir: manifestDir,
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultContinue, result)
	assert.True(t, MarkerExists())
}

func TestCheck_EmptyInputDefaultsContinue(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", stateDir)

	manifestDir := t.TempDir()
	writeTestManifest(t, manifestDir, map[string]manifest.PackageInfo{
		"shell": {Links: []string{".bashrc"}, LinkCount: 1},
	})

	result, err := Check(context.Background(), GuardOptions{
		In:          strings.NewReader("\n"),
		Out:         &bytes.Buffer{},
		ManifestDir: manifestDir,
		TargetDir:   t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, ResultContinue, result)
	assert.True(t, MarkerExists())
}
