package install

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor_Defaults(t *testing.T) {
	e := NewExecutor()
	assert.NotNil(t, e)
	assert.Equal(t, DefaultTimeout, e.timeout)
	assert.False(t, e.dryRun)
}

func TestNewExecutor_WithOptions(t *testing.T) {
	var stdout, stderr bytes.Buffer

	e := NewExecutor(
		WithStdout(&stdout),
		WithStderr(&stderr),
		WithDryRun(true),
		WithTimeout(30*time.Second),
	)

	assert.Equal(t, &stdout, e.stdout)
	assert.Equal(t, &stderr, e.stderr)
	assert.True(t, e.dryRun)
	assert.Equal(t, 30*time.Second, e.timeout)
}

func TestExecutor_Execute_NilCommand(t *testing.T) {
	e := NewExecutor()
	_, err := e.Execute(context.Background(), nil)
	assert.Error(t, err)
}

func TestExecutor_Execute_DryRun(t *testing.T) {
	e := NewExecutor(WithDryRun(true))

	cmd, err := NewCommand(SourceHomebrew, "dot")
	require.NoError(t, err)

	output, err := e.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "[dry-run]")
	assert.Contains(t, output, "brew upgrade dot")
}

func TestExecutor_Execute_RealCommand(t *testing.T) {
	var stdout bytes.Buffer
	e := NewExecutor(
		WithStdout(&stdout),
		WithTimeout(5*time.Second),
	)

	// Use echo as a safe test command
	// Note: This test actually executes echo
	cmd := &Command{
		name:   "echo",
		args:   []string{"hello", "world"},
		source: SourceManual, // Just for testing
	}

	output, err := e.Execute(context.Background(), cmd)

	require.NoError(t, err)
	assert.Contains(t, output, "hello world")
}

func TestExecutor_ExecuteCapture_NilCommand(t *testing.T) {
	e := NewExecutor()
	_, _, err := e.ExecuteCapture(context.Background(), nil)
	assert.Error(t, err)
}

func TestExecutor_ExecuteCapture_DryRun(t *testing.T) {
	e := NewExecutor(WithDryRun(true))

	cmd, err := NewCommand(SourceApt, "dot")
	require.NoError(t, err)

	stdout, stderr, err := e.ExecuteCapture(context.Background(), cmd)

	require.NoError(t, err)
	assert.Contains(t, stdout, "[dry-run]")
	assert.Empty(t, stderr)
}

func TestExecutor_Execute_Timeout(t *testing.T) {
	e := NewExecutor(WithTimeout(10 * time.Millisecond))

	// Create a command that would hang (but we use very short timeout)
	cmd := &Command{
		name:   "sleep",
		args:   []string{"10"},
		source: SourceManual,
	}

	_, err := e.Execute(context.Background(), cmd)

	// Should error due to timeout
	assert.Error(t, err)
}
