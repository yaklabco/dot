package doctor_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/doctor"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSensitivePatterns(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	assert.NotEmpty(t, patterns)

	// Check that we have the expected categories
	names := make([]string, len(patterns))
	for i, p := range patterns {
		names[i] = p.Name
	}

	assert.Contains(t, names, "ssh-keys")
	assert.Contains(t, names, "gpg-keys")
	assert.Contains(t, names, "credentials")
	assert.Contains(t, names, "environment")

	// Verify SSH keys pattern has correct severity
	var sshPattern doctor.SensitivePattern
	for _, p := range patterns {
		if p.Name == "ssh-keys" {
			sshPattern = p
			break
		}
	}
	assert.Equal(t, "critical", sshPattern.Severity)
	assert.NotEmpty(t, sshPattern.Patterns)
}

func TestDetectSecrets(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	tests := []struct {
		name             string
		links            []string
		expectedCount    int
		expectedPatterns []string
		expectedNotFound []string
	}{
		{
			name: "detects SSH keys",
			links: []string{
				"/home/user/.ssh/id_rsa",
				"/home/user/.ssh/id_rsa.pub",
				"/home/user/.ssh/config",
			},
			expectedCount:    2, // id_rsa and id_rsa.pub
			expectedPatterns: []string{"ssh-keys"},
			expectedNotFound: []string{"/home/user/.ssh/config"},
		},
		{
			name: "detects GPG keyring",
			links: []string{
				"/home/user/.gnupg",
				"/home/user/.gnupg/pubring.kbx",
				"/home/user/.bashrc",
			},
			expectedCount:    2, // .gnupg directory and file inside
			expectedPatterns: []string{"gpg-keys"},
		},
		{
			name: "detects environment files",
			links: []string{
				"/app/.env",
				"/app/.env.local",
				"/app/config.yaml",
			},
			expectedCount:    2, // .env and .env.local
			expectedPatterns: []string{"environment"},
		},
		{
			name: "detects credentials",
			links: []string{
				"/home/user/.aws/credentials",
				"/home/user/.docker/config.json",
				"/home/user/secrets.yaml",
			},
			expectedCount:    3,
			expectedPatterns: []string{"credentials"},
		},
		{
			name: "detects multiple types",
			links: []string{
				"/home/user/.ssh/id_ed25519",
				"/home/user/.gnupg",
				"/app/.env",
				"/home/user/README.md",
			},
			expectedCount:    3,
			expectedPatterns: []string{"ssh-keys", "gpg-keys", "environment"},
		},
		{
			name: "no detections for safe files",
			links: []string{
				"/home/user/.bashrc",
				"/home/user/.vimrc",
				"/home/user/.config/nvim/init.vim",
			},
			expectedCount: 0,
		},
		{
			name: "detects named SSH keys",
			links: []string{
				"/home/user/.ssh/github_rsa",
				"/home/user/.ssh/gitlab_ecdsa",
				"/home/user/.ssh/bitbucket_ed25519",
			},
			expectedCount:    3,
			expectedPatterns: []string{"ssh-keys"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecrets(tt.links, patterns)

			assert.Len(t, detections, tt.expectedCount, "unexpected number of detections")

			// Check expected patterns are found
			if len(tt.expectedPatterns) > 0 {
				foundPatterns := make(map[string]bool)
				for _, d := range detections {
					foundPatterns[d.Pattern.Name] = true
				}

				for _, expected := range tt.expectedPatterns {
					assert.True(t, foundPatterns[expected], "expected pattern %s not found", expected)
				}
			}

			// Check that specific paths are NOT detected
			if len(tt.expectedNotFound) > 0 {
				detectedPaths := make(map[string]bool)
				for _, d := range detections {
					detectedPaths[d.Path] = true
				}

				for _, path := range tt.expectedNotFound {
					assert.False(t, detectedPaths[path], "path %s should not be detected", path)
				}
			}
		})
	}
}

