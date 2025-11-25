package markdown

import (
	"regexp"
	"strings"

	"github.com/yaklabco/dot/internal/cli/render"
)

// Renderer converts Markdown to terminal-friendly output.
type Renderer struct {
	colorizer *render.Colorizer
	maxWidth  int
}

// NewRenderer creates a Markdown renderer.
func NewRenderer(colorizer *render.Colorizer, maxWidth int) *Renderer {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	return &Renderer{
		colorizer: colorizer,
		maxWidth:  maxWidth,
	}
}

// Render converts Markdown text to terminal output.
func (r *Renderer) Render(markdown string) string {
	var result strings.Builder
	lines := strings.Split(markdown, "\n")
	inCodeBlock := false

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			result.WriteString("    ")
			result.WriteString(r.colorizer.Dim(line))
			result.WriteString("\n")
			continue
		}

		rendered := r.renderLine(line)
		result.WriteString(rendered)
		result.WriteString("\n")
	}

	return result.String()
}

// RenderLines converts Markdown text and returns as slice of lines.
func (r *Renderer) RenderLines(markdown string) []string {
	rendered := r.Render(markdown)
	return strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
}

func (r *Renderer) renderLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// Empty lines
	if trimmed == "" {
		return ""
	}

	// Headers: ## Title -> bold
	if strings.HasPrefix(trimmed, "## ") {
		return r.colorizer.Bold(strings.TrimPrefix(trimmed, "## "))
	}
	if strings.HasPrefix(trimmed, "### ") {
		return r.colorizer.Bold(strings.TrimPrefix(trimmed, "### "))
	}
	if strings.HasPrefix(trimmed, "# ") {
		return r.colorizer.Bold(strings.TrimPrefix(trimmed, "# "))
	}

	// Horizontal rules
	if trimmed == "---" || trimmed == "***" || trimmed == "___" {
		return r.colorizer.Dim(strings.Repeat("-", 40))
	}

	// Bullet points: - item or * item
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
		prefix := strings.Repeat("  ", indent/2)
		return prefix + r.colorizer.Dim("*") + " " + r.renderInline(content)
	}

	// Numbered lists: 1. item
	if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); matched {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		// Extract number and content
		parts := strings.SplitN(trimmed, ". ", 2)
		if len(parts) == 2 {
			prefix := strings.Repeat("  ", indent/2)
			return prefix + r.colorizer.Dim(parts[0]+".") + " " + r.renderInline(parts[1])
		}
	}

	// Blockquotes: > text
	if strings.HasPrefix(trimmed, "> ") {
		content := strings.TrimPrefix(trimmed, "> ")
		return r.colorizer.Dim("|") + " " + r.colorizer.Dim(r.renderInline(content))
	}

	return r.renderInline(line)
}

func (r *Renderer) renderInline(text string) string {
	// Bold: **text** -> bold
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.Trim(match, "*")
		return r.colorizer.Bold(inner)
	})

	// Code: `code` -> accent
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := strings.Trim(match, "`")
		return r.colorizer.Accent(inner)
	})

	// Links: [text](url) -> text
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRe.ReplaceAllString(text, "$1")

	// PR/Issue references: #123 -> accent
	refRe := regexp.MustCompile(`#(\d+)`)
	text = refRe.ReplaceAllStringFunc(text, func(match string) string {
		return r.colorizer.Accent(match)
	})

	// User mentions: @username -> accent
	mentionRe := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	text = mentionRe.ReplaceAllStringFunc(text, func(match string) string {
		return r.colorizer.Dim(match)
	})

	return text
}
