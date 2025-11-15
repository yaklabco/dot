// Package adopt provides interactive adoption functionality.
package adopt

import "time"

// DotfileCandidate represents a file discovered for potential adoption.
type DotfileCandidate struct {
	Path         string    // Absolute path
	RelPath      string    // Path relative to scan root (for display)
	Size         int64     // File size in bytes
	ModTime      time.Time // Last modification time
	IsDir        bool      // True if directory
	Category     string    // Categorization hint: "shell", "git", "vim", etc.
	SuggestedPkg string    // Suggested package name
}

// AdoptGroup represents a logical grouping of files for adoption.
type AdoptGroup struct {
	PackageName string   // Package name to adopt into
	Files       []string // Absolute paths of files to adopt
	Category    string   // Category of the group
}

// DiscoveryOptions configures dotfile discovery.
type DiscoveryOptions struct {
	ScanDirs       []string // Directories to scan
	ExcludeDirs    []string // Directories to exclude
	MaxFileSize    int64    // Maximum file size (0 = no limit)
	IgnorePatterns []string // Additional ignore patterns
}
