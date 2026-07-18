package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runDoctorSubcommand executes `dot doctor <args...>` against the given dirs
// and returns captured stdout. XDG paths are redirected under the target dir
// so the user's real config and manifest are never read or written.
func runDoctorSubcommand(t *testing.T, packageDir, targetDir string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(targetDir, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(targetDir, ".local", "share"))
	setupIntegrationTestFlags(t, CLIFlags{
		packageDir: packageDir,
		targetDir:  targetDir,
	})

	cmd := newDoctorCommand()
	cmd.SetContext(context.Background())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func newDoctorIgnoreTestDirs(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")
	require.NoError(t, os.MkdirAll(packageDir, 0o755))
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	return packageDir, targetDir
}

func TestDoctorIgnoreCommand_PatternRoundTrip(t *testing.T) {
	packageDir, targetDir := newDoctorIgnoreTestDirs(t)

	_, err := runDoctorSubcommand(t, packageDir, targetDir, "ignore", "--pattern", "Code/*")
	require.NoError(t, err)

	out, err := runDoctorSubcommand(t, packageDir, targetDir, "ignores")
	require.NoError(t, err)
	assert.Contains(t, out, "Code/*")

	_, err = runDoctorSubcommand(t, packageDir, targetDir, "unignore", "--pattern", "Code/*")
	require.NoError(t, err)

	out, err = runDoctorSubcommand(t, packageDir, targetDir, "ignores")
	require.NoError(t, err)
	assert.NotContains(t, out, "Code/*")
}

func TestDoctorIgnoreCommand_LinkRoundTrip(t *testing.T) {
	packageDir, targetDir := newDoctorIgnoreTestDirs(t)

	// A foreign symlink in the target dir, e.g. created by another tool.
	linkTarget := filepath.Join(t.TempDir(), "profile")
	require.NoError(t, os.WriteFile(linkTarget, []byte("x"), 0o644))
	require.NoError(t, os.Symlink(linkTarget, filepath.Join(targetDir, ".nix-profile")))

	_, err := runDoctorSubcommand(t, packageDir, targetDir, "ignore", ".nix-profile", "--reason", "nix managed")
	require.NoError(t, err)

	out, err := runDoctorSubcommand(t, packageDir, targetDir, "ignores")
	require.NoError(t, err)
	assert.Contains(t, out, ".nix-profile")
	assert.Contains(t, out, "nix managed")

	_, err = runDoctorSubcommand(t, packageDir, targetDir, "unignore", ".nix-profile")
	require.NoError(t, err)

	out, err = runDoctorSubcommand(t, packageDir, targetDir, "ignores")
	require.NoError(t, err)
	assert.NotContains(t, out, ".nix-profile")
}

func TestDoctorIgnoreCommand_RequiresPathOrPattern(t *testing.T) {
	packageDir, targetDir := newDoctorIgnoreTestDirs(t)

	_, err := runDoctorSubcommand(t, packageDir, targetDir, "ignore")
	assert.Error(t, err, "ignore without a path or --pattern must fail")

	_, err = runDoctorSubcommand(t, packageDir, targetDir, "ignore", ".foo", "--pattern", "bar/*")
	assert.Error(t, err, "ignore with both a path and --pattern must fail")
}

func TestDoctorIgnoresCommand_EmptyState(t *testing.T) {
	packageDir, targetDir := newDoctorIgnoreTestDirs(t)

	out, err := runDoctorSubcommand(t, packageDir, targetDir, "ignores")
	require.NoError(t, err)
	assert.Contains(t, out, "No ignored links or patterns")
}
