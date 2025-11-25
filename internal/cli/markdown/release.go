package markdown

import (
	"fmt"
	"strings"

	"github.com/yaklabco/dot/internal/cli/render"
)

// ReleaseStats contains counts of items in release notes.
type ReleaseStats struct {
	BreakingChanges int
	Features        int
	Fixes           int
	Other           int
	TotalItems      int
}

// ReleaseSections contains parsed release note sections.
type ReleaseSections struct {
	Breaking []string
	Features []string
	Fixes    []string
	Other    []string
	Raw      string
}

// ParseReleaseSections parses GitHub release notes into categorized sections.
func ParseReleaseSections(body string) *ReleaseSections {
	sections := &ReleaseSections{Raw: body}
	var currentSection *[]string

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Detect section headers
		if isHeader(trimmed) {
			switch {
			case containsAny(lower, "breaking", "incompatible"):
				currentSection = &sections.Breaking
				continue
			case containsAny(lower, "feature", "added", "new"):
				currentSection = &sections.Features
				continue
			case containsAny(lower, "fix", "bug", "patch", "resolved"):
				currentSection = &sections.Fixes
				continue
			case containsAny(lower, "change", "update", "improve", "refactor", "other"):
				currentSection = &sections.Other
				continue
			default:
				// Unknown section header - continue collecting to current section
				continue
			}
		}

		// Collect bullet items
		if isBulletItem(trimmed) && currentSection != nil {
			item := extractBulletContent(trimmed)
			if item != "" {
				*currentSection = append(*currentSection, item)
			}
		}
	}

	return sections
}

// Stats returns statistics about the release sections.
func (s *ReleaseSections) Stats() ReleaseStats {
	return ReleaseStats{
		BreakingChanges: len(s.Breaking),
		Features:        len(s.Features),
		Fixes:           len(s.Fixes),
		Other:           len(s.Other),
		TotalItems:      len(s.Breaking) + len(s.Features) + len(s.Fixes) + len(s.Other),
	}
}

// IsEmpty returns true if no sections were parsed.
func (s *ReleaseSections) IsEmpty() bool {
	return len(s.Breaking) == 0 && len(s.Features) == 0 &&
		len(s.Fixes) == 0 && len(s.Other) == 0
}

// RenderSummary renders a compact summary of the release notes.
func (s *ReleaseSections) RenderSummary(c *render.Colorizer) string {
	var result strings.Builder
	stats := s.Stats()

	if stats.TotalItems == 0 {
		return ""
	}

	result.WriteString(c.Bold("Release Summary:"))
	result.WriteString("\n")

	if stats.BreakingChanges > 0 {
		result.WriteString(fmt.Sprintf("  %s %d breaking change(s)\n",
			c.Warning("!"), stats.BreakingChanges))
	}
	if stats.Features > 0 {
		result.WriteString(fmt.Sprintf("  %s %d new feature(s)\n",
			c.Success("+"), stats.Features))
	}
	if stats.Fixes > 0 {
		result.WriteString(fmt.Sprintf("  %s %d bug fix(es)\n",
			c.Info("~"), stats.Fixes))
	}
	if stats.Other > 0 {
		result.WriteString(fmt.Sprintf("  %s %d other change(s)\n",
			c.Dim("*"), stats.Other))
	}

	return result.String()
}

// RenderDetailed renders the full categorized release notes.
func (s *ReleaseSections) RenderDetailed(c *render.Colorizer, maxPerSection int) string {
	var result strings.Builder

	if maxPerSection <= 0 {
		maxPerSection = 10
	}

	// Breaking changes first with warning styling
	if len(s.Breaking) > 0 {
		result.WriteString("\n")
		result.WriteString(c.Warning("Breaking Changes:"))
		result.WriteString("\n")
		for i, item := range s.Breaking {
			if i >= maxPerSection {
				result.WriteString(fmt.Sprintf("  %s\n",
					c.Dim(fmt.Sprintf("... and %d more", len(s.Breaking)-maxPerSection))))
				break
			}
			result.WriteString(fmt.Sprintf("  %s %s\n", c.Warning("!"), renderItem(item, c)))
		}
	}

	// Features
	if len(s.Features) > 0 {
		result.WriteString("\n")
		result.WriteString(c.Bold("New Features:"))
		result.WriteString("\n")
		for i, item := range s.Features {
			if i >= maxPerSection {
				result.WriteString(fmt.Sprintf("  %s\n",
					c.Dim(fmt.Sprintf("... and %d more", len(s.Features)-maxPerSection))))
				break
			}
			result.WriteString(fmt.Sprintf("  %s %s\n", c.Success("+"), renderItem(item, c)))
		}
	}

	// Bug fixes
	if len(s.Fixes) > 0 {
		result.WriteString("\n")
		result.WriteString(c.Bold("Bug Fixes:"))
		result.WriteString("\n")
		for i, item := range s.Fixes {
			if i >= maxPerSection {
				result.WriteString(fmt.Sprintf("  %s\n",
					c.Dim(fmt.Sprintf("... and %d more", len(s.Fixes)-maxPerSection))))
				break
			}
			result.WriteString(fmt.Sprintf("  %s %s\n", c.Info("~"), renderItem(item, c)))
		}
	}

	// Other changes
	if len(s.Other) > 0 {
		result.WriteString("\n")
		result.WriteString(c.Bold("Other Changes:"))
		result.WriteString("\n")
		for i, item := range s.Other {
			if i >= maxPerSection {
				result.WriteString(fmt.Sprintf("  %s\n",
					c.Dim(fmt.Sprintf("... and %d more", len(s.Other)-maxPerSection))))
				break
			}
			result.WriteString(fmt.Sprintf("  %s %s\n", c.Dim("*"), renderItem(item, c)))
		}
	}

	return result.String()
}

// Helper functions

func isHeader(line string) bool {
	return strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "##") ||
		strings.HasPrefix(line, "###")
}

func isBulletItem(line string) bool {
	return strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "* ") ||
		strings.HasPrefix(line, "+ ")
}

func extractBulletContent(line string) string {
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")
	line = strings.TrimPrefix(line, "+ ")
	return strings.TrimSpace(line)
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func renderItem(item string, c *render.Colorizer) string {
	// Render inline code
	codeRe := strings.NewReplacer("`", "")
	// Simple inline rendering - could use full Renderer for more features
	result := item

	// Remove markdown links but keep text
	for {
		linkStart := strings.Index(result, "[")
		if linkStart < 0 {
			break
		}

		linkEnd := strings.Index(result[linkStart:], "]")
		urlStart := strings.Index(result[linkStart:], "(")
		urlEnd := strings.Index(result[linkStart:], ")")

		if linkEnd > 0 && urlStart == linkEnd+1 && urlEnd > urlStart {
			// Extract link text
			text := result[linkStart+1 : linkStart+linkEnd]
			// Replace [text](url) with just text
			result = result[:linkStart] + text + result[linkStart+urlEnd+1:]
			continue
		}

		// If we get here, the pattern does not match a full [text](url)
		// Avoid potential infinite loops by stopping.
		break
	}

	// Clean up backticks for display
	result = codeRe.Replace(result)

	return result
}
