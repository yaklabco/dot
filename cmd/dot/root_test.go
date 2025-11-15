package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand_Version(t *testing.T) {
	rootCmd := NewRootCommand("1.0.0", "abc123", "2025-01-01")
	rootCmd.SetArgs([]string{"--version"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "1.0.0")
	require.Contains(t, out.String(), "abc123")
	require.Contains(t, out.String(), "2025-01-01")
}

func TestRootCommand_Help(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")
	rootCmd.SetArgs([]string{"--help"})

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)

	err := rootCmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "dot")
	require.Contains(t, out.String(), "manage")
	require.Contains(t, out.String(), "dotfile manager")
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")

	// Verify global flags exist
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("dir"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("target"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("dry-run"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("quiet"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("log-json"))
}

func TestRootCommand_ShortFlags(t *testing.T) {
	rootCmd := NewRootCommand("dev", "none", "unknown")

	// Verify short flag aliases exist
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("d"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("t"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("n"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("v"))
	require.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("q"))
}

func TestParseFileSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "bytes",
			input:    "100B",
			expected: 100,
			wantErr:  false,
		},
		{
			name:     "bytes lowercase",
			input:    "50b",
			expected: 50,
			wantErr:  false,
		},
		{
			name:     "kilobytes",
			input:    "10KB",
			expected: 10240,
			wantErr:  false,
		},
		{
			name:     "kilobytes lowercase",
			input:    "10kb",
			expected: 10240,
			wantErr:  false,
		},
		{
			name:     "kilobytes short",
			input:    "10K",
			expected: 10240,
			wantErr:  false,
		},
		{
			name:     "kilobytes short lowercase",
			input:    "10k",
			expected: 10240,
			wantErr:  false,
		},
		{
			name:     "megabytes",
			input:    "5MB",
			expected: 5242880,
			wantErr:  false,
		},
		{
			name:     "megabytes lowercase",
			input:    "5mb",
			expected: 5242880,
			wantErr:  false,
		},
		{
			name:     "megabytes short",
			input:    "5M",
			expected: 5242880,
			wantErr:  false,
		},
		{
			name:     "gigabytes",
			input:    "2GB",
			expected: 2147483648,
			wantErr:  false,
		},
		{
			name:     "gigabytes lowercase",
			input:    "2gb",
			expected: 2147483648,
			wantErr:  false,
		},
		{
			name:     "gigabytes short",
			input:    "2G",
			expected: 2147483648,
			wantErr:  false,
		},
		{
			name:     "terabytes",
			input:    "1TB",
			expected: 1099511627776,
			wantErr:  false,
		},
		{
			name:     "terabytes lowercase",
			input:    "1tb",
			expected: 1099511627776,
			wantErr:  false,
		},
		{
			name:     "terabytes short",
			input:    "1T",
			expected: 1099511627776,
			wantErr:  false,
		},
		{
			name:     "decimal value",
			input:    "1.5MB",
			expected: 1572864,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid unit",
			input:    "100XX",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "missing unit",
			input:    "100",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFileSize(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
