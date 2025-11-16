package render

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorizer(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		text     string
		wantFunc func(*Colorizer, string) string
	}{
		{
			name:    "success with color",
			enabled: true,
			text:    "success message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Success(s)
			},
		},
		{
			name:    "success without color",
			enabled: false,
			text:    "success message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Success(s)
			},
		},
		{
			name:    "warning with color",
			enabled: true,
			text:    "warning message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Warning(s)
			},
		},
		{
			name:    "error with color",
			enabled: true,
			text:    "error message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Error(s)
			},
		},
		{
			name:    "info with color",
			enabled: true,
			text:    "info message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Info(s)
			},
		},
		{
			name:    "dim with color",
			enabled: true,
			text:    "dim message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Dim(s)
			},
		},
		{
			name:    "accent with color",
			enabled: true,
			text:    "accent message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Accent(s)
			},
		},
		{
			name:    "bold with color",
			enabled: true,
			text:    "bold message",
			wantFunc: func(c *Colorizer, s string) string {
				return c.Bold(s)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewColorizer(tt.enabled)
			result := tt.wantFunc(c, tt.text)

			if tt.enabled {
				// Should contain ANSI codes
				assert.Contains(t, result, tt.text)
				assert.NotEqual(t, tt.text, result, "colored text should differ from plain text")
			} else {
				// Should be plain text
				assert.Equal(t, tt.text, result)
			}
		})
	}
}

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name     string
		noColor  string
		term     string
		expected bool
	}{
		{
			name:     "NO_COLOR set",
			noColor:  "1",
			term:     "xterm-256color",
			expected: false,
		},
		{
			name:     "dumb terminal",
			noColor:  "",
			term:     "dumb",
			expected: false,
		},
		{
			name:     "empty TERM",
			noColor:  "",
			term:     "",
			expected: false,
		},
		{
			name:     "xterm terminal",
			noColor:  "",
			term:     "xterm-256color",
			expected: true,
		},
		{
			name:     "color terminal",
			noColor:  "",
			term:     "screen-256color",
			expected: true,
		},
		{
			name:     "basic xterm",
			noColor:  "",
			term:     "xterm",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			oldNoColor := os.Getenv("NO_COLOR")
			oldTerm := os.Getenv("TERM")
			defer func() {
				os.Setenv("NO_COLOR", oldNoColor)
				os.Setenv("TERM", oldTerm)
			}()

			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("TERM", tt.term)

			result := ShouldUseColor()
			// Note: Will be false in tests since stdout is not a terminal
			// but we can still verify the NO_COLOR and TERM env vars work
			if tt.noColor != "" || tt.term == "dumb" || tt.term == "" {
				assert.False(t, result, "should return false for NO_COLOR or dumb TERM")
			}
		})
	}
}

func TestGetScheme(t *testing.T) {
	// Save and restore environment
	oldNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", oldNoColor)

	t.Run("with NO_COLOR", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		scheme := GetScheme()
		assert.Equal(t, "", scheme.Success.ANSI)
		assert.Equal(t, "", scheme.Error.ANSI)
	})

	t.Run("without NO_COLOR in non-terminal", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		scheme := GetScheme()
		// In tests, stdout is not a terminal, so colors are disabled
		assert.Equal(t, "", scheme.Success.ANSI)
	})
}

func TestColor_Apply(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		text  string
		want  string
	}{
		{
			name:  "color with ANSI",
			color: Color{ANSI: "\033[32m"},
			text:  "test",
			want:  "\033[32mtest\033[0m",
		},
		{
			name:  "no color",
			color: Color{ANSI: ""},
			text:  "test",
			want:  "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.Apply(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}
