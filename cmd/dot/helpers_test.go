package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatError(t *testing.T) {
	err := errors.New("test error")
	result := formatError(err)
	assert.Equal(t, err, result)
}

func TestShouldColorize(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{"always", "always", true},
		{"never", "never", false},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore globalCfg
			previous := globalCfg
			t.Cleanup(func() {
				globalCfg = previous
			})
			globalCfg.noColor = false

			// Unset NO_COLOR for this test to ensure it doesn't interfere
			original := os.Getenv("NO_COLOR")
			os.Unsetenv("NO_COLOR")
			t.Cleanup(func() {
				if original != "" {
					os.Setenv("NO_COLOR", original)
				}
			})

			result := shouldColorize(tt.color)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestShouldColorizeWithNoColorFlag(t *testing.T) {
	t.Run("--no-color takes precedence over always", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg.noColor = true

		result := shouldColorize("always")
		assert.False(t, result, "--no-color should disable colors even with --color=always")
	})

	t.Run("--no-color takes precedence over NO_COLOR unset", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg.noColor = true

		original := os.Getenv("NO_COLOR")
		os.Unsetenv("NO_COLOR")
		t.Cleanup(func() {
			if original != "" {
				os.Setenv("NO_COLOR", original)
			}
		})

		result := shouldColorize("always")
		assert.False(t, result, "--no-color should disable colors")
	})

	t.Run("NO_COLOR env takes precedence over --color=always when --no-color is false", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg.noColor = false

		original := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		t.Cleanup(func() {
			if original == "" {
				os.Unsetenv("NO_COLOR")
			} else {
				os.Setenv("NO_COLOR", original)
			}
		})

		result := shouldColorize("always")
		assert.False(t, result, "NO_COLOR should disable colors when --no-color is not set")
	})

	t.Run("colors enabled when neither flag nor env set", func(t *testing.T) {
		previous := globalCfg
		t.Cleanup(func() {
			globalCfg = previous
		})
		globalCfg.noColor = false

		original := os.Getenv("NO_COLOR")
		os.Unsetenv("NO_COLOR")
		t.Cleanup(func() {
			if original != "" {
				os.Setenv("NO_COLOR", original)
			}
		})

		result := shouldColorize("always")
		assert.True(t, result, "colors should be enabled with --color=always")
	})
}

func TestShouldColorize_Auto(t *testing.T) {
	// Auto detection depends on if stdout is a TTY
	// In tests, it's typically not a TTY, so should be false
	result := shouldColorize("auto")
	// Just verify it doesn't panic, actual result depends on environment
	_ = result
}

func TestBuildConfig_ValidatesPackageDir(t *testing.T) {
	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: ".",
		targetDir:  ".",
	}

	cfg, err := buildConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.PackageDir)
}

func TestCreateLogger_AllModes(t *testing.T) {
	tests := []struct {
		name    string
		quiet   bool
		logJSON bool
		verbose int
	}{
		{"quiet", true, false, 0},
		{"json", false, true, 0},
		{"text", false, false, 0},
		{"verbose", false, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := globalCfg
			t.Cleanup(func() {
				globalCfg = previous
			})

			globalCfg = globalConfig{
				quiet:   tt.quiet,
				logJSON: tt.logJSON,
				verbose: tt.verbose,
			}

			logger := createLogger()
			assert.NotNil(t, logger)
		})
	}
}

func TestIsHiddenOrIgnored(t *testing.T) {
	tests := []struct {
		name     string
		dirname  string
		expected bool
	}{
		{"hidden directory", ".hidden", true},
		{"git directory", ".git", true},
		{"node_modules", "node_modules", true},
		{"vendor", "vendor", true},
		{"normal package", "vim", false},
		{"normal package with dash", "dot-vim", false},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHiddenOrIgnored(tt.dirname)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAvailablePackages(t *testing.T) {
	tmpDir := t.TempDir()

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: tmpDir,
	}

	// Create some test package directories
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "vim"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tmux"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".hidden"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644))

	packages := getAvailablePackages()

	// Should return package directories, but not hidden or files
	assert.Contains(t, packages, "vim")
	assert.Contains(t, packages, "tmux")
	assert.NotContains(t, packages, ".hidden")
	assert.NotContains(t, packages, "file.txt")
}

func TestPackageCompletion_Available(t *testing.T) {
	tmpDir := t.TempDir()

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: tmpDir,
	}

	// Create test packages
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "vim"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "tmux"), 0755))

	completionFunc := packageCompletion(false)
	cmd := &cobra.Command{}

	completions, directive := completionFunc(cmd, []string{}, "")

	assert.Contains(t, completions, "vim")
	assert.Contains(t, completions, "tmux")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestGetInstalledPackages(t *testing.T) {
	// Test that getInstalledPackages doesn't crash
	packages := getInstalledPackages()
	// Should return a list (possibly empty, possibly with real packages)
	// We can't fully isolate this without mocking the entire config system
	assert.NotNil(t, packages)
}

func TestPackageCompletion_Installed(t *testing.T) {
	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: t.TempDir(),
		targetDir:  t.TempDir(),
	}

	completionFunc := packageCompletion(true)
	cmd := &cobra.Command{}

	completions, directive := completionFunc(cmd, []string{}, "")

	// Should return empty for no installed packages
	assert.NotNil(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestDerivePackageName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"dotfile with leading dot", ".ssh", ".ssh"},      // Changed: keeps dot
		{"dotfile vimrc", ".vimrc", ".vimrc"},             // Changed: keeps dot
		{"dotfile config", ".config", ".config"},          // Changed: keeps dot
		{"dotfile tmux.conf", ".tmux.conf", ".tmux.conf"}, // Changed: keeps dot
		{"regular file", "file.txt", "file.txt"},
		{"directory path", ".config/nvim", "nvim"}, // Base name has no dot
		{"nested dotfile", ".local/bin", "bin"},    // Base name has no dot
		{"no leading dot", "myfile", "myfile"},
		{"just dot", ".", ""},
		{"double dot", "..", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derivePackageName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAvailablePackages_WithFlags(t *testing.T) {
	tmpDir := t.TempDir()

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	// Test with explicit package dir
	globalCfg = globalConfig{
		packageDir: tmpDir,
	}

	// Create test packages
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "package1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "package2"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".hidden"), 0755))

	packages := getAvailablePackages()

	assert.Contains(t, packages, "package1")
	assert.Contains(t, packages, "package2")
	assert.NotContains(t, packages, ".hidden")
}

func TestGetAvailablePackages_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: tmpDir,
	}

	packages := getAvailablePackages()
	assert.Empty(t, packages)
}

func TestGetAvailablePackages_InvalidDir(t *testing.T) {
	previous := globalCfg
	t.Cleanup(func() {
		globalCfg = previous
	})

	globalCfg = globalConfig{
		packageDir: "/this/path/definitely/does/not/exist/anywhere",
	}

	packages := getAvailablePackages()
	assert.Nil(t, packages)
}
