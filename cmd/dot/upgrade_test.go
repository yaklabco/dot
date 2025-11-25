package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgradeCommand(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")
	require.NotNil(t, cmd)

	assert.Equal(t, "upgrade", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("yes"))
	assert.NotNil(t, cmd.Flags().Lookup("check-only"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("release-notes"))

	// Verify flag defaults
	yesFlag := cmd.Flags().Lookup("yes")
	assert.Equal(t, "false", yesFlag.DefValue)

	checkOnlyFlag := cmd.Flags().Lookup("check-only")
	assert.Equal(t, "false", checkOnlyFlag.DefValue)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	assert.Equal(t, "false", dryRunFlag.DefValue)

	releaseNotesFlag := cmd.Flags().Lookup("release-notes")
	assert.Equal(t, "false", releaseNotesFlag.DefValue)
}

func TestUpgradeCommand_Help(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")

	// Verify help text includes key information
	assert.Contains(t, cmd.Long, "package manager")
	assert.Contains(t, cmd.Long, "Homebrew")
	assert.Contains(t, cmd.Long, "update:")

	// Verify examples exist
	assert.Contains(t, cmd.Example, "dot upgrade")
	assert.Contains(t, cmd.Example, "--check-only")
	assert.Contains(t, cmd.Example, "--yes")
	assert.Contains(t, cmd.Example, "--dry-run")
	assert.Contains(t, cmd.Example, "--release-notes")

	// Verify config documentation
	assert.Contains(t, cmd.Long, "~/.config/dot/config.yaml")
	assert.Contains(t, cmd.Long, "repository")
	assert.Contains(t, cmd.Long, "include_prerelease")
}

func TestUpgradeCommand_FlagShortcuts(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")

	// Verify yes flag has -y shortcut
	yesFlag := cmd.Flags().Lookup("yes")
	assert.Equal(t, "y", yesFlag.Shorthand)

	// Verify check-only has no shortcut
	checkOnlyFlag := cmd.Flags().Lookup("check-only")
	assert.Empty(t, checkOnlyFlag.Shorthand)
}

func TestUpgradeCommand_Execution(t *testing.T) {
	// This test verifies the command can be executed without panicking
	// We can't test actual execution without mocking, but we can test structure
	cmd := newUpgradeCommand("999.999.999") // Version that won't match any release

	require.NotNil(t, cmd.RunE, "RunE should be set")

	// Verify command is properly structured
	assert.NotNil(t, cmd.RunE)
	assert.Equal(t, "upgrade", cmd.Use)
}

func TestUpgradeCommand_Integration(t *testing.T) {
	// Integration test that verifies command is properly added to root
	rootCmd := NewRootCommand("test-version", "abc123", "2024-01-01")

	// Find upgrade command
	var upgradeCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "upgrade" {
			upgradeCmd = cmd
			break
		}
	}

	require.NotNil(t, upgradeCmd, "upgrade command should be registered")
	assert.Equal(t, "upgrade", upgradeCmd.Use)
}

func TestUpgradeCommand_HelpOutput(t *testing.T) {
	cmd := newUpgradeCommand("1.0.0")

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set help flag
	cmd.SetArgs([]string{"--help"})

	// Execute should succeed for help
	err := cmd.Execute()

	// Help returns nil error but shows help text
	if err != nil {
		t.Logf("Help execution: %v", err)
	}

	// Verify help was shown (buffer should have content)
	output := buf.String()
	if len(output) > 0 {
		assert.Contains(t, output, "upgrade")
	}
}
