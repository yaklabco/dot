package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesainslie/dot/internal/cli/golden"
	"github.com/stretchr/testify/require"
)

// TestManageCommand_Golden tests the manage command with various scenarios
func TestManageCommand_Golden(t *testing.T) {
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
			name: "manage_single_package",
			args: []string{"manage", "vim"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create a simple package
				vimPkg := filepath.Join(packageDir, "vim")
				os.MkdirAll(vimPkg, 0755)
				os.WriteFile(filepath.Join(vimPkg, "dot-vimrc"), []byte("\" Vim config\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_multiple_packages",
			args: []string{"manage", "vim", "bash", "git"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create multiple packages
				for _, pkg := range []string{"vim", "bash", "git"} {
					pkgDir := filepath.Join(packageDir, pkg)
					os.MkdirAll(pkgDir, 0755)
					os.WriteFile(filepath.Join(pkgDir, "dot-"+pkg+"rc"), []byte("# "+pkg+" config\n"), 0644)
				}

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_package_with_nested_files",
			args: []string{"manage", "config"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package with nested structure
				configPkg := filepath.Join(packageDir, "config")
				os.MkdirAll(filepath.Join(configPkg, "dot-config-app"), 0755)
				os.WriteFile(filepath.Join(configPkg, "dot-config-app", "config.yml"), []byte("key: value\n"), 0644)
				os.WriteFile(filepath.Join(configPkg, "dot-config-app", "settings.json"), []byte("{}\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_with_dry_run",
			args: []string{"manage", "--dry-run", "vim"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create a simple package
				vimPkg := filepath.Join(packageDir, "vim")
				os.MkdirAll(vimPkg, 0755)
				os.WriteFile(filepath.Join(vimPkg, "dot-vimrc"), []byte("\" Vim config\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_with_verbose",
			args: []string{"manage", "-v", "bash"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create a simple package
				bashPkg := filepath.Join(packageDir, "bash")
				os.MkdirAll(bashPkg, 0755)
				os.WriteFile(filepath.Join(bashPkg, "dot-bashrc"), []byte("# Bash config\n"), 0644)
				os.WriteFile(filepath.Join(bashPkg, "dot-bash_profile"), []byte("# Bash profile\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_package_with_hidden_dirs",
			args: []string{"manage", "ssh"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package with hidden directory structure
				sshPkg := filepath.Join(packageDir, "ssh")
				os.MkdirAll(filepath.Join(sshPkg, "dot-ssh"), 0755)
				os.WriteFile(filepath.Join(sshPkg, "dot-ssh", "config"), []byte("Host *\n"), 0600)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_already_managed_package",
			args: []string{"manage", "vim"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package
				vimPkg := filepath.Join(packageDir, "vim")
				os.MkdirAll(vimPkg, 0755)
				vimrcPath := filepath.Join(vimPkg, "dot-vimrc")
				os.WriteFile(vimrcPath, []byte("\" Vim config\n"), 0644)

				// Create symlink (already managed)
				targetFile := filepath.Join(targetDir, ".vimrc")
				os.Symlink(vimrcPath, targetFile)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_package_with_executable",
			args: []string{"manage", "scripts"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package with executable file
				scriptsPkg := filepath.Join(packageDir, "scripts")
				os.MkdirAll(scriptsPkg, 0755)
				scriptPath := filepath.Join(scriptsPkg, "dot-local-bin-myscript")
				os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'Hello'\n"), 0755)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},
		{
			name: "manage_package_with_multiple_hidden_files",
			args: []string{"manage", "dotfiles"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package with multiple hidden files
				dotfilesPkg := filepath.Join(packageDir, "dotfiles")
				os.MkdirAll(dotfilesPkg, 0755)
				os.WriteFile(filepath.Join(dotfilesPkg, "dot-bashrc"), []byte("# bashrc\n"), 0644)
				os.WriteFile(filepath.Join(dotfilesPkg, "dot-zshrc"), []byte("# zshrc\n"), 0644)
				os.WriteFile(filepath.Join(dotfilesPkg, "dot-profile"), []byte("# profile\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false,
		},

		// Error scenarios
		{
			name:        "manage_no_arguments",
			args:        []string{"manage"},
			setupFunc:   func(t *testing.T) (string, string, func()) { return "", "", func() {} },
			expectError: true,
		},
		{
			name: "manage_nonexistent_package",
			args: []string{"manage", "nonexistent"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: true,
		},
		{
			name: "manage_empty_package",
			args: []string{"manage", "empty"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create empty package directory
				emptyPkg := filepath.Join(packageDir, "empty")
				os.MkdirAll(emptyPkg, 0755)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: true, // Empty packages produce "cannot execute empty plan" error
		},
		{
			name: "manage_package_with_conflict",
			args: []string{"manage", "vim"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages")

				os.MkdirAll(targetDir, 0755)
				os.MkdirAll(packageDir, 0755)

				// Create package
				vimPkg := filepath.Join(packageDir, "vim")
				os.MkdirAll(vimPkg, 0755)
				os.WriteFile(filepath.Join(vimPkg, "dot-vimrc"), []byte("\" Package vimrc\n"), 0644)

				// Create conflicting file in target
				os.WriteFile(filepath.Join(targetDir, ".vimrc"), []byte("\" Existing vimrc\n"), 0644)

				os.Setenv("HOME", tmpDir)

				return targetDir, packageDir, func() {
					os.Unsetenv("HOME")
				}
			},
			expectError: false, // Manage succeeds even with existing files (just skips them)
		},
		{
			name: "manage_with_invalid_package_dir",
			args: []string{"manage", "--dir", "/nonexistent/path", "vim"},
			setupFunc: func(t *testing.T) (string, string, func()) {
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return "", "", func() { os.Unsetenv("HOME") }
			},
			expectError: true,
		},
	}

	// Get current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Ensure testdata directory exists
	testdataDir := filepath.Join("testdata", "golden", "manage")
	os.MkdirAll(testdataDir, 0755)

	g := golden.New(t, "manage")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			targetDir, packageDir, cleanup := tt.setupFunc(t)
			defer cleanup()

			// Ensure we restore working directory
			defer func() {
				os.Chdir(originalWd)
			}()

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

// TestManageCommand_Verification tests that managed packages create correct symlinks
func TestManageCommand_Verification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping verification tests in short mode")
	}

	t.Run("verify_symlink_creation", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		packageDir := filepath.Join(tmpDir, "packages")

		os.MkdirAll(targetDir, 0755)
		os.MkdirAll(packageDir, 0755)

		// Create package
		vimPkg := filepath.Join(packageDir, "vim")
		os.MkdirAll(vimPkg, 0755)
		vimrcPath := filepath.Join(vimPkg, "dot-vimrc")
		testContent := "\" Vim config\nset number\n"
		os.WriteFile(vimrcPath, []byte(testContent), 0644)

		os.Setenv("HOME", tmpDir)
		defer os.Unsetenv("HOME")

		// Execute manage command
		var stdout, stderr bytes.Buffer
		rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"--target", targetDir, "--dir", packageDir, "manage", "vim"})

		ctx := context.Background()
		_, err := executeCommand(ctx, rootCmd)
		require.NoError(t, err)

		// Verify symlink was created
		symlinkPath := filepath.Join(targetDir, ".vimrc")
		info, err := os.Lstat(symlinkPath)
		require.NoError(t, err)
		require.True(t, info.Mode()&os.ModeSymlink != 0, "should be a symlink")

		// Verify symlink points to correct location
		linkTarget, err := os.Readlink(symlinkPath)
		require.NoError(t, err)
		require.Equal(t, vimrcPath, linkTarget, "symlink should point to package file")

		// Verify content is accessible through symlink
		content, err := os.ReadFile(symlinkPath)
		require.NoError(t, err)
		require.Equal(t, testContent, string(content), "content should be accessible")
	})

	t.Run("verify_nested_directory_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		packageDir := filepath.Join(tmpDir, "packages")

		os.MkdirAll(targetDir, 0755)
		os.MkdirAll(packageDir, 0755)

		// Create package with nested structure
		configPkg := filepath.Join(packageDir, "config")
		nestedDir := filepath.Join(configPkg, "dot-config-app")
		os.MkdirAll(nestedDir, 0755)
		os.WriteFile(filepath.Join(nestedDir, "config.yml"), []byte("key: value\n"), 0644)

		os.Setenv("HOME", tmpDir)
		defer os.Unsetenv("HOME")

		// Execute manage command
		var stdout, stderr bytes.Buffer
		rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"--target", targetDir, "--dir", packageDir, "manage", "config"})

		ctx := context.Background()
		_, err := executeCommand(ctx, rootCmd)
		require.NoError(t, err)

		// Verify nested directory structure was created in target
		expectedDir := filepath.Join(targetDir, ".config", "app")
		require.DirExists(t, expectedDir, "nested directory should be created")

		// Verify file symlink in nested directory
		expectedFile := filepath.Join(expectedDir, "config.yml")
		info, err := os.Lstat(expectedFile)
		require.NoError(t, err)
		require.True(t, info.Mode()&os.ModeSymlink != 0, "nested file should be a symlink")
	})

	t.Run("verify_multiple_packages", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		packageDir := filepath.Join(tmpDir, "packages")

		os.MkdirAll(targetDir, 0755)
		os.MkdirAll(packageDir, 0755)

		// Create multiple packages
		packages := []string{"vim", "bash", "git"}
		for _, pkg := range packages {
			pkgDir := filepath.Join(packageDir, pkg)
			os.MkdirAll(pkgDir, 0755)
			os.WriteFile(filepath.Join(pkgDir, "dot-"+pkg+"rc"), []byte("# "+pkg+"\n"), 0644)
		}

		os.Setenv("HOME", tmpDir)
		defer os.Unsetenv("HOME")

		// Execute manage command with multiple packages
		var stdout, stderr bytes.Buffer
		rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
		rootCmd.SetOut(&stdout)
		rootCmd.SetErr(&stderr)
		rootCmd.SetArgs([]string{"--target", targetDir, "--dir", packageDir, "manage", "vim", "bash", "git"})

		ctx := context.Background()
		_, err := executeCommand(ctx, rootCmd)
		require.NoError(t, err)

		// Verify all symlinks were created
		for _, pkg := range packages {
			symlinkPath := filepath.Join(targetDir, "."+pkg+"rc")
			info, err := os.Lstat(symlinkPath)
			require.NoError(t, err, "symlink should exist for %s", pkg)
			require.True(t, info.Mode()&os.ModeSymlink != 0, "%s should be a symlink", pkg)
		}
	})
}
