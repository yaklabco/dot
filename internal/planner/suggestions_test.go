package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/domain"
)

// Task 7.3.3: Test Suggestion Generation
func TestGenerateSuggestionsForFileExists(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)
	assert.GreaterOrEqual(t, len(suggestions), 2)

	// Should suggest backup
	hasBackup := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "backup") {
			hasBackup = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasBackup, "Should suggest backup option")

	// Should suggest adopt
	hasAdopt := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "adopt") {
			hasAdopt = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasAdopt, "Should suggest adopt option")
}

func TestGenerateSuggestionsForWrongLink(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictWrongLink, targetPath, "Symlink points elsewhere")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should suggest unmanaging other package
	hasUnmanage := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "unmanage") {
			hasUnmanage = true
			assert.NotEmpty(t, s.Explanation)
		}
	}
	assert.True(t, hasUnmanage, "Should suggest unmanage option")
}

func TestGenerateSuggestionsForPermission(t *testing.T) {
	targetPath := domain.NewFilePath("/etc/config").Unwrap()
	conflict := NewConflict(ConflictPermission, targetPath, "Permission denied")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should mention checking permissions
	hasPermCheck := false
	for _, s := range suggestions {
		if containsIgnoreCase(s.Action, "permission") || containsIgnoreCase(s.Action, "access") {
			hasPermCheck = true
		}
	}
	assert.True(t, hasPermCheck, "Should suggest checking permissions")
}

func TestGenerateSuggestionsForCircular(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.config").Unwrap()
	conflict := NewConflict(ConflictCircular, targetPath, "Circular dependency")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should have actionable suggestions
	for _, s := range suggestions {
		assert.NotEmpty(t, s.Action)
		assert.NotEmpty(t, s.Explanation)
	}
}

func TestGenerateSuggestionsForTypeMismatch(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.config").Unwrap()
	conflict := NewConflict(ConflictFileExpected, targetPath, "File exists where directory expected")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)
}

// Task 7.3.4: Test Conflict Enrichment
func TestEnrichConflictWithSuggestions(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictFileExists, targetPath, "File exists")

	// Initially no suggestions
	assert.Empty(t, conflict.Suggestions)

	// Enrich with suggestions
	enriched := enrichConflictWithSuggestions(conflict)

	// Should now have suggestions
	assert.NotEmpty(t, enriched.Suggestions)
	assert.GreaterOrEqual(t, len(enriched.Suggestions), 2)

	// All suggestions should have required fields
	for _, s := range enriched.Suggestions {
		assert.NotEmpty(t, s.Action, "Suggestion should have action")
		assert.NotEmpty(t, s.Explanation, "Suggestion should have explanation")
	}
}

func TestEnrichMultipleConflicts(t *testing.T) {
	path1 := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict1 := NewConflict(ConflictFileExists, path1, "File exists")

	path2 := domain.NewFilePath("/home/user/.vimrc").Unwrap()
	conflict2 := NewConflict(ConflictWrongLink, path2, "Wrong link")

	enriched1 := enrichConflictWithSuggestions(conflict1)
	enriched2 := enrichConflictWithSuggestions(conflict2)

	// Both should have suggestions
	assert.NotEmpty(t, enriched1.Suggestions)
	assert.NotEmpty(t, enriched2.Suggestions)

	// Suggestions should be different for different conflict types
	assert.NotEqual(t, enriched1.Suggestions, enriched2.Suggestions)
}

// Additional coverage tests for suggestion generation edge cases
func TestGenerateSuggestionsForDirExpected(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.config").Unwrap()
	conflict := NewConflict(ConflictDirExpected, targetPath, "Directory expected but file found")

	suggestions := generateSuggestions(conflict)

	assert.NotEmpty(t, suggestions)

	// Should have actionable suggestions
	for _, s := range suggestions {
		assert.NotEmpty(t, s.Action)
		assert.NotEmpty(t, s.Explanation)
	}
}

func TestGenerateSuggestionsForUnknownType(t *testing.T) {
	targetPath := domain.NewFilePath("/home/user/.bashrc").Unwrap()
	conflict := NewConflict(ConflictType(999), targetPath, "Unknown conflict")

	suggestions := generateSuggestions(conflict)

	// Unknown conflict types should return empty suggestions
	assert.Empty(t, suggestions)
}

func TestGeneratePermissionSuggestionsWithRoot(t *testing.T) {
	// Test path at root level where Parent() might fail
	rootPath := domain.NewFilePath("/etc").Unwrap()
	conflict := NewConflict(ConflictPermission, rootPath, "Permission denied")

	suggestions := generatePermissionSuggestions(conflict)

	assert.NotEmpty(t, suggestions)
	assert.GreaterOrEqual(t, len(suggestions), 2)
}

func TestGenerateTypeMismatchBothDirections(t *testing.T) {
	t.Run("file expected", func(t *testing.T) {
		path := domain.NewFilePath("/home/user/.config").Unwrap()
		conflict := NewConflict(ConflictFileExpected, path, "File expected")

		suggestions := generateTypeMismatchSuggestions(conflict)

		assert.NotEmpty(t, suggestions)
		// Should mention removing directory
		found := false
		for _, s := range suggestions {
			if containsIgnoreCase(s.Action, "directory") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("dir expected", func(t *testing.T) {
		path := domain.NewFilePath("/home/user/.config").Unwrap()
		conflict := NewConflict(ConflictDirExpected, path, "Dir expected")

		suggestions := generateTypeMismatchSuggestions(conflict)

		assert.NotEmpty(t, suggestions)
		// Should mention removing file
		found := false
		for _, s := range suggestions {
			if containsIgnoreCase(s.Action, "file") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestAllSuggestionTemplatesHaveExamples(t *testing.T) {
	testCases := []struct {
		name           string
		conflictType   ConflictType
		minSuggestions int
	}{
		{"file exists", ConflictFileExists, 2},
		{"wrong link", ConflictWrongLink, 2},
		{"permission", ConflictPermission, 2},
		{"circular", ConflictCircular, 2},
		{"type mismatch", ConflictFileExpected, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := domain.NewFilePath("/home/user/test").Unwrap()
			conflict := NewConflict(tc.conflictType, path, "Test conflict")

			suggestions := generateSuggestions(conflict)

			assert.GreaterOrEqual(t, len(suggestions), tc.minSuggestions,
				"Should have at least %d suggestions", tc.minSuggestions)

			for i, s := range suggestions {
				assert.NotEmpty(t, s.Action, "Suggestion %d should have action", i)
				assert.NotEmpty(t, s.Explanation, "Suggestion %d should have explanation", i)
				// Example is optional, so we don't assert it
			}
		})
	}
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	// Simple ASCII lowercase
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
