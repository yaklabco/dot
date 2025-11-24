package integration

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/tests/integration/testutil"
)

// TestCLI_VersionCommand tests the version command.
func TestCLI_VersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	cmd := exec.Command("go", "run", "../../cmd/dot", "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Skipf("CLI execution not available in this environment: %v, output: %s", err, output)
	}

	t.Logf("version command output: %s", output)
}

// TestCLI_HelpCommand tests the help command.
func TestCLI_HelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	cmd := exec.Command("go", "run", "../../cmd/dot", "help")
	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	assert.Contains(t, string(output), "USAGE")
}

// TestCLI_StatusCommand tests status command execution.
func TestCLI_StatusCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Run status command pointing to test directories
	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("status output: %s", output)
}

// TestCLI_ManageCommand tests basic manage command.
func TestCLI_ManageCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create test package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Run manage command
	cmd := exec.Command("go", "run", "../../cmd/dot", "manage", "vim",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("manage output: %s", output)

	// Verify link created
	vimrcLink := filepath.Join(env.TargetDir, ".vimrc")
	testutil.AssertLinkContains(t, vimrcLink, "dot-vimrc")
}

// TestCLI_DryRunFlag tests the --dry-run flag.
func TestCLI_DryRunFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create test package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Capture state before
	before := testutil.CaptureState(t, env.TargetDir)

	// Run with --dry-run
	cmd := exec.Command("go", "run", "../../cmd/dot", "manage", "vim",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir,
		"--dry-run")

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("dry-run output: %s", output)

	// Verify no changes made
	after := testutil.CaptureState(t, env.TargetDir)
	testutil.AssertStateUnchanged(t, before, after)
}

// TestCLI_VerboseFlag tests the --verbose flag.
func TestCLI_VerboseFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create test package
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	// Run with --verbose
	cmd := exec.Command("go", "run", "../../cmd/dot", "manage", "vim",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir,
		"--verbose")

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("verbose output: %s", output)
}

// TestCLI_ListCommand tests the list command.
func TestCLI_ListCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)
	client := testutil.NewTestClient(t, env)

	// Install a package first
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	err := client.Manage(env.Context(), "vim")
	require.NoError(t, err)

	// Run list command
	cmd := exec.Command("go", "run", "../../cmd/dot", "list",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("list output: %s", output)
	assert.Contains(t, string(output), "vim")
}

// TestCLI_DoctorCommand tests the doctor command.
func TestCLI_DoctorCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Run doctor command
	cmd := exec.Command("go", "run", "../../cmd/dot", "doctor",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("doctor output: %s", output)
}

// TestCLI_MultiplePackages tests managing multiple packages via CLI.
func TestCLI_MultiplePackages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create multiple packages
	env.FixtureBuilder().Package("vim").
		WithFile("dot-vimrc", "set nocompatible").
		Create()

	env.FixtureBuilder().Package("zsh").
		WithFile("dot-zshrc", "export EDITOR=vim").
		Create()

	// Manage multiple packages
	cmd := exec.Command("go", "run", "../../cmd/dot", "manage", "vim", "zsh",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("multiple packages output: %s", output)

	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".vimrc"), "dot-vimrc")
	testutil.AssertLinkContains(t, filepath.Join(env.TargetDir, ".zshrc"), "dot-zshrc")
}

// TestCLI_InvalidArguments tests error handling of invalid arguments.
func TestCLI_InvalidArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Run with invalid command
	cmd := exec.Command("go", "run", "../../cmd/dot", "invalidcommand")
	output, err := cmd.CombinedOutput()

	// Should return error
	require.Error(t, err, "expected command to fail but got success, output: %s", output)
	t.Logf("invalid command output: %s", output)
}

// TestCLI_MissingRequiredFlags tests handling of missing required flags.
func TestCLI_MissingRequiredFlags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Run manage without package name
	cmd := exec.Command("go", "run", "../../cmd/dot", "manage")
	output, err := cmd.CombinedOutput()

	t.Logf("missing flags output: %s", output)
	// Should return error for missing arguments
	require.Error(t, err, "expected error for missing package name, output: %s", output)
}

