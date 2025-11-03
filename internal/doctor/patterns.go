package doctor

import "path/filepath"

// PatternCategory describes type of symlink based on its target.
type PatternCategory struct {
	Name        string
	Description string
	Patterns    []string // Glob patterns for targets
	Confidence  string   // "high", "medium", "low"
}

// DefaultPatternCategories returns hardcoded system patterns.
func DefaultPatternCategories() []PatternCategory {
	return []PatternCategory{
		{
			Name:        "cargo",
			Description: "Rust/Cargo managed binaries",
			Patterns:    []string{"*/.cargo/bin/*", "*/cargo/bin/*"},
			Confidence:  "high",
		},
		{
			Name:        "npm",
			Description: "NPM/Node managed tools",
			Patterns:    []string{"*/.npm/*", "*/node_modules/*", "*/.nvm/*"},
			Confidence:  "high",
		},
		{
			Name:        "system",
			Description: "System package manager",
			Patterns:    []string{"/usr/bin/*", "/usr/local/bin/*", "/opt/*"},
			Confidence:  "high",
		},
		{
			Name:        "vscode",
			Description: "VSCode managed extensions",
			Patterns:    []string{"*/.vscode/*", "*/.vscode-server/*"},
			Confidence:  "high",
		},
		{
			Name:        "flatpak",
			Description: "Flatpak managed applications",
			Patterns:    []string{"*/.local/share/flatpak/*"},
			Confidence:  "high",
		},
		{
			Name:        "jetbrains",
			Description: "JetBrains IDE managed",
			Patterns:    []string{"*/.local/share/JetBrains/*"},
			Confidence:  "high",
		},
	}
}

// CategorizeSymlink returns category for a symlink target, or nil if unknown.
func CategorizeSymlink(target string, categories []PatternCategory) *PatternCategory {
	for i := range categories {
		for _, pattern := range categories[i].Patterns {
			// filepath.Match doesn't support ** or multiple path segments
			// Use simple substring matching for patterns with * prefix
			if len(pattern) > 2 && pattern[:2] == "*/" {
				// Pattern like "*/bin/*" - check if path contains this segment
				segments := pattern[2:] // Remove */
				if matchesPathSegment(target, segments) {
					return &categories[i]
				}
			} else {
				// Direct glob matching for simpler patterns
				matched, _ := filepath.Match(pattern, target)
				if matched {
					return &categories[i]
				}
			}
		}
	}
	return nil
}

// matchesPathSegment checks if a path contains the given segment pattern.
func matchesPathSegment(path, segment string) bool {
	// Convert to forward slashes for consistent matching
	path = filepath.ToSlash(path)
	segment = filepath.ToSlash(segment)

	// Check if segment appears anywhere in the path
	// For patterns like ".cargo/bin/*", we want to match "/home/user/.cargo/bin/rustup"
	parts := filepath.SplitList(path)
	if len(parts) == 0 {
		parts = []string{path}
	}

	// Simple contains check for segment
	return containsSegment(path, segment)
}

// containsSegment checks if the path contains the segment pattern.
func containsSegment(path, segment string) bool {
	// For a pattern like ".cargo/bin/*", check if ".cargo/bin/" is in the path
	if segment[len(segment)-1] == '*' {
		prefix := segment[:len(segment)-1]
		return containsSubstring(path, prefix)
	}
	return containsSubstring(path, segment)
}

// containsSubstring performs case-sensitive substring matching.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

// indexOfSubstring returns the index of substr in s, or -1 if not found.
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
