package ignore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/internal/ignore"
)

func TestNewIgnoreSet(t *testing.T) {
	set := ignore.NewIgnoreSet()
	assert.NotNil(t, set)
}

func TestIgnoreSet_Add(t *testing.T) {
	set := ignore.NewIgnoreSet()

	err := set.Add("*.txt")
	assert.NoError(t, err)

	err = set.Add(".git")
	assert.NoError(t, err)
}

func TestIgnoreSet_ShouldIgnore(t *testing.T) {
	set := ignore.NewIgnoreSet()
	set.Add("*.txt")
	set.Add(".git")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "matches txt pattern",
			path:     "file.txt",
			expected: true,
		},
		{
			name:     "matches git pattern",
			path:     ".git",
			expected: true,
		},
		{
			name:     "no match",
			path:     "README.md",
			expected: false,
		},
		{
			name:     "matches git in subdir",
			path:     "project/.git",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := set.ShouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := ignore.DefaultIgnorePatterns()

	assert.NotEmpty(t, patterns)
	assert.Contains(t, patterns, ".git")
	assert.Contains(t, patterns, ".DS_Store")

	// Test security-sensitive patterns are included
	assert.Contains(t, patterns, ".gnupg")
	assert.Contains(t, patterns, ".ssh/*.pem")
	assert.Contains(t, patterns, ".ssh/id_*")
	assert.Contains(t, patterns, ".ssh/*_rsa")
	assert.Contains(t, patterns, ".ssh/*_ecdsa")
	assert.Contains(t, patterns, ".ssh/*_ed25519")
	assert.Contains(t, patterns, ".password-store")
}

func TestNewDefaultIgnoreSet(t *testing.T) {
	set := ignore.NewDefaultIgnoreSet()

	// Should ignore .git
	assert.True(t, set.ShouldIgnore(".git"))
	assert.True(t, set.ShouldIgnore("path/.git"))

	// Should ignore .DS_Store
	assert.True(t, set.ShouldIgnore(".DS_Store"))
	assert.True(t, set.ShouldIgnore("path/.DS_Store"))

	// Should not ignore regular files
	assert.False(t, set.ShouldIgnore("README.md"))
}

func TestIgnoreSet_AddPattern(t *testing.T) {
	set := ignore.NewIgnoreSet()

	pattern := ignore.NewPattern("*.log").Unwrap()
	set.AddPattern(pattern)

	assert.True(t, set.ShouldIgnore("error.log"))
	assert.False(t, set.ShouldIgnore("error.txt"))
}

func TestIgnoreSet_Empty(t *testing.T) {
	set := ignore.NewIgnoreSet()

	// Empty set should not ignore anything
	assert.False(t, set.ShouldIgnore("any/file"))
	assert.False(t, set.ShouldIgnore(".git"))
}

func TestIgnoreSet_Size(t *testing.T) {
	set := ignore.NewIgnoreSet()

	assert.Equal(t, 0, set.Size())

	set.Add("*.txt")
	assert.Equal(t, 1, set.Size())

	set.Add(".git")
	assert.Equal(t, 2, set.Size())
}

func TestIgnoreSet_Negation(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		testPath string
		expected bool
	}{
		{
			name:     "normal ignore",
			patterns: []string{"*.log"},
			testPath: "error.log",
			expected: true,
		},
		{
			name:     "negation un-ignores",
			patterns: []string{"*.log", "!important.log"},
			testPath: "important.log",
			expected: false,
		},
		{
			name:     "negation doesn't affect other matches",
			patterns: []string{"*.log", "!important.log"},
			testPath: "error.log",
			expected: true,
		},
		{
			name:     "order matters - last pattern wins",
			patterns: []string{"!important.log", "*.log"},
			testPath: "important.log",
			expected: true,
		},
		{
			name:     "multiple negations",
			patterns: []string{"*.tmp", "!*.keep", "*.cache"},
			testPath: "data.keep",
			expected: false,
		},
		{
			name:     "negation with no prior ignore",
			patterns: []string{"!*.txt"},
			testPath: "file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := ignore.NewIgnoreSet()
			for _, pattern := range tt.patterns {
				err := set.Add(pattern)
				assert.NoError(t, err)
			}

			result := set.ShouldIgnore(tt.testPath)
			assert.Equal(t, tt.expected, result, "path: %s", tt.testPath)
		})
	}
}

