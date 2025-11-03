package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestNewListCommand(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Contains(t, cmd.Use, "list")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestListCommand_Flags(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	// Check that format flag exists (default is text for clean output)
	formatFlag := cmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag)
	assert.Equal(t, "text", formatFlag.DefValue)

	// Check that color flag exists
	colorFlag := cmd.Flags().Lookup("color")
	require.NotNil(t, colorFlag)
	assert.Equal(t, "auto", colorFlag.DefValue)

	// Check that sort flag exists
	sortFlag := cmd.Flags().Lookup("sort")
	require.NotNil(t, sortFlag)
	assert.Equal(t, "name", sortFlag.DefValue)
}

func TestListCommand_Help(t *testing.T) {
	cfg := &dot.Config{}
	cmd := NewListCommand(cfg)

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}

func TestRenderCleanList(t *testing.T) {
	tests := []struct {
		name       string
		packages   []dot.PackageInfo
		packageDir string
		wantOutput string
	}{
		{
			name:       "empty package list",
			packages:   []dot.PackageInfo{},
			packageDir: "/home/user/dotfiles",
			wantOutput: "No packages installed\n",
		},
		{
			name: "single package",
			packages: []dot.PackageInfo{
				{Name: "vim", LinkCount: 1, InstalledAt: time.Now().Add(-1 * time.Hour)},
			},
			packageDir: "/home/user/dotfiles",
			wantOutput: "Packages: 1 package in /home/user/dotfiles\n\nvim  (1 link)  installed 1 hour ago\n",
		},
		{
			name: "multiple packages with alignment",
			packages: []dot.PackageInfo{
				{Name: "vim", LinkCount: 1, InstalledAt: time.Now().Add(-1 * time.Hour)},
				{Name: "dot-ssh", LinkCount: 5, InstalledAt: time.Now().Add(-2 * time.Hour)},
			},
			packageDir: "/home/user/dotfiles",
			wantOutput: "Packages: 2 packages in /home/user/dotfiles\n\nvim      (1 link)   installed 1 hour ago\ndot-ssh  (5 links)  installed 2 hours ago\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderCleanList(&buf, tt.packages, tt.packageDir, false)
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now, "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2 hours ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"3 days ago", now.Add(-72 * time.Hour), "3 days ago"},
		{"1 week ago", now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"1 month ago", now.Add(-30 * 24 * time.Hour), "1 month ago"},
		{"3 months ago", now.Add(-90 * 24 * time.Hour), "3 months ago"},
		{"1 year ago", now.Add(-365 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", now.Add(-730 * 24 * time.Hour), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}