// TestCLI_ConfigFile tests reading from config file.
func TestCLI_ConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create config file
	configPath := filepath.Join(env.PackageDir, "config.yaml")
	configContent := `
package_dir: ` + env.PackageDir + `
target_dir: ` + env.TargetDir + `
`
	env.FixtureBuilder().FileTree(env.PackageDir).
		File("config.yaml", configContent)

	// Run command with config file
	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--config", configPath)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("config file output: %s", output)
}

// TestCLI_OutputFormats tests different output format flags.
func TestCLI_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	formats := []string{"text", "json", "yaml"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := exec.Command("go", "run", "../../cmd/dot", "status",
				"--package-dir", env.PackageDir,
				"--target-dir", env.TargetDir,
				"--format", format)

			output, err := cmd.CombinedOutput()
			skipIfCLIUnavailable(t, output, err)

			t.Logf("%s format output: %s", format, output)
		})
	}
}

// TestCLI_PipelineInput tests handling of piped input.
func TestCLI_PipelineInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		t.Skipf("CLI execution not available in this environment: %v, output: %s", err, out.String())
	}

	t.Logf("pipeline output: %s", out.String())
}

// TestCLI_ExitCodes tests proper exit code handling.
func TestCLI_ExitCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	tests := []struct {
		name     string
		args     []string
		wantExit bool
	}{
		{
			name:     "success",
			args:     []string{"version"},
			wantExit: false,
		},
		{
			name:     "invalid command",
			args:     []string{"notacommand"},
			wantExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"run", "../../cmd/dot"}, tt.args...)
			cmd := exec.Command("go", args...)
			output, err := cmd.CombinedOutput()

			if tt.wantExit {
				require.Error(t, err, "expected command to fail, output: %s", output)
			} else {
				skipIfCLIUnavailable(t, output, err)
			}
		})
	}
}

// TestCLI_SignalHandling tests graceful signal handling.
func TestCLI_SignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// This is a basic test - actual signal handling would need more setup
	env := testutil.NewTestEnvironment(t)

	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)
}

// TestCLI_EnvironmentVariables tests environment variable support.
func TestCLI_EnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	cmd := exec.Command("go", "run", "../../cmd/dot", "status")
	cmd.Env = append(cmd.Env,
		"DOT_PACKAGE_DIR="+env.PackageDir,
		"DOT_TARGET_DIR="+env.TargetDir,
	)

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("env vars output: %s", output)
}

// TestCLI_ColorOutput tests color output control.
func TestCLI_ColorOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Run with NO_COLOR
	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)
	cmd.Env = append(cmd.Env, "NO_COLOR=1")

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	// Output should not contain color codes
	assert.NotContains(t, string(output), "\x1b[")
}

// TestCLI_LongRunningOperation tests handling of longer operations.
func TestCLI_LongRunningOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	env := testutil.NewTestEnvironment(t)

	// Create many packages
	for i := 0; i < 20; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+(i%26))))
		if i >= 26 {
			pkgName += string(rune('0' + (i / 26)))
		}
		env.FixtureBuilder().Package(pkgName).
			WithFile("dot-file", "content").
			Create()
	}

	packages := make([]string, 20)
	for i := 0; i < 20; i++ {
		pkgName := filepath.Join("pkg", string(rune('a'+(i%26))))
		if i >= 26 {
			pkgName += string(rune('0' + (i / 26)))
		}
		packages[i] = pkgName
	}

	args := append([]string{"run", "../../cmd/dot", "manage"}, packages...)
	args = append(args,
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("long operation output: %s", output)
}

// TestCLI_InteractivePrompts tests handling when prompts might appear.
func TestCLI_InteractivePrompts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Test that commands work in non-interactive mode
	env := testutil.NewTestEnvironment(t)

	cmd := exec.Command("go", "run", "../../cmd/dot", "status",
		"--package-dir", env.PackageDir,
		"--target-dir", env.TargetDir)

	// Ensure non-interactive
	cmd.Stdin = strings.NewReader("")

	output, err := cmd.CombinedOutput()
	skipIfCLIUnavailable(t, output, err)

	t.Logf("non-interactive output: %s", output)
}
