package dot

import "github.com/yaklabco/dot/internal/doctor"

// SensitivePattern represents a pattern for detecting sensitive information.
type SensitivePattern = doctor.SensitivePattern

// SecretDetection represents a detected potential secret.
type SecretDetection = doctor.SecretDetection

// DefaultSensitivePatterns returns the default patterns for secret detection.
func DefaultSensitivePatterns() []SensitivePattern {
	return doctor.DefaultSensitivePatterns()
}

// DetectSecrets scans files for potential secrets using the provided patterns.
func DetectSecrets(files []string, patterns []SensitivePattern) []SecretDetection {
	return doctor.DetectSecrets(files, patterns)
}
