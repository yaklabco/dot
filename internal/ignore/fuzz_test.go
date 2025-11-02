package ignore

import (
	"testing"
)

// FuzzGlobToRegex tests glob pattern to regex conversion with random input.
// Run with: go test -fuzz=FuzzGlobToRegex -fuzztime=30s
func FuzzGlobToRegex(f *testing.F) {
	// Seed corpus with common glob patterns
	f.Add("*.log")
	f.Add(".git")
	f.Add("**/node_modules/**")
	f.Add("*.{js,ts}")
	f.Add("[0-9]*")
	f.Add("file?.txt")
	f.Add("**/.DS_Store")

	// Seed with potentially problematic input
	f.Add("[")
	f.Add("]")
	f.Add("[]")
	f.Add("[abc")
	f.Add("***")
	f.Add("???")
	f.Add("[[[[")
	f.Add("]]]]")
	f.Add(string(make([]byte, 1000)))
	f.Add("\x00\x01\x02")

	f.Fuzz(func(t *testing.T, pattern string) {
		// Should not panic on any input
		_ = GlobToRegex(pattern)
	})
}

// FuzzPatternMatch tests pattern matching with random input.
func FuzzPatternMatch(f *testing.F) {
	// Seed corpus with pattern and path combinations
	f.Add("*.log", "/var/log/app.log")
	f.Add(".git", ".git")
	f.Add(".git", "project/.git")
	f.Add("**/node_modules/**", "project/node_modules/package/index.js")
	f.Add("*.txt", "file.txt")
	f.Add("file?.txt", "file1.txt")

	// Seed with potentially problematic input
	f.Add("", "")
	f.Add("*", string(make([]byte, 10000)))
	f.Add("[", "/path/to/file")
	f.Add("***", "...")
	f.Add("\x00", "\x00\x01")

	f.Fuzz(func(t *testing.T, pattern, path string) {
		// Should not panic on any input
		result := NewPattern(pattern)
		if result.IsOk() {
			p := result.Unwrap()
			_ = p.Match(path)
			_ = p.MatchBasename(path)
		}
	})
}

// FuzzIgnoreSetShouldIgnore tests ignore set matching with random input.
func FuzzIgnoreSetShouldIgnore(f *testing.F) {
	// Seed corpus
	f.Add("*.log", "/var/log/app.log")
	f.Add(".git", ".git/config")
	f.Add("node_modules", "project/node_modules/package")

	// Seed with potentially problematic input
	f.Add("", "")
	f.Add("[[[", string(make([]byte, 1000)))
	f.Add("\x00\x01", "\x00\x01\x02")

	f.Fuzz(func(t *testing.T, pattern, path string) {
		// Create ignore set with random pattern
		ignoreSet := NewIgnoreSet()
		patternResult := NewPattern(pattern)
		if patternResult.IsOk() {
			p := patternResult.Unwrap()
			ignoreSet.AddPattern(p)

			// Should not panic on any path input
			_ = ignoreSet.ShouldIgnore(path)
		}
	})
}
