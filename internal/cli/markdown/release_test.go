package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/cli/render"
)

func TestParseReleaseSections_GitHub(t *testing.T) {
	// Typical GitHub release notes format
	input := `## What's Changed

### Features
- Add new clone command by @user in #123
- Improve status display

### Bug Fixes
- Fix config loading issue in #456
- Resolve symlink handling

### Breaking Changes
- Remove deprecated API

### Other
- Update dependencies`

	sections := ParseReleaseSections(input)

	assert.Len(t, sections.Features, 2)
	assert.Len(t, sections.Fixes, 2)
	assert.Len(t, sections.Breaking, 1)
	assert.Len(t, sections.Other, 1)

	assert.Contains(t, sections.Features[0], "clone command")
	assert.Contains(t, sections.Fixes[0], "config loading")
	assert.Contains(t, sections.Breaking[0], "deprecated API")
}

func TestParseReleaseSections_KeepAChangelog(t *testing.T) {
	// Keep a Changelog format
	input := `## [1.2.0] - 2024-01-15

### Added
- New feature one
- New feature two

### Fixed
- Bug fix one
- Bug fix two

### Changed
- Updated behavior`

	sections := ParseReleaseSections(input)

	assert.Len(t, sections.Features, 2) // "Added" maps to Features
	assert.Len(t, sections.Fixes, 2)    // "Fixed" maps to Fixes
	assert.Len(t, sections.Other, 1)    // "Changed" maps to Other
}

func TestParseReleaseSections_Empty(t *testing.T) {
	sections := ParseReleaseSections("")
	assert.True(t, sections.IsEmpty())

	sections = ParseReleaseSections("Just some text without structure")
	assert.True(t, sections.IsEmpty())
}

func TestParseReleaseSections_BulletVariants(t *testing.T) {
	input := `## Features
- Dash bullet
* Asterisk bullet
+ Plus bullet`

	sections := ParseReleaseSections(input)

	assert.Len(t, sections.Features, 3)
	assert.Contains(t, sections.Features[0], "Dash bullet")
	assert.Contains(t, sections.Features[1], "Asterisk bullet")
	assert.Contains(t, sections.Features[2], "Plus bullet")
}

func TestReleaseSections_Stats(t *testing.T) {
	sections := &ReleaseSections{
		Breaking: []string{"one"},
		Features: []string{"a", "b", "c"},
		Fixes:    []string{"x", "y"},
		Other:    []string{"z"},
	}

	stats := sections.Stats()

	assert.Equal(t, 1, stats.BreakingChanges)
	assert.Equal(t, 3, stats.Features)
	assert.Equal(t, 2, stats.Fixes)
	assert.Equal(t, 1, stats.Other)
	assert.Equal(t, 7, stats.TotalItems)
}

func TestReleaseSections_RenderSummary(t *testing.T) {
	c := render.NewColorizer(false)

	sections := &ReleaseSections{
		Breaking: []string{"one"},
		Features: []string{"a", "b"},
		Fixes:    []string{"x"},
	}

	summary := sections.RenderSummary(c)

	assert.Contains(t, summary, "Release Summary")
	assert.Contains(t, summary, "1 breaking")
	assert.Contains(t, summary, "2 new feature")
	assert.Contains(t, summary, "1 bug fix")
}

func TestReleaseSections_RenderSummary_Empty(t *testing.T) {
	c := render.NewColorizer(false)
	sections := &ReleaseSections{}

	summary := sections.RenderSummary(c)
	assert.Empty(t, summary)
}

func TestReleaseSections_RenderDetailed(t *testing.T) {
	c := render.NewColorizer(false)

	sections := &ReleaseSections{
		Breaking: []string{"Breaking change 1"},
		Features: []string{"Feature 1", "Feature 2"},
		Fixes:    []string{"Fix 1"},
	}

	detailed := sections.RenderDetailed(c, 10)

	assert.Contains(t, detailed, "Breaking Changes")
	assert.Contains(t, detailed, "Breaking change 1")
	assert.Contains(t, detailed, "New Features")
	assert.Contains(t, detailed, "Feature 1")
	assert.Contains(t, detailed, "Feature 2")
	assert.Contains(t, detailed, "Bug Fixes")
	assert.Contains(t, detailed, "Fix 1")
}

