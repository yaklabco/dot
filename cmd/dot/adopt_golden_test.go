package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/jamesainslie/dot/internal/cli/golden"
	"github.com/stretchr/testify/require"
)

// normalizePaths replaces dynamic paths with placeholders for consistent comparison
func normalizePaths(output string) string {
	// Replace temp directory paths
	re := regexp.MustCompile(`/var/folders/[^/]+/[^/]+/T/[^\s]+`)
	output = re.ReplaceAllString(output, "<TMPDIR>")

	re = regexp.MustCompile(`/tmp/[^\s]+`)
	output = re.ReplaceAllString(output, "<TMPDIR>")

	// Replace current working directory (for error messages that include it)
	if wd, err := os.Getwd(); err == nil {
		re = regexp.MustCompile(regexp.QuoteMeta(wd))
		output = re.ReplaceAllString(output, "<CWD>")
	}

	// Replace home directory references
	if home := os.Getenv("HOME"); home != "" {
		re = regexp.MustCompile(regexp.QuoteMeta(home))
		output = re.ReplaceAllString(output, "<HOME>")
	}

	return output
}

// TestAdoptCommand_Golden tests the adopt command with various scenarios
func TestAdoptCommand_Golden(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping golden tests in short mode")
	}

	tests := []struct {
		name        string
		args        []string
		setupFunc   func(*testing.T) (targetDir, packageDir string, cleanup func())
		expectError bool
	}{
		{
			name: "adopt_single_file_auto_name",
			args: []string{"adopt", ".bashrc"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create file to adopt
				bashrc := filepath.Join(targetDir, ".bashrc")
				os.WriteFile(bashrc, []byte("# My bashrc\nexport PATH=$PATH:/usr/local/bin\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_single_file_explicit_package",
			args: []string{"adopt", "vim", ".vimrc"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create file to adopt
				vimrc := filepath.Join(targetDir, ".vimrc")
				os.WriteFile(vimrc, []byte("\" Vim config\nset number\nset ruler\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_multiple_files_explicit_package",
			args: []string{"adopt", "bash", ".bashrc", ".bash_profile"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create files to adopt
				os.WriteFile(filepath.Join(targetDir, ".bashrc"), []byte("# bashrc\n"), 0644)
				os.WriteFile(filepath.Join(targetDir, ".bash_profile"), []byte("# bash_profile\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_nested_config_file",
			args: []string{"adopt", "fish", ".config/fish/config.fish"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create nested config file
				configDir := filepath.Join(targetDir, ".config", "fish")
				os.MkdirAll(configDir, 0755)
				os.WriteFile(filepath.Join(configDir, "config.fish"), []byte("# Fish config\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_with_absolute_path",
			args: []string{"adopt", "testpkg", "test-adopt-file.txt"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := tmpDir // Use tmpDir as target
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(packageDir, 0755)

				// Create file in target directory
				testFile := filepath.Join(targetDir, "test-adopt-file.txt")
				os.WriteFile(testFile, []byte("test content\n"), 0644)

				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)

				return targetDir, packageDir, func() {
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_dry_run",
			args: []string{"adopt", "--dry-run", "git", ".gitconfig"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create file to adopt
				os.WriteFile(filepath.Join(targetDir, ".gitconfig"), []byte("[user]\nname=Test\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_with_verbose",
			args: []string{"adopt", "-v", "zsh", ".zshrc"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create file to adopt
				os.WriteFile(filepath.Join(targetDir, ".zshrc"), []byte("# Zsh config\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
		{
			name: "adopt_hidden_directory",
			args: []string{"adopt", "ssh", ".ssh/config"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create .ssh directory and config
				sshDir := filepath.Join(targetDir, ".ssh")
				os.MkdirAll(sshDir, 0700)
				os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host github.com\n"), 0600)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},

		// Error scenarios
		{
			name:        "adopt_no_arguments",
			args:        []string{"adopt"},
			setupFunc:   func(t *testing.T) (string, string, func()) { return "", "", func() {} },
			expectError: true,
		},
		{
			name: "adopt_nonexistent_file",
			args: []string{"adopt", "pkg", ".nonexistent"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: true,
		},
		{
			name: "adopt_multiple_files_no_package",
			args: []string{"adopt", ".file1", ".file2"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				os.WriteFile(filepath.Join(targetDir, ".file1"), []byte("content1\n"), 0644)
				os.WriteFile(filepath.Join(targetDir, ".file2"), []byte("content2\n"), 0644)

				os.Setenv("HOME", tmpDir)
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false, // With new adopt logic, it auto-names from first file
		},
		{
			name: "adopt_with_pwd_relative_path",
			args: []string{"adopt", "ado-cli", "./ado-cli"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target", ".config")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create directory in .config directory
				testFile := filepath.Join(targetDir, "ado-cli")
				os.MkdirAll(testFile, 0755)
				os.WriteFile(filepath.Join(testFile, "config.json"), []byte(`{"key": "value"}`), 0644)

				os.Setenv("HOME", filepath.Join(tmpDir, "target"))
				oldDir, _ := os.Getwd()
				os.Chdir(targetDir) // Change to .config directory

				return filepath.Join(tmpDir, "target"), packageDir, func() {
					os.Unsetenv("HOME")
					os.Chdir(oldDir)
				}
			},
			expectError: false,
		},
	}

	// Get current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Ensure testdata directory exists
	testdataDir := filepath.Join("testdata", "golden", "adopt")
	os.MkdirAll(testdataDir, 0755)

	g := golden.New(t, "adopt")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			targetDir, packageDir, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer

			// Create root command
			rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)

			// Set global flags for package and target directories
			if targetDir != "" && packageDir != "" {
				tt.args = append([]string{"--target", targetDir, "--dir", packageDir}, tt.args...)
			}
			rootCmd.SetArgs(tt.args)

			// Execute command
			ctx := context.Background()
			_, err := executeCommand(ctx, rootCmd)

			// Restore working directory immediately after command execution
			// This must be done before golden file comparison
			os.Chdir(originalWd)

			// Verify error expectation
			if tt.expectError {
				require.Error(t, err, "expected error for: %s", tt.name)
			} else {
				if err != nil {
					t.Logf("Unexpected error: %v", err)
					t.Logf("Stdout: %s", stdout.String())
					t.Logf("Stderr: %s", stderr.String())
				}
				require.NoError(t, err, "unexpected error for: %s", tt.name)
			}

			// Combine stdout and stderr for golden comparison
			output := stdout.String()
			if stderr.Len() > 0 {
				if len(output) > 0 {
					output += "\n--- stderr ---\n"
				}
				output += stderr.String()
			}
			if err != nil && len(output) == 0 {
				// If there's an error but no output captured, use the error message
				output = "Error: " + err.Error() + "\n"
			}

			// Normalize paths for consistent comparison
			output = normalizePaths(output)

			// Store/compare golden file
			g.AssertString(tt.name, output)
		})
	}
}

// TestAdoptCommand_Verification tests that adopted files are correctly structured
func TestAdoptCommand_Verification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping verification tests in short mode")
	}

	t.Run("verify_package_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		packageDir := filepath.Join(tmpDir, "packages")

		os.MkdirAll(targetDir, 0755)
		os.MkdirAll(packageDir, 0755)

		// Create file to adopt
		testFile := filepath.Join(targetDir, ".testfile")
		testContent := "test content\n"
		os.WriteFile(testFile, []byte(testContent), 0644)

		os.Setenv("HOME", tmpDir)
		defer os.Unsetenv("HOME")

		oldDir, _ := os.Getwd()
		os.Chdir(targetDir)
		defer os.Chdir(oldDir)

		// Execute adopt command
		var stdout, stderr bytes.Buffer
		rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"--target", targetDir, "--dir", packageDir, "adopt", ".testfile"})

		ctx := context.Background()
		_, err := executeCommand(ctx, rootCmd)
		require.NoError(t, err)

		// Verify package structure
		expectedPkgDir := filepath.Join(packageDir, "dot-testfile")
		require.DirExists(t, expectedPkgDir, "package directory should be created")

		expectedFile := filepath.Join(expectedPkgDir, "dot-testfile")
		require.FileExists(t, expectedFile, "adopted file should exist in package")

		// Verify content
		content, err := os.ReadFile(expectedFile)
		require.NoError(t, err)
		require.Equal(t, testContent, string(content), "file content should match")

		// Verify symlink was created
		symlinkPath := filepath.Join(targetDir, ".testfile")
		info, err := os.Lstat(symlinkPath)
		require.NoError(t, err)
		require.True(t, info.Mode()&os.ModeSymlink != 0, "file should be a symlink")
	})

	t.Run("verify_nested_file_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		packageDir := filepath.Join(tmpDir, "packages")

		os.MkdirAll(targetDir, 0755)
		os.MkdirAll(packageDir, 0755)

		// Create nested file to adopt
		configDir := filepath.Join(targetDir, ".config", "test")
		os.MkdirAll(configDir, 0755)
		testFile := filepath.Join(configDir, "config.yml")
		os.WriteFile(testFile, []byte("key: value\n"), 0644)

		os.Setenv("HOME", tmpDir)
		defer os.Unsetenv("HOME")

		oldDir, _ := os.Getwd()
		os.Chdir(targetDir)
		defer os.Chdir(oldDir)

		// Execute adopt command
		var stdout, stderr bytes.Buffer
		rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"--target", targetDir, "--dir", packageDir, "adopt", "test", ".config/test/config.yml"})

		ctx := context.Background()
		_, err := executeCommand(ctx, rootCmd)
		require.NoError(t, err)

		// Verify package structure with nested path (adopt uses base name only)
		expectedFile := filepath.Join(packageDir, "test", "config.yml")
		require.FileExists(t, expectedFile, "adopted nested file should exist in package")

		// Verify symlink at original location
		symlinkPath := filepath.Join(targetDir, ".config", "test", "config.yml")
		info, err := os.Lstat(symlinkPath)
		require.NoError(t, err)
		require.True(t, info.Mode()&os.ModeSymlink != 0, "file at original location should be a symlink")
	})
}
