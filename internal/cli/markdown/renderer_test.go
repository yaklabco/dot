package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/cli/render"
)

func TestRenderer_RenderHeaders(t *testing.T) {
	c := render.NewColorizer(false) // No color for easier testing
	r := NewRenderer(c, 80)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "h1 header",
			input:    "# Main Title",
			contains: "Main Title",
		},
		{
			name:     "h2 header",
			input:    "## Section",
			contains: "Section",
		},
		{
			name:     "h3 header",
			input:    "### Subsection",
			contains: "Subsection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Render(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestRenderer_RenderBulletLists(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "dash bullet",
			input:    "- Item one",
			contains: "Item one",
		},
		{
			name:     "asterisk bullet",
			input:    "* Item two",
			contains: "Item two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Render(tt.input)
			assert.Contains(t, result, tt.contains)
			assert.Contains(t, result, "*") // Marker should be present
		})
	}
}

func TestRenderer_RenderCodeBlocks(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "```go\nfunc main() {}\n```"
	result := r.Render(input)

	assert.Contains(t, result, "func main()")
}

func TestRenderer_RenderInlineElements(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	tests := []struct {
		name     string
		input    string
		contains string
		excludes string
	}{
		{
			name:     "bold text",
			input:    "This is **bold** text",
			contains: "bold",
			excludes: "**",
		},
		{
			name:     "inline code",
			input:    "Run `command` here",
			contains: "command",
		},
		{
			name:     "link",
			input:    "See [docs](https://example.com)",
			contains: "docs",
			excludes: "https://",
		},
		{
			name:     "issue reference",
			input:    "Fixes #123",
			contains: "#123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Render(tt.input)
			assert.Contains(t, result, tt.contains)
			if tt.excludes != "" {
				assert.NotContains(t, result, tt.excludes)
			}
		})
	}
}

func TestRenderer_RenderLines(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "Line 1\nLine 2\nLine 3"
	lines := r.RenderLines(input)

	require.Len(t, lines, 3)
	assert.Equal(t, "Line 1", lines[0])
	assert.Equal(t, "Line 2", lines[1])
	assert.Equal(t, "Line 3", lines[2])
}

func TestRenderer_HorizontalRule(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	tests := []string{"---", "***", "___"}
	for _, hr := range tests {
		t.Run(hr, func(t *testing.T) {
			result := r.Render(hr)
			// Should contain dashes (the rendered horizontal rule)
			assert.Contains(t, result, "-")
		})
	}
}

func TestRenderer_Blockquote(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "> This is a quote"
	result := r.Render(input)

	assert.Contains(t, result, "This is a quote")
	assert.Contains(t, result, "|") // Quote marker
}

func TestRenderer_NumberedList(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "1. First item\n2. Second item"
	result := r.Render(input)

	assert.Contains(t, result, "First item")
	assert.Contains(t, result, "Second item")
	assert.Contains(t, result, "1.")
	assert.Contains(t, result, "2.")
}

func TestNewRenderer_DefaultWidth(t *testing.T) {
	c := render.NewColorizer(false)

	// Test with zero width
	r := NewRenderer(c, 0)
	assert.Equal(t, 80, r.maxWidth)

	// Test with negative width
	r = NewRenderer(c, -10)
	assert.Equal(t, 80, r.maxWidth)

	// Test with explicit width
	r = NewRenderer(c, 120)
	assert.Equal(t, 120, r.maxWidth)
}

func TestRenderer_EmptyInput(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	result := r.Render("")
	assert.Equal(t, "\n", result)
}

func TestRenderer_ComplexDocument(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := `# Release v1.2.0

## Features

- Added new command
- Improved performance by **50%**

## Bug Fixes

- Fixed issue #123
- Resolved [memory leak](https://example.com/issue)

## Breaking Changes

- Removed deprecated API`

	result := r.Render(input)

	// All content should be present
	assert.Contains(t, result, "Release v1.2.0")
	assert.Contains(t, result, "Features")
	assert.Contains(t, result, "Added new command")
	assert.Contains(t, result, "50%")
	assert.Contains(t, result, "Bug Fixes")
	assert.Contains(t, result, "#123")
	assert.Contains(t, result, "memory leak")
	assert.Contains(t, result, "Breaking Changes")

	// Markdown syntax should be processed
	assert.NotContains(t, result, "**50%**")
	assert.NotContains(t, result, "[memory leak]")
}

func TestRenderer_UserMentions(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "Thanks to @username for the contribution"
	result := r.Render(input)

	assert.Contains(t, result, "@username")
}

func TestRenderer_PreservesIndentation(t *testing.T) {
	c := render.NewColorizer(false)
	r := NewRenderer(c, 80)

	input := "- Top level\n  - Nested item"
	result := r.Render(input)

	lines := strings.Split(result, "\n")
	require.GreaterOrEqual(t, len(lines), 2)

	// Nested item should have more indentation
	// Both lines contain the marker *
	topLine := lines[0]
	nestedLine := lines[1]

	topIndent := len(topLine) - len(strings.TrimLeft(topLine, " "))
	nestedIndent := len(nestedLine) - len(strings.TrimLeft(nestedLine, " "))

	assert.Greater(t, nestedIndent, topIndent, "nested item should have more indentation")
}
