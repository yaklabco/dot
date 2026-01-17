package doctor

import (
	"path/filepath"
	"strings"
)

// SeverityLevel represents the severity of a detected secret or sensitive file.
type SeverityLevel int

const (
	// SeverityLow indicates minimal risk.
	SeverityLow SeverityLevel = iota
	// SeverityMedium indicates moderate risk requiring attention.
	SeverityMedium
	// SeverityHigh indicates significant risk requiring prompt action.
	SeverityHigh
	// SeverityCritical indicates critical risk requiring immediate action.
	SeverityCritical
)

// String returns the string representation of the severity level.
func (s SeverityLevel) String() string {
	switch s {
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParseSeverityLevel converts a severity string to a SeverityLevel.
// Returns SeverityLow for unrecognized strings.
func ParseSeverityLevel(severity string) SeverityLevel {
	switch severity {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityLow
	}
}

// SensitivePattern defines patterns that indicate potential secrets.
type SensitivePattern struct {
	Name        string
	Description string
	Patterns    []string
	Severity    SeverityLevel
}

// SecretDetection represents a detected potential secret.
type SecretDetection struct {
	Path    string
	Pattern SensitivePattern
	Target  string
}

// DefaultSensitivePatterns returns patterns for detecting sensitive files.
func DefaultSensitivePatterns() []SensitivePattern {
	return []SensitivePattern{
		{
			Name:        "ssh-keys",
			Description: "SSH private and public keys",
			Patterns: []string{
				"*/.ssh/id_*",
				"*/ssh/id_*", // Without dot prefix
				"*/.ssh/*.pem",
				"*/ssh/*.pem", // Without dot prefix
				"*/.ssh/*_rsa",
				"*/ssh/*_rsa", // Without dot prefix
				"*/.ssh/*_ecdsa",
				"*/ssh/*_ecdsa", // Without dot prefix
				"*/.ssh/*_ed25519",
				"*/ssh/*_ed25519", // Without dot prefix
			},
			Severity: SeverityCritical,
		},
		{
			Name:        "gpg-keys",
			Description: "GPG/PGP keyrings",
			Patterns: []string{
				"*/.gnupg/*",
				"*/.gnupg",
				"*/gnupg/*", // Without dot prefix
				"*/gnupg",   // Without dot prefix
			},
			Severity: SeverityCritical,
		},
		{
			Name:        "credentials",
			Description: "Credential files",
			Patterns: []string{
				"*/.aws/credentials",
				"*/.docker/config.json",
				"*/secrets.*",
			},
			Severity: SeverityHigh,
		},
		{
			Name:        "environment",
			Description: "Environment variable files",
			Patterns: []string{
				"*/.env",
				"*/.env.*",
			},
			Severity: SeverityHigh,
		},
	}
}

// DetectSecrets scans links for potential secrets using provided patterns.
// Returns a list of detections where sensitive files were found.
func DetectSecrets(links []string, patterns []SensitivePattern) []SecretDetection {
	detections := make([]SecretDetection, 0)

	for _, link := range links {
		for _, pattern := range patterns {
			if matchesAnyPattern(link, pattern.Patterns) {
				detections = append(detections, SecretDetection{
					Path:    link,
					Pattern: pattern,
					Target:  link,
				})
				break // Only report first matching pattern per file
			}
		}
	}

	return detections
}

// DetectSecretsWithTargets scans links and their targets for potential secrets.
// This version allows checking both the link path and the target path.
func DetectSecretsWithTargets(links map[string]string, patterns []SensitivePattern) []SecretDetection {
	detections := make([]SecretDetection, 0)

	for link, target := range links {
		for _, pattern := range patterns {
			// Check both link path and target path
			if matchesAnyPattern(link, pattern.Patterns) || matchesAnyPattern(target, pattern.Patterns) {
				detections = append(detections, SecretDetection{
					Path:    link,
					Pattern: pattern,
					Target:  target,
				})
				break // Only report first matching pattern per file
			}
		}
	}

	return detections
}

// matchesAnyPattern checks if a path matches any of the given glob patterns.
func matchesAnyPattern(path string, patterns []string) bool {
	// Normalize path for consistent matching
	normalizedPath := filepath.ToSlash(path)

	for _, pattern := range patterns {
		if matchesPattern(normalizedPath, pattern) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a path matches a glob pattern.
// Supports simple glob patterns with * wildcards.
func matchesPattern(path, pattern string) bool {
	normalizedPattern := filepath.ToSlash(pattern)

	// Handle patterns starting with */
	if strings.HasPrefix(normalizedPattern, "*/") {
		// Remove */ prefix and check if path contains this segment
		segment := normalizedPattern[2:]
		return pathContainsSegment(path, segment)
	}

	// Try direct glob match
	matched, err := filepath.Match(normalizedPattern, path)
	if err == nil && matched {
		return true
	}

	// Try matching basename
	basename := filepath.Base(path)
	matched, err = filepath.Match(normalizedPattern, basename)
	if err == nil && matched {
		return true
	}

	return false
}

// pathContainsSegment checks if path contains the given segment pattern.
func pathContainsSegment(path, segment string) bool {
	// For patterns like ".ssh/id_*", check if path contains this segment
	if strings.Contains(segment, "*") {
		// Split on * and check if all parts are in path in order
		parts := strings.Split(segment, "*")
		if len(parts) == 0 {
			return false
		}

		// Simple case: single * wildcard
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]

			// Find prefix in path
			idx := strings.Index(path, prefix)
			if idx == -1 {
				return false
			}

			// Check if suffix appears after prefix
			remaining := path[idx+len(prefix):]
			if suffix == "" {
				return true
			}
			return strings.Contains(remaining, suffix) || strings.HasSuffix(remaining, suffix)
		}
	}

	// Direct substring match for non-wildcard patterns
	return strings.Contains(path, segment)
}

// FilterBySeverity filters detections by minimum severity level.
func FilterBySeverity(detections []SecretDetection, minSeverity SeverityLevel) []SecretDetection {
	filtered := make([]SecretDetection, 0)

	for _, detection := range detections {
		if detection.Pattern.Severity >= minSeverity {
			filtered = append(filtered, detection)
		}
	}

	return filtered
}