func TestDetectSecretsWithTargets(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	tests := []struct {
		name          string
		linksTargets  map[string]string
		expectedCount int
	}{
		{
			name: "detects secrets in link paths",
			linksTargets: map[string]string{
				"/home/user/.ssh/id_rsa": "/dotfiles/ssh/id_rsa",
			},
			expectedCount: 1,
		},
		{
			name: "detects secrets in target paths",
			linksTargets: map[string]string{
				"/home/user/.ssh/config": "/dotfiles/ssh/id_rsa",
			},
			expectedCount: 1,
		},
		{
			name: "detects secrets in both link and target",
			linksTargets: map[string]string{
				"/home/user/.ssh/id_rsa": "/dotfiles/ssh/id_rsa",
				"/home/user/.gnupg/key":  "/dotfiles/gpg/key",
				"/home/user/.ssh/config": "/dotfiles/ssh/config",
			},
			expectedCount: 2, // SSH and GPG keys, not config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecretsWithTargets(tt.linksTargets, patterns)
			assert.Len(t, detections, tt.expectedCount)
		})
	}
}

func TestGetSeverityLevel(t *testing.T) {
	tests := []struct {
		severity string
		expected int
	}{
		{"critical", 3},
		{"high", 2},
		{"medium", 1},
		{"low", 0},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			level := doctor.GetSeverityLevel(tt.severity)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestFilterBySeverity(t *testing.T) {
	patterns := []doctor.SensitivePattern{
		{Name: "critical-pattern", Severity: "critical"},
		{Name: "high-pattern", Severity: "high"},
		{Name: "medium-pattern", Severity: "medium"},
	}

	detections := []doctor.SecretDetection{
		{Path: "/path1", Pattern: patterns[0]},
		{Path: "/path2", Pattern: patterns[1]},
		{Path: "/path3", Pattern: patterns[2]},
	}

	tests := []struct {
		name          string
		minSeverity   string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "filter critical only",
			minSeverity:   "critical",
			expectedCount: 1,
			expectedNames: []string{"critical-pattern"},
		},
		{
			name:          "filter high and above",
			minSeverity:   "high",
			expectedCount: 2,
			expectedNames: []string{"critical-pattern", "high-pattern"},
		},
		{
			name:          "filter medium and above",
			minSeverity:   "medium",
			expectedCount: 3,
			expectedNames: []string{"critical-pattern", "high-pattern", "medium-pattern"},
		},
		{
			name:          "filter low (all)",
			minSeverity:   "low",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := doctor.FilterBySeverity(detections, tt.minSeverity)
			assert.Len(t, filtered, tt.expectedCount)

			if len(tt.expectedNames) > 0 {
				foundNames := make([]string, len(filtered))
				for i, d := range filtered {
					foundNames[i] = d.Pattern.Name
				}

				for _, expected := range tt.expectedNames {
					assert.Contains(t, foundNames, expected)
				}
			}
		})
	}
}

func TestMatchesPattern_SSHKeys(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	var sshPattern doctor.SensitivePattern
	for _, p := range patterns {
		if p.Name == "ssh-keys" {
			sshPattern = p
			break
		}
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Should match
		{
			name:     "id_rsa private key",
			path:     "/home/user/.ssh/id_rsa",
			expected: true,
		},
		{
			name:     "id_rsa public key",
			path:     "/home/user/.ssh/id_rsa.pub",
			expected: true,
		},
		{
			name:     "id_ecdsa private key",
			path:     "/home/user/.ssh/id_ecdsa",
			expected: true,
		},
		{
			name:     "id_ed25519 private key",
			path:     "/home/user/.ssh/id_ed25519",
			expected: true,
		},
		{
			name:     "pem file",
			path:     "/home/user/.ssh/mykey.pem",
			expected: true,
		},
		{
			name:     "named rsa key",
			path:     "/home/user/.ssh/github_rsa",
			expected: true,
		},
		{
			name:     "named ecdsa key",
			path:     "/home/user/.ssh/gitlab_ecdsa",
			expected: true,
		},
		{
			name:     "named ed25519 key",
			path:     "/home/user/.ssh/bitbucket_ed25519",
			expected: true,
		},

		// Should NOT match
		{
			name:     "ssh config",
			path:     "/home/user/.ssh/config",
			expected: false,
		},
		{
			name:     "known_hosts",
			path:     "/home/user/.ssh/known_hosts",
			expected: false,
		},
		{
			name:     "authorized_keys",
			path:     "/home/user/.ssh/authorized_keys",
			expected: false,
		},
		{
			name:     "random file in ssh",
			path:     "/home/user/.ssh/notes.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecrets([]string{tt.path}, []doctor.SensitivePattern{sshPattern})

			if tt.expected {
				assert.NotEmpty(t, detections, "expected detection for %s", tt.path)
			} else {
				assert.Empty(t, detections, "unexpected detection for %s", tt.path)
			}
		})
	}
}

