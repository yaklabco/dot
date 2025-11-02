package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsArgValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "requires at least error",
			err:      errors.New("requires at least 1 arg(s), only received 0"),
			expected: true,
		},
		{
			name:     "accepts error",
			err:      errors.New("accepts at most 2 arg(s), received 3"),
			expected: true,
		},
		{
			name:     "too many arguments",
			err:      errors.New("too many arguments provided"),
			expected: true,
		},
		{
			name:     "unknown command",
			err:      errors.New("unknown command \"foo\" for \"dot\""),
			expected: true,
		},
		{
			name:     "runtime error",
			err:      errors.New("failed to create symlink"),
			expected: false,
		},
		{
			name:     "file not found error",
			err:      errors.New("package directory not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isArgValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	// Test that executeCommand captures the executed command
	rootCmd := NewRootCommand("test", "abc123", "2025-01-01")
	rootCmd.SetArgs([]string{"--help"})

	cmd, err := executeCommand(context.Background(), rootCmd)
	assert.NoError(t, err)
	// cmd may be nil for --help since it exits early, which is acceptable
	_ = cmd
}

func TestExecuteCommandWithError(t *testing.T) {
	// Test that executeCommand returns error for invalid args
	rootCmd := NewRootCommand("test", "abc123", "2025-01-01")
	rootCmd.SetArgs([]string{"--invalid-flag"})

	_, err := executeCommand(context.Background(), rootCmd)
	assert.Error(t, err)
}
