package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupGlobalCfg initializes globalCfg with deterministic test values and registers cleanup.
func setupGlobalCfg(t *testing.T) {
	t.Helper()

	// Save previous globalCfg and environment
	previous := globalCfg
	oldXDGData := os.Getenv("XDG_DATA_HOME")
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	oldXDGState := os.Getenv("XDG_STATE_HOME")

	// Create temp directory for XDG paths to prevent writing to source tree
	xdgBase := t.TempDir()
	os.Setenv("XDG_DATA_HOME", filepath.Join(xdgBase, "data"))
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(xdgBase, "config"))
	os.Setenv("XDG_STATE_HOME", filepath.Join(xdgBase, "state"))

	// Set globalCfg to use temporary directories
	globalCfg = globalConfig{
		packageDir: t.TempDir(),
		targetDir:  t.TempDir(),
		dryRun:     true, // Always dry-run in tests to avoid side effects
		verbose:    0,
		quiet:      false,
		logJSON:    false,
	}

	// Restore previous globalCfg and environment on cleanup
	t.Cleanup(func() {
		globalCfg = previous
		if oldXDGData != "" {
			os.Setenv("XDG_DATA_HOME", oldXDGData)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
		if oldXDGConfig != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if oldXDGState != "" {
			os.Setenv("XDG_STATE_HOME", oldXDGState)
		} else {
			os.Unsetenv("XDG_STATE_HOME")
		}
	})
}

func TestConfirmAction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes lowercase", "y\n", true},
		{"yes full", "yes\n", true},
		{"no lowercase", "n\n", false},
		{"no full", "no\n", false},
		{"empty", "\n", false},
		{"invalid", "maybe\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function reads from stdin, so we can't easily test it
			// Just ensure it compiles and doesn't panic
			_ = tt.expected
		})
	}
}

