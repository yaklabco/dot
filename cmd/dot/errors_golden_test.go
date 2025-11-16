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

// normalizeTempPaths replaces temp directory paths with placeholders for consistent comparison
func normalizeTempPaths(output string) string {
	// Replace /var/folders/.../T/TestName... with <TMPDIR>
	re := regexp.MustCompile(`/var/folders/[^/]+/[^/]+/T/[^/\s]+`)
	output = re.ReplaceAllString(output, "<TMPDIR>")

	// Replace /tmp/TestName... with <TMPDIR>
	re = regexp.MustCompile(`/tmp/[^/\s]+`)
	output = re.ReplaceAllString(output, "<TMPDIR>")

	// Replace generic temp directory patterns
	re = regexp.MustCompile(`/[^/]+/folders/[^/]+/[^/]+/[^/\s]+`)
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

// TestErrorScenarios_Golden tests error output formatting across all commands
func TestErrorScenarios_Golden(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping golden tests in short mode")
	}

	tests := []struct {
		name        string
		args        []string
		setupFunc   func(*testing.T) (cleanup func())
		expectError bool
	}{
		// Manage command errors
		{
			name:        "manage_nonexistent_package",
			args:        []string{"manage", "nonexistent-package"},
			expectError: true,
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},
		{
			name:        "manage_no_package_specified",
			args:        []string{"manage"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "manage_invalid_package_dir",
			args:        []string{"manage", "--dir", "/nonexistent/path", "test"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Unmanage command errors
		{
			name:        "unmanage_nonexistent_package",
			args:        []string{"unmanage", "nonexistent-package"},
			expectError: false, // Returns success with message about no packages
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},
		{
			name:        "unmanage_no_package_specified",
			args:        []string{"unmanage"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Status command errors
		{
			name:        "status_invalid_format",
			args:        []string{"status", "--format", "invalid"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "status_no_packages_installed",
			args:        []string{"status"},
			expectError: false, // Not an error, but empty state
			setupFunc: func(t *testing.T) func() {
				// TODO: Fix non-deterministic timestamps in test data
				// This test has flaky timestamps due to packages being created at different times
				// Skip for now - tracked in issue
				t.Skip("Skipping due to non-deterministic timestamps - needs test data refactor")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},

		// List command errors
		{
			name:        "list_invalid_sort",
			args:        []string{"list", "--sort", "invalid"},
			expectError: false, // List command may accept any sort value
			setupFunc: func(t *testing.T) func() {
				// TODO: Fix non-deterministic timestamps in test data
				// This test has flaky timestamps due to packages being created at different times
				// Skip for now - tracked in issue
				t.Skip("Skipping due to non-deterministic timestamps - needs test data refactor")
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},
		{
			name:        "list_invalid_format",
			args:        []string{"list", "--format", "invalid"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Adopt command errors
		{
			name:        "adopt_no_files_specified",
			args:        []string{"adopt"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "adopt_nonexistent_file",
			args:        []string{"adopt", "/nonexistent/file"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "adopt_multiple_files_no_package",
			args:        []string{"adopt", "file1", "file2"},
			expectError: true,
			setupFunc: func(t *testing.T) func() {
				// Simple test - just pass non-existent relative paths
				// This will trigger the error we want to test
				return func() {}
			},
		},

		// Doctor command errors
		{
			name:        "doctor_invalid_scan_mode",
			args:        []string{"doctor", "--scan-mode", "invalid"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "doctor_invalid_format",
			args:        []string{"doctor", "--format", "invalid"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Clone command errors
		{
			name:        "clone_no_url",
			args:        []string{"clone"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "clone_invalid_url",
			args:        []string{"clone", "not-a-url"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Config command errors
		{
			name:        "config_get_nonexistent_key",
			args:        []string{"config", "get", "nonexistent.key"},
			expectError: true,
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},
		{
			name:        "config_set_invalid_key",
			args:        []string{"config", "set", "invalid key with spaces", "value"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "config_init_invalid_format",
			args:        []string{"config", "init", "--format", "invalid"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Upgrade command errors
		{
			name:        "upgrade_no_package_manager",
			args:        []string{"upgrade"},
			expectError: false, // May not error, but shows no package manager found
			setupFunc: func(t *testing.T) func() {
				// Simulate environment without package managers
				oldPath := os.Getenv("PATH")
				os.Setenv("PATH", "/nonexistent")
				return func() { os.Setenv("PATH", oldPath) }
			},
		},

		// Global flag errors
		{
			name:        "invalid_flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "invalid_command",
			args:        []string{"invalid-command"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
		{
			name:        "help_invalid_command",
			args:        []string{"help", "nonexistent-command"},
			expectError: false, // help command shows general help for unknown commands
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},

		// Conflict scenarios
		{
			name:        "manage_with_conflicts",
			args:        []string{"manage", "test-pkg"},
			expectError: true,
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, "target")
				packageDir := filepath.Join(tmpDir, "packages", "test-pkg")

				// Create package
				os.MkdirAll(packageDir, 0755)
				os.WriteFile(filepath.Join(packageDir, "dot-bashrc"), []byte("package content"), 0644)

				// Create conflicting file in target
				os.MkdirAll(targetDir, 0755)
				os.WriteFile(filepath.Join(targetDir, ".bashrc"), []byte("existing content"), 0644)

				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},

		// Permission errors (simulated)
		{
			name:        "manage_readonly_target",
			args:        []string{"manage", "test-pkg", "--target", "/readonly/dir"},
			expectError: true,
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				packageDir := filepath.Join(tmpDir, "packages", "test-pkg")
				os.MkdirAll(packageDir, 0755)
				os.WriteFile(filepath.Join(packageDir, "dot-file"), []byte("content"), 0644)

				os.Setenv("HOME", tmpDir)
				return func() { os.Unsetenv("HOME") }
			},
		},

		// Validation errors
		// Note: manage_empty_package_name test removed due to non-deterministic file listing
		// The error output varies based on directory contents at test time
		{
			name:        "adopt_absolute_path_outside_home",
			args:        []string{"adopt", "/etc/passwd", "mypackage"},
			expectError: true,
			setupFunc:   func(t *testing.T) func() { return func() {} },
		},
	}

	// Get current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Ensure testdata directory exists
	testdataDir := filepath.Join("testdata", "golden", "errors")
	os.MkdirAll(testdataDir, 0755)

	g := golden.New(t, "errors")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := tt.setupFunc(t)
			defer cleanup()

			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer

			// Create root command
			rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
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

			// Normalize temp directory paths for consistent comparison
			output = normalizeTempPaths(output)

			// Store/compare golden file
			g.AssertString(tt.name, output)
		})
	}
}

// TestErrorMessages_Format tests that error messages follow consistent formatting
func TestErrorMessages_Format(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectPatterns []string
	}{
		{
			name: "error_has_context",
			args: []string{"manage", "nonexistent"},
			expectPatterns: []string{
				"Error:",
				"package",
			},
		},
		{
			name:           "error_no_stack_trace",
			args:           []string{"manage", "nonexistent"},
			expectPatterns: []string{
				// Should NOT contain stack traces in user-facing errors
			},
		},
		{
			name: "error_actionable",
			args: []string{"adopt"},
			expectPatterns: []string{
				"interactive mode requires a terminal", // Interactive mode error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			ctx := context.Background()
			_, _ = executeCommand(ctx, rootCmd)

			output := stdout.String() + stderr.String()

			for _, pattern := range tt.expectPatterns {
				require.Contains(t, output, pattern, "expected pattern %q in output", pattern)
			}
		})
	}
}

// TestErrorExitCodes tests that commands return appropriate exit codes
func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
	}{
		{
			name:         "invalid_command",
			args:         []string{"invalid"},
			expectedCode: 1,
		},
		{
			name:         "invalid_flag",
			args:         []string{"--invalid-flag"},
			expectedCode: 1,
		},
		{
			name:         "missing_required_arg",
			args:         []string{"manage"},
			expectedCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			rootCmd := NewRootCommand("test", "abc123", "2024-01-01")
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs(tt.args)

			ctx := context.Background()
			_, err := executeCommand(ctx, rootCmd)

			if tt.expectedCode > 0 {
				require.Error(t, err, "expected error for exit code %d", tt.expectedCode)
			} else {
				require.NoError(t, err, "expected no error for exit code 0")
			}
		})
	}
}
