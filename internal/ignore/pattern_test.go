package ignore_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "simple pattern",
			pattern: "*.txt",
			wantErr: false,
		},
		{
			name:    "exact match",
			pattern: ".git",
			wantErr: false,
		},
		{
			name:    "unclosed bracket treated as literal",
			pattern: "[invalid",
			wantErr: false, // Glob converter escapes to literal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignore.NewPattern(tt.pattern)
			if tt.wantErr {
				assert.True(t, result.IsErr())
			} else {
				require.True(t, result.IsOk())
				pattern := result.Unwrap()
				assert.NotNil(t, pattern)
			}
		})
	}
}

func TestPattern_Match(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  ".git",
			path:     ".git",
			expected: true,
		},
		{
			name:     "exact match no match",
			pattern:  ".git",
			path:     ".gitignore",
			expected: false,
		},
		{
			name:     "glob star",
			pattern:  "*.txt",
			path:     "file.txt",
			expected: true,
		},
		{
			name:     "glob star no match",
			pattern:  "*.txt",
			path:     "file.md",
			expected: false,
		},
		{
			name:     "glob in middle",
			pattern:  "test*.txt",
			path:     "test123.txt",
			expected: true,
		},
		{
			name:     "question mark",
			pattern:  "file?.txt",
			path:     "file1.txt",
			expected: true,
		},
		{
			name:     "directory pattern",
			pattern:  "*/.git",
			path:     "project/.git",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := ignore.NewPattern(tt.pattern).Unwrap()
			matches := pattern.Match(tt.path)
			assert.Equal(t, tt.expected, matches, "pattern=%s, path=%s", tt.pattern, tt.path)
		})
	}
}

func TestPattern_MatchBasename(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{
			name:     "basename match",
			pattern:  ".DS_Store",
			path:     "dir/.DS_Store",
			expected: true,
		},
		{
			name:     "basename match nested",
			pattern:  ".DS_Store",
			path:     "a/b/c/.DS_Store",
			expected: true,
		},
		{
			name:     "basename no match",
			pattern:  ".DS_Store",
			path:     "file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := ignore.NewPattern(tt.pattern).Unwrap()
			matches := pattern.MatchBasename(tt.path)
			assert.Equal(t, tt.expected, matches)
		})
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		name     string
		glob     string
		testPath string
		expected bool
	}{
		{
			name:     "star matches anything",
			glob:     "*.txt",
			testPath: "file.txt",
			expected: true,
		},
		{
			name:     "question matches single char",
			glob:     "file?.txt",
			testPath: "file1.txt",
			expected: true,
		},
		{
			name:     "literal dot",
			glob:     ".git",
			testPath: ".git",
			expected: true,
		},
		{
			name:     "escape special chars",
			glob:     "test[1].txt",
			testPath: "test[1].txt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := ignore.GlobToRegex(tt.glob)
			pattern := ignore.NewPatternFromRegex(regex).Unwrap()
			matches := pattern.Match(tt.testPath)
			assert.Equal(t, tt.expected, matches)
		})
	}
}

func TestPattern_CaseSensitive(t *testing.T) {
	pattern := ignore.NewPattern("Test.txt").Unwrap()

	assert.True(t, pattern.Match("Test.txt"))
	assert.False(t, pattern.Match("test.txt"))
	assert.False(t, pattern.Match("TEST.TXT"))
}

func TestPattern_String(t *testing.T) {
	pattern := ignore.NewPattern("*.txt").Unwrap()
	str := pattern.String()
	assert.Contains(t, str, "*.txt")
}

func TestPattern_Negation(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		wantNegated bool
	}{
		{
			name:        "normal pattern",
			pattern:     "*.txt",
			wantNegated: false,
		},
		{
			name:        "negation pattern",
			pattern:     "!important.txt",
			wantNegated: true,
		},
		{
			name:        "negation with wildcard",
			pattern:     "!*.keep",
			wantNegated: true,
		},
		{
			name:        "directory pattern",
			pattern:     "node_modules/",
			wantNegated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignore.NewPattern(tt.pattern)
			require.True(t, result.IsOk())
			pattern := result.Unwrap()

			assert.Equal(t, tt.wantNegated, pattern.IsNegation())

			// Verify the original pattern string is preserved
			assert.Equal(t, tt.pattern, pattern.String())
		})
	}
}

func TestPattern_NegationMatching(t *testing.T) {
	// Test that negation patterns match the same paths as normal patterns
	// (just with different interpretation)
	tests := []struct {
		name        string
		pattern     string
		testPath    string
		shouldMatch bool
	}{
		{
			name:        "negation matches file",
			pattern:     "!important.txt",
			testPath:    "important.txt",
			shouldMatch: true,
		},
		{
			name:        "negation matches with wildcard",
			pattern:     "!*.keep",
			testPath:    "cache.keep",
			shouldMatch: true,
		},
		{
			name:        "negation doesn't match unrelated",
			pattern:     "!*.keep",
			testPath:    "file.txt",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignore.NewPattern(tt.pattern)
			require.True(t, result.IsOk())
			pattern := result.Unwrap()

			matched := pattern.Match(tt.testPath) || pattern.MatchBasename(tt.testPath)
			assert.Equal(t, tt.shouldMatch, matched)
		})
	}
}