func TestReleaseSections_RenderDetailed_Truncation(t *testing.T) {
	c := render.NewColorizer(false)

	sections := &ReleaseSections{
		Features: []string{"F1", "F2", "F3", "F4", "F5", "F6"},
	}

	// Limit to 3
	detailed := sections.RenderDetailed(c, 3)

	assert.Contains(t, detailed, "F1")
	assert.Contains(t, detailed, "F2")
	assert.Contains(t, detailed, "F3")
	assert.NotContains(t, detailed, "F4")
	assert.Contains(t, detailed, "and 3 more")
}

func TestReleaseSections_RenderDetailed_DefaultLimit(t *testing.T) {
	c := render.NewColorizer(false)

	sections := &ReleaseSections{
		Features: []string{"F1"},
	}

	// Test with 0 limit (should default to 10)
	detailed := sections.RenderDetailed(c, 0)
	assert.Contains(t, detailed, "F1")
}

func TestParseReleaseSections_BreakingKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "breaking",
			input: "## Breaking Changes\n- Change one",
		},
		{
			name:  "incompatible",
			input: "## Incompatible Changes\n- Change one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := ParseReleaseSections(tt.input)
			require.Len(t, sections.Breaking, 1)
			assert.Contains(t, sections.Breaking[0], "Change one")
		})
	}
}

func TestParseReleaseSections_FeatureKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "features",
			input: "## Features\n- Item",
		},
		{
			name:  "added",
			input: "## Added\n- Item",
		},
		{
			name:  "new",
			input: "## New\n- Item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := ParseReleaseSections(tt.input)
			require.Len(t, sections.Features, 1)
			assert.Equal(t, "Item", sections.Features[0])
		})
	}
}

func TestParseReleaseSections_FixKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "fixes",
			input: "## Bug Fixes\n- Item",
		},
		{
			name:  "fixed",
			input: "## Fixed\n- Item",
		},
		{
			name:  "patch",
			input: "## Patch Notes\n- Item",
		},
		{
			name:  "resolved",
			input: "## Resolved Issues\n- Item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := ParseReleaseSections(tt.input)
			require.Len(t, sections.Fixes, 1)
			assert.Equal(t, "Item", sections.Fixes[0])
		})
	}
}

func TestRenderItem_Links(t *testing.T) {
	c := render.NewColorizer(false)

	tests := []struct {
		name     string
		input    string
		contains string
		excludes string
	}{
		{
			name:     "simple link",
			input:    "[text](http://example.com)",
			contains: "text",
			excludes: "http://",
		},
		{
			name:     "link with description",
			input:    "See [the docs](http://docs.example.com) for more",
			contains: "the docs",
			excludes: "docs.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderItem(tt.input, c)
			assert.Contains(t, result, tt.contains)
			if tt.excludes != "" {
				assert.NotContains(t, result, tt.excludes)
			}
		})
	}
}

func TestIsHeader(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"# Title", true},
		{"## Section", true},
		{"### Subsection", true},
		{"Normal text", false},
		{"", false},
		{"-item", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			assert.Equal(t, tt.expected, isHeader(tt.line))
		})
	}
}

func TestIsBulletItem(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"- Item", true},
		{"* Item", true},
		{"+ Item", true},
		{"Normal text", false},
		{"1. Numbered", false},
		{"-no space", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			assert.Equal(t, tt.expected, isBulletItem(tt.line))
		})
	}
}

func TestExtractBulletContent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"- Item text", "Item text"},
		{"* Item text", "Item text"},
		{"+ Item text", "Item text"},
		{"- Spaced   ", "Spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractBulletContent(tt.input))
		})
	}
}

func TestContainsAny(t *testing.T) {
	assert.True(t, containsAny("hello world", "world", "universe"))
	assert.True(t, containsAny("hello world", "hello"))
	assert.False(t, containsAny("hello world", "foo", "bar"))
	assert.False(t, containsAny("hello", "world"))
}
