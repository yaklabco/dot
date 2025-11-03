package doctor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPatternCategories(t *testing.T) {
	categories := DefaultPatternCategories()

	assert.NotEmpty(t, categories)
	assert.GreaterOrEqual(t, len(categories), 6) // At least cargo, npm, system, vscode, flatpak, jetbrains

	// Verify each category has required fields
	for _, cat := range categories {
		assert.NotEmpty(t, cat.Name, "Category name should not be empty")
		assert.NotEmpty(t, cat.Description, "Category description should not be empty")
		assert.NotEmpty(t, cat.Patterns, "Category patterns should not be empty")
		assert.NotEmpty(t, cat.Confidence, "Category confidence should not be empty")
	}
}

func TestCategorizeSymlink(t *testing.T) {
	categories := DefaultPatternCategories()

	tests := []struct {
		name     string
		target   string
		wantName string
		wantNil  bool
	}{
		{
			name:     "cargo binary",
			target:   "/home/user/.cargo/bin/rustup",
			wantName: "cargo",
		},
		{
			name:     "npm module",
			target:   "/home/user/.npm/bin/eslint",
			wantName: "npm",
		},
		{
			name:     "node_modules",
			target:   "/home/user/project/node_modules/bin/webpack",
			wantName: "npm",
		},
		{
			name:     "system bin",
			target:   "/usr/bin/python",
			wantName: "system",
		},
		{
			name:     "system local bin",
			target:   "/usr/local/bin/git",
			wantName: "system",
		},
		{
			name:     "vscode extension",
			target:   "/home/user/.vscode/extensions/something",
			wantName: "vscode",
		},
		{
			name:     "vscode server",
			target:   "/home/user/.vscode-server/data/something",
			wantName: "vscode",
		},
		{
			name:     "flatpak",
			target:   "/home/user/.local/share/flatpak/app/something",
			wantName: "flatpak",
		},
		{
			name:     "jetbrains",
			target:   "/home/user/.local/share/JetBrains/IdeaIC/something",
			wantName: "jetbrains",
		},
		{
			name:    "uncategorized",
			target:  "/home/user/custom/path/file",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeSymlink(tt.target, categories)

			if tt.wantNil {
				assert.Nil(t, result, "Expected no category match for %s", tt.target)
			} else {
				assert.NotNil(t, result, "Expected category match for %s", tt.target)
				if result != nil {
					assert.Equal(t, tt.wantName, result.Name, "Wrong category for %s", tt.target)
				}
			}
		})
	}
}

func TestCategorizeSymlink_EmptyCategories(t *testing.T) {
	result := CategorizeSymlink("/any/path", []PatternCategory{})
	assert.Nil(t, result)
}

func TestCategorizeSymlink_MultipleMatches(t *testing.T) {
	// Create categories where multiple patterns could match
	categories := []PatternCategory{
		{
			Name:        "first",
			Description: "First match",
			Patterns:    []string{"*/bin/*"},
			Confidence:  "high",
		},
		{
			Name:        "second",
			Description: "Second match",
			Patterns:    []string{"*/bin/*"},
			Confidence:  "high",
		},
	}

	result := CategorizeSymlink("/home/user/bin/tool", categories)
	// Should return first match
	if assert.NotNil(t, result) {
		assert.Equal(t, "first", result.Name)
	}
}
