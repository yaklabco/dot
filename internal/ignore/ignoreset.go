package ignore

// IgnoreSet is a collection of patterns for ignoring files.
type IgnoreSet struct {
	patterns []*Pattern
}

// NewIgnoreSet creates a new empty ignore set.
func NewIgnoreSet() *IgnoreSet {
	return &IgnoreSet{
		patterns: make([]*Pattern, 0),
	}
}

// NewDefaultIgnoreSet creates an ignore set with default patterns.
func NewDefaultIgnoreSet() *IgnoreSet {
	set := NewIgnoreSet()

	for _, glob := range DefaultIgnorePatterns() {
		set.Add(glob)
	}

	return set
}

// Add adds a glob pattern to the ignore set.
func (s *IgnoreSet) Add(glob string) error {
	result := NewPattern(glob)
	if result.IsErr() {
		return result.UnwrapErr()
	}

	s.patterns = append(s.patterns, result.Unwrap())
	return nil
}

// AddPattern adds a compiled pattern to the ignore set.
func (s *IgnoreSet) AddPattern(pattern *Pattern) {
	s.patterns = append(s.patterns, pattern)
}

// ShouldIgnore checks if a path should be ignored.
// Returns true if the path matches any pattern and is not un-ignored by a negation pattern.
//
// The function checks both full path match and basename match
// to support patterns like ".DS_Store" matching anywhere in the tree.
// Patterns are processed in order, with later patterns overriding earlier ones.
// Negation patterns (starting with !) un-ignore previously ignored files.
func (s *IgnoreSet) ShouldIgnore(path string) bool {
	ignored := false

	// Process patterns in order
	for _, pattern := range s.patterns {
		// Check if pattern matches (full path or basename)
		matches := pattern.Match(path) || pattern.MatchBasename(path)

		if matches {
			if pattern.IsNegation() {
				// Negation pattern un-ignores
				ignored = false
			} else {
				// Normal pattern ignores
				ignored = true
			}
		}
	}

	return ignored
}

// Size returns the number of patterns in the set.
func (s *IgnoreSet) Size() int {
	return len(s.patterns)
}

// Patterns returns the list of patterns in the set.
// Used for merging pattern sets.
func (s *IgnoreSet) Patterns() []*Pattern {
	return s.patterns
}

// DefaultIgnorePatterns returns the default set of patterns to ignore.
// These are common files that should not be managed.
func DefaultIgnorePatterns() []string {
	return []string{
		// Version control
		".git",
		".svn",
		".hg",

		// OS metadata
		".DS_Store",
		"Thumbs.db",
		"desktop.ini",
		".Trash",
		".Spotlight-V100",
		".TemporaryItems",

		// Security-sensitive directories and files
		".gnupg",          // GPG keyring
		".ssh/*.pem",      // SSH private keys (PEM format)
		".ssh/id_*",       // SSH identity files (includes .pub)
		".ssh/*_rsa",      // RSA keys
		".ssh/*_ecdsa",    // ECDSA keys
		".ssh/*_ed25519",  // Ed25519 keys
		".password-store", // pass utility storage
	}
}
