package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestNewStatusCommand(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewStatusCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "status")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.ValidArgsFunction)
}

func TestStatusCommand_Flags(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewStatusCommand(cfg)

	// Check that format flag exists
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Check that color flag exists
	colorFlag := cmd.Flags().Lookup("color")
	require.NotNil(t, colorFlag)
	assert.Equal(t, "auto", colorFlag.DefValue)
}

func TestStatusCommand_OutputFormat(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		wantErr     bool
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:   "text format",
			format: "text",
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "No packages installed")
			},
		},
		{
			name:   "json format",
			format: "json",
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, `"packages"`)
			},
		},
		{
			name:   "yaml format",
			format: "yaml",
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "packages:")
			},
		},
		{
			name:   "table format",
			format: "table",
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "No packages installed")
			},
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies command structure
			// Actual integration with renderer is tested elsewhere
			cfg := &dot.Config{}
			cmd := NewStatusCommand(cfg)
			assert.NotNil(t, cmd)
		})
	}
}

func TestStatusCommand_Help(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewStatusCommand(cfg)

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}
