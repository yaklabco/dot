package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/cli/render"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestReportUnmanageAllResults(t *testing.T) {
	tests := []struct {
		name   string
		count  int
		opts   dot.UnmanageOptions
		dryRun bool
	}{
		{
			name:   "dry run unmanage",
			count:  3,
			opts:   dot.UnmanageOptions{},
			dryRun: true,
		},
		{
			name:   "actual unmanage",
			count:  2,
			opts:   dot.UnmanageOptions{},
			dryRun: false,
		},
		{
			name:   "dry run purge",
			count:  1,
			opts:   dot.UnmanageOptions{Purge: true},
			dryRun: true,
		},
		{
			name:   "actual purge",
			count:  4,
			opts:   dot.UnmanageOptions{Purge: true},
			dryRun: false,
		},
		{
			name:   "dry run restore",
			count:  2,
			opts:   dot.UnmanageOptions{Restore: true},
			dryRun: true,
		},
		{
			name:   "actual restore",
			count:  1,
			opts:   dot.UnmanageOptions{Restore: true},
			dryRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output (function prints to stdout)
			// We just verify it doesn't panic
			reportUnmanageAllResults(tt.count, tt.opts, tt.dryRun)
		})
	}
}

func TestDisplayUnmanageAllSummary(t *testing.T) {
	packages := []dot.PackageInfo{
		{
			Name:   "vim",
			Source: "managed",
			Links:  []string{"/home/user/.vimrc", "/home/user/.vim"},
		},
		{
			Name:   "bash",
			Source: "adopted",
			Links:  []string{"/home/user/.bashrc"},
		},
	}
	opts := dot.UnmanageOptions{}
	packageDir := "/home/user/.dotfiles"

	// Function prints to stdout, just verify it doesn't panic
	displayUnmanageAllSummary(packages, opts, packageDir)
}

func TestIsTerminal(t *testing.T) {
	// Create a mock command with a bytes.Buffer as input
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetIn(buf)

	result := isTerminal(cmd)

	// bytes.Buffer is not a terminal
	assert.False(t, result, "bytes.Buffer should not be a terminal")
}

func TestRenderColorizer(t *testing.T) {
	// Test that colorizer functions work correctly
	c := render.NewColorizer(false)

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Bold", c.Bold, "test"},
		{"Dim", c.Dim, "test"},
		{"Accent", c.Accent, "test"},
		{"Error", c.Error, "test"},
		{"Success", c.Success, "test"},
		{"Info", c.Info, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			assert.NotEmpty(t, result, "colorizer function should return non-empty string")
		})
	}
}
