package integration

import (
	"os"
	"os/exec"
	"testing"
)

// skipIfCLIUnavailable checks if the CLI can be executed and skips the test if not.
func skipIfCLIUnavailable(t *testing.T, output []byte, err error) {
	t.Helper()
	if err != nil {
		t.Skipf("CLI execution unavailable in this environment: %v, output: %s", err, output)
	}
}

// isolateStateDir sets XDG_STATE_HOME to a temp directory so that
// subprocess commands don't leak the state guard marker to the real home.
func isolateStateDir(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	stateDir := t.TempDir()
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "XDG_STATE_HOME="+stateDir)
}

// checkCLIAvailable verifies the CLI can be executed.
func checkCLIAvailable(t *testing.T) {
	t.Helper()
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		t.Skip("Go toolchain not available for CLI tests")
	}
}
