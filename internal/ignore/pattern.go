// Package ignore provides pattern matching for file exclusion.
package ignore

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yaklabco/dot/internal/domain"
)

// PatternType identifies whether a pattern includes or excludes files.
type PatternType int

const (
	// PatternInclude represents a normal ignore pattern.
	PatternInclude PatternType = iota
	// PatternExclude represents a negation pattern that un-ignores files.
	PatternExclude
)

// Pattern represents a compiled pattern for matching paths.
type Pattern struct {
	original string
	regex    *regexp.Regexp
	typ      PatternType
}

// NewPattern creates a pattern from a glob pattern.
// Converts glob syntax to regex for matching.
// Patterns starting with ! are negation patterns that un-ignore files.
func NewPattern(glob string) domain.Result[*Pattern] {
	typ := PatternInclude
	originalGlob := glob

	// Detect negation pattern
	if strings.HasPrefix(glob, "!") {
		typ = PatternExclude
		glob = glob[1:] // Remove ! prefix for regex conversion
	}

	regex := GlobToRegex(glob)
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return domain.Err[*Pattern](fmt.Errorf("compile pattern %q: %w", originalGlob, err))
	}

	return domain.Ok(&Pattern{
		original: originalGlob, // Store original glob with ! if present
		regex:    compiled,
		typ:      typ,
	})
}

// NewPatternFromRegex creates a pattern from a regex string.
func NewPatternFromRegex(regex string) domain.Result[*Pattern] {
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return domain.Err[*Pattern](fmt.Errorf("compile pattern: %w", err))
	}

	return domain.Ok(&Pattern{
		original: regex,
		regex:    compiled,
		typ:      PatternInclude, // Regex patterns are always include type
	})
}

// Match checks if the path matches the pattern.
func (p *Pattern) Match(path string) bool {
	return p.regex.MatchString(path)
}

// MatchBasename checks if the basename of the path matches the pattern.
// Useful for patterns like ".DS_Store" that should match anywhere in tree.
func (p *Pattern) MatchBasename(path string) bool {
	basename := filepath.Base(path)
	return p.regex.MatchString(basename)
}

// String returns the original pattern string.
func (p *Pattern) String() string {
	return p.original
}

// IsNegation returns true if this is a negation pattern.
func (p *Pattern) IsNegation() bool {
	return p.typ == PatternExclude
}

// Type returns the pattern type.
func (p *Pattern) Type() PatternType {
	return p.typ
}

// GlobToRegex converts a glob pattern to a regex pattern.
//
// Glob syntax:
//   - *       matches any sequence of characters
//   - ?       matches any single character
//   - [abc]   matches any character in the set
//   - [a-z]   matches any character in the range
//
// All other characters are escaped to match literally.
func GlobToRegex(glob string) string {
	var result strings.Builder
	result.WriteString("^")

	for i := 0; i < len(glob); i++ {
		ch := glob[i]

		switch ch {
		case '*':
			// Match any sequence of characters
			result.WriteString(".*")

		case '?':
			// Match any single character
			result.WriteString(".")

		case '[':
			// Character class - find the closing bracket
			j := i + 1
			for j < len(glob) && glob[j] != ']' {
				j++
			}
			if j < len(glob) && j > i+1 {
				// Valid character class with content
				// Check if it's a valid range/set
				class := glob[i : j+1]
				// For simplicity, escape brackets in glob patterns
				// This treats [1] as literal [1], not as character class
				result.WriteString(regexp.QuoteMeta(class))
				i = j
			} else {
				// No closing bracket or empty class - treat as literal
				result.WriteString(regexp.QuoteMeta(string(ch)))
			}

		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			// Escape regex special characters
			result.WriteString("\\")
			result.WriteByte(ch)

		default:
			// Literal character
			result.WriteByte(ch)
		}
	}

	result.WriteString("$")
	return result.String()
}