func TestMatchesPattern_GPGKeys(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	var gpgPattern doctor.SensitivePattern
	for _, p := range patterns {
		if p.Name == "gpg-keys" {
			gpgPattern = p
			break
		}
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "gnupg directory",
			path:     "/home/user/.gnupg",
			expected: true,
		},
		{
			name:     "gnupg file",
			path:     "/home/user/.gnupg/pubring.kbx",
			expected: true,
		},
		{
			name:     "gnupg private keys",
			path:     "/home/user/.gnupg/private-keys-v1.d/123.key",
			expected: true,
		},
		{
			name:     "other file",
			path:     "/home/user/.bashrc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecrets([]string{tt.path}, []doctor.SensitivePattern{gpgPattern})

			if tt.expected {
				assert.NotEmpty(t, detections, "expected detection for %s", tt.path)
			} else {
				assert.Empty(t, detections, "unexpected detection for %s", tt.path)
			}
		})
	}
}

func TestMatchesPattern_Credentials(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	var credPattern doctor.SensitivePattern
	for _, p := range patterns {
		if p.Name == "credentials" {
			credPattern = p
			break
		}
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "aws credentials",
			path:     "/home/user/.aws/credentials",
			expected: true,
		},
		{
			name:     "docker config",
			path:     "/home/user/.docker/config.json",
			expected: true,
		},
		{
			name:     "secrets yaml",
			path:     "/app/secrets.yaml",
			expected: true,
		},
		{
			name:     "secrets json",
			path:     "/app/secrets.json",
			expected: true,
		},
		{
			name:     "regular config",
			path:     "/home/user/.config/app/config.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecrets([]string{tt.path}, []doctor.SensitivePattern{credPattern})

			if tt.expected {
				assert.NotEmpty(t, detections, "expected detection for %s", tt.path)
			} else {
				assert.Empty(t, detections, "unexpected detection for %s", tt.path)
			}
		})
	}
}

func TestMatchesPattern_Environment(t *testing.T) {
	patterns := doctor.DefaultSensitivePatterns()

	var envPattern doctor.SensitivePattern
	for _, p := range patterns {
		if p.Name == "environment" {
			envPattern = p
			break
		}
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "env file",
			path:     "/app/.env",
			expected: true,
		},
		{
			name:     "env.local file",
			path:     "/app/.env.local",
			expected: true,
		},
		{
			name:     "env.production file",
			path:     "/app/.env.production",
			expected: true,
		},
		{
			name:     "env.example should not match",
			path:     "/app/.env.example",
			expected: true, // Still matches .env.* pattern
		},
		{
			name:     "regular file",
			path:     "/app/config.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detections := doctor.DetectSecrets([]string{tt.path}, []doctor.SensitivePattern{envPattern})

			if tt.expected {
				assert.NotEmpty(t, detections, "expected detection for %s", tt.path)
			} else {
				assert.Empty(t, detections, "unexpected detection for %s", tt.path)
			}
		})
	}
}
