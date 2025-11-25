package install

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// DefaultTimeout is the default command execution timeout.
const DefaultTimeout = 5 * time.Minute

// Executor executes validated commands securely.
type Executor struct {
	stdout  io.Writer
	stderr  io.Writer
	stdin   io.Reader
	dryRun  bool
	timeout time.Duration
}

// ExecutorOption configures an Executor.
type ExecutorOption func(*Executor)

// WithStdout sets the stdout writer.
func WithStdout(w io.Writer) ExecutorOption {
	return func(e *Executor) {
		e.stdout = w
	}
}

// WithStderr sets the stderr writer.
func WithStderr(w io.Writer) ExecutorOption {
	return func(e *Executor) {
		e.stderr = w
	}
}

// WithStdin sets the stdin reader.
func WithStdin(r io.Reader) ExecutorOption {
	return func(e *Executor) {
		e.stdin = r
	}
}

// WithDryRun enables dry-run mode where commands are not executed.
func WithDryRun(dryRun bool) ExecutorOption {
	return func(e *Executor) {
		e.dryRun = dryRun
	}
}

// WithTimeout sets the command execution timeout.
func WithTimeout(d time.Duration) ExecutorOption {
	return func(e *Executor) {
		e.timeout = d
	}
}

// NewExecutor creates a new command executor.
func NewExecutor(opts ...ExecutorOption) *Executor {
	e := &Executor{
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		stdin:   os.Stdin,
		timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Execute runs the validated command.
func (e *Executor) Execute(ctx context.Context, cmd *Command) (string, error) {
	if cmd == nil {
		return "", fmt.Errorf("nil command")
	}

	// In dry-run mode, just return the command that would be executed
	if e.dryRun {
		return fmt.Sprintf("[dry-run] would execute: %s", cmd.String()), nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build the exec.Cmd directly - no shell involved
	//nolint:gosec // Arguments are validated by Command construction
	execCmd := exec.CommandContext(ctx, cmd.Name(), cmd.Args()...)

	// Capture output
	var stdout, stderr bytes.Buffer

	// Set up output handling
	if e.stdout != nil {
		execCmd.Stdout = io.MultiWriter(&stdout, e.stdout)
	} else {
		execCmd.Stdout = &stdout
	}

	if e.stderr != nil {
		execCmd.Stderr = io.MultiWriter(&stderr, e.stderr)
	} else {
		execCmd.Stderr = &stderr
	}

	// Connect stdin for interactive commands (like sudo)
	if e.stdin != nil {
		execCmd.Stdin = e.stdin
	}

	// Execute the command
	err := execCmd.Run()
	output := stdout.String()

	if err != nil {
		// Include stderr in error message if available
		errOutput := stderr.String()
		if errOutput != "" {
			return output, fmt.Errorf("command failed: %w: %s", err, errOutput)
		}
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

// ExecuteCapture executes a command and captures all output without streaming.
func (e *Executor) ExecuteCapture(ctx context.Context, cmd *Command) (stdout, stderr string, err error) {
	if cmd == nil {
		return "", "", fmt.Errorf("nil command")
	}

	// In dry-run mode, just return the command that would be executed
	if e.dryRun {
		return fmt.Sprintf("[dry-run] would execute: %s", cmd.String()), "", nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build the exec.Cmd directly - no shell involved
	//nolint:gosec // Arguments are validated by Command construction
	execCmd := exec.CommandContext(ctx, cmd.Name(), cmd.Args()...)

	var stdoutBuf, stderrBuf bytes.Buffer
	execCmd.Stdout = &stdoutBuf
	execCmd.Stderr = &stderrBuf

	err = execCmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}