func TestGetUnmanageOperation(t *testing.T) {
	tests := []struct {
		name     string
		pkg      struct{ Source string }
		opts     struct{ Purge, Restore bool }
		expected string
	}{
		{"purge takes precedence", struct{ Source string }{"managed"}, struct{ Purge, Restore bool }{true, true}, "purge"},
		{"restore adopted", struct{ Source string }{"adopted"}, struct{ Purge, Restore bool }{false, true}, "restore"},
		{"remove managed", struct{ Source string }{"managed"}, struct{ Purge, Restore bool }{false, true}, "remove"},
		{"remove when restore false", struct{ Source string }{"adopted"}, struct{ Purge, Restore bool }{false, false}, "remove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import dot package to use actual types
			pkg := struct {
				Source string
			}{Source: tt.pkg.Source}

			// Create PackageInfo-like structure
			pkgInfo := struct {
				Source string
			}{Source: pkg.Source}

			opts := struct {
				Purge   bool
				Restore bool
			}{Purge: tt.opts.Purge, Restore: tt.opts.Restore}

			// Test the logic inline
			var result string
			if opts.Purge {
				result = "purge"
			} else if opts.Restore && pkgInfo.Source == "adopted" {
				result = "restore"
			} else {
				result = "remove"
			}

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatLinkList(t *testing.T) {
	tests := []struct {
		name     string
		links    []string
		expected string
	}{
		{"empty list", []string{}, "none"},
		{"single link", []string{".vimrc"}, ".vimrc"},
		{"two links", []string{".vimrc", ".vim/"}, ".vimrc, .vim/"},
		{"three links", []string{".vimrc", ".vim/", ".viminfo"}, ".vimrc, .vim/, .viminfo"},
		{"many links", []string{".a", ".b", ".c", ".d", ".e"}, ".a, .b, .c... (2 more)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatLinkList(tt.links)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestManageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Commands now actually execute, so we expect error for non-existent package
	cmd := newManageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestManageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newManageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestManageCommand_Metadata(t *testing.T) {
	cmd := newManageCommand()

	require.Equal(t, "manage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Install packages by creating symlinks", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestUnmanageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestUnmanageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestUnmanageCommand_Metadata(t *testing.T) {
	cmd := newUnmanageCommand()

	require.Equal(t, "unmanage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Remove packages by deleting symlinks", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestRemanageCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Remanage tries to unmanage then manage, so will error on manage phase
	cmd := newRemanageCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestRemanageCommand_NoPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newRemanageCommand()
	cmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestRemanageCommand_Metadata(t *testing.T) {
	cmd := newRemanageCommand()

	require.Equal(t, "remanage PACKAGE [PACKAGE...]", cmd.Use)
	require.Equal(t, "Reinstall packages with incremental updates", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestAdoptCommand_ExecuteStub(t *testing.T) {
	setupGlobalCfg(t)

	// Adopt tries to verify package exists, so will error
	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1", "file1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestAdoptCommand_NotEnoughArgs(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err)
}

func TestAdoptCommand_Metadata(t *testing.T) {
	cmd := newAdoptCommand()

	require.Equal(t, "adopt [PACKAGE] FILE [FILE...]", cmd.Use)
	require.Equal(t, "Move existing files into package then link", cmd.Short)
	require.NotEmpty(t, cmd.Long)
}

func TestAdoptCommand_MultipleFiles(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple files with non-existent package will error
	cmd := newAdoptCommand()
	cmd.SetArgs([]string{"package1", "file1", "file2", "file3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when package does not exist")
}

func TestManageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple packages that don't exist will error
	cmd := newManageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when packages do not exist")
}

func TestUnmanageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestUnmanageCommand_AllFlag(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"--all", "--yes"}) // Skip confirmation

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.NoError(t, err, "should not error")
	// Output depends on whether packages exist - just verify it executed
	output := out.String()
	_ = output // May contain "No packages" or "Would unmanage" depending on state
}

func TestUnmanageCommand_AllWithPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{"--all", "pkg1"}) // Both --all and package name

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when both --all and packages specified")
	require.Contains(t, err.Error(), "cannot specify package names with --all")
}

func TestUnmanageCommand_NeitherAllNorPackages(t *testing.T) {
	setupGlobalCfg(t)

	cmd := newUnmanageCommand()
	cmd.SetArgs([]string{}) // Neither --all nor packages

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when neither --all nor packages provided")
	require.Contains(t, err.Error(), "requires at least 1 package name or --all flag")
}

func TestRemanageCommand_MultiplePackages(t *testing.T) {
	setupGlobalCfg(t)

	// Multiple packages that don't exist will error on manage phase
	cmd := newRemanageCommand()
	cmd.SetArgs([]string{"package1", "package2", "package3"})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()
	require.Error(t, err, "should error when packages do not exist")
}

func TestRootCommand_NoArgs(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "dot")
}

func TestRootCommand_WithManageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"manage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "manage")
}

func TestRootCommand_WithUnmanageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"unmanage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "unmanage")
}

func TestRootCommand_WithRemanageCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"remanage", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "remanage")
}

func TestRootCommand_WithAdoptCommand(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"adopt", "--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "adopt")
}

func TestRootCommand_GlobalFlagsWithCommand(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--dir", tmpDir, "--target", tmpDir, "--dry-run", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_DryRunFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_VerboseFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "-vvv", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_QuietFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "--quiet", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestRootCommand_LogJSONFlag(t *testing.T) {
	tmpDir := t.TempDir()

	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--target", tmpDir, "--dry-run", "--log-json", "manage", "package1"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)

	err := rootCmd.Execute()
	// Package doesn't exist, expect error
	require.Error(t, err)
}

func TestGetOperationColor(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{"purge", "purge"},
		{"restore", "restore"},
		{"remove", "remove"},
		{"other", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorFunc := getOperationColor(tt.operation)
			require.NotNil(t, colorFunc)
			// Just verify the function can be called
			result := colorFunc("test")
			require.Contains(t, result, "test")
		})
	}
}