func TestIgnoreSet_ComplexNegation(t *testing.T) {
	// Test a realistic scenario: ignore all cache files except .keep files
	set := ignore.NewIgnoreSet()
	set.Add(".cache/**")
	set.Add("*.tmp")
	set.Add("!*.keep")

	tests := []struct {
		path     string
		expected bool
	}{
		{".cache/data.txt", true},       // Ignored by .cache/**
		{".cache/preserve.keep", false}, // Un-ignored by !*.keep
		{"temp.tmp", true},              // Ignored by *.tmp
		{"preserve.keep", false},        // Un-ignored by !*.keep
		{"normal.txt", false},           // Not matched by any pattern
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := set.ShouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultIgnoreSet_SecuritySensitiveFiles(t *testing.T) {
	set := ignore.NewDefaultIgnoreSet()

	tests := []struct {
		name     string
		path     string
		expected bool
		reason   string
	}{
		// GPG keyring
		{
			name:     "gnupg directory",
			path:     ".gnupg",
			expected: true,
			reason:   "GPG keyring should be ignored",
		},
		{
			name:     "gnupg in subdirectory",
			path:     "home/user/.gnupg",
			expected: true,
			reason:   "GPG keyring in subdirectory should be ignored",
		},

		// SSH keys - private keys
		{
			name:     "ssh id_rsa",
			path:     ".ssh/id_rsa",
			expected: true,
			reason:   "SSH RSA private key should be ignored",
		},
		{
			name:     "ssh id_rsa.pub",
			path:     ".ssh/id_rsa.pub",
			expected: true,
			reason:   "SSH RSA public key should be ignored",
		},
		{
			name:     "ssh id_ecdsa",
			path:     ".ssh/id_ecdsa",
			expected: true,
			reason:   "SSH ECDSA private key should be ignored",
		},
		{
			name:     "ssh id_ecdsa.pub",
			path:     ".ssh/id_ecdsa.pub",
			expected: true,
			reason:   "SSH ECDSA public key should be ignored",
		},
		{
			name:     "ssh id_ed25519",
			path:     ".ssh/id_ed25519",
			expected: true,
			reason:   "SSH Ed25519 private key should be ignored",
		},
		{
			name:     "ssh id_ed25519.pub",
			path:     ".ssh/id_ed25519.pub",
			expected: true,
			reason:   "SSH Ed25519 public key should be ignored",
		},
		{
			name:     "ssh pem file",
			path:     ".ssh/mykey.pem",
			expected: true,
			reason:   "SSH PEM key should be ignored",
		},
		{
			name:     "ssh named rsa key",
			path:     ".ssh/github_rsa",
			expected: true,
			reason:   "Named RSA key should be ignored",
		},
		{
			name:     "ssh named ecdsa key",
			path:     ".ssh/gitlab_ecdsa",
			expected: true,
			reason:   "Named ECDSA key should be ignored",
		},
		{
			name:     "ssh named ed25519 key",
			path:     ".ssh/bitbucket_ed25519",
			expected: true,
			reason:   "Named Ed25519 key should be ignored",
		},

		// SSH config files should NOT be ignored
		{
			name:     "ssh config",
			path:     ".ssh/config",
			expected: false,
			reason:   "SSH config should NOT be ignored",
		},
		{
			name:     "ssh known_hosts",
			path:     ".ssh/known_hosts",
			expected: false,
			reason:   "SSH known_hosts should NOT be ignored",
		},
		{
			name:     "ssh authorized_keys",
			path:     ".ssh/authorized_keys",
			expected: false,
			reason:   "SSH authorized_keys should NOT be ignored",
		},

		// pass password store
		{
			name:     "password-store directory",
			path:     ".password-store",
			expected: true,
			reason:   "pass password store should be ignored",
		},
		{
			name:     "password-store in subdirectory",
			path:     "home/user/.password-store",
			expected: true,
			reason:   "pass password store in subdirectory should be ignored",
		},

		// Version control (existing tests)
		{
			name:     "git directory",
			path:     ".git",
			expected: true,
			reason:   "git directory should be ignored",
		},

		// Regular files should not be ignored
		{
			name:     "regular file",
			path:     "README.md",
			expected: false,
			reason:   "regular files should not be ignored",
		},
		{
			name:     "dotfile",
			path:     ".bashrc",
			expected: false,
			reason:   "regular dotfiles should not be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := set.ShouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}
