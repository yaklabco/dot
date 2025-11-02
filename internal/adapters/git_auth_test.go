package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAuth_WithGitHubToken(t *testing.T) {
	ctx := context.Background()

	// Set GITHUB_TOKEN environment variable
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	os.Setenv("GITHUB_TOKEN", "ghp_test123")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	tokenAuth, ok := auth.(TokenAuth)
	assert.True(t, ok)
	assert.Equal(t, "ghp_test123", tokenAuth.Token)
}

func TestResolveAuth_WithGitToken(t *testing.T) {
	ctx := context.Background()

	// Clear GITHUB_TOKEN and set GIT_TOKEN
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		} else {
			os.Unsetenv("GIT_TOKEN")
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Setenv("GIT_TOKEN", "token123")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	tokenAuth, ok := auth.(TokenAuth)
	assert.True(t, ok)
	assert.Equal(t, "token123", tokenAuth.Token)
}

func TestResolveAuth_WithSSHKey(t *testing.T) {
	ctx := context.Background()

	// Clear token environment variables
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")

	// Create a temporary SSH key file
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "id_rsa")
	err := os.WriteFile(keyPath, []byte("fake-ssh-key"), 0600)
	require.NoError(t, err)

	// Create mock home directory structure
	sshDir := filepath.Join(tempDir, ".ssh")
	err = os.MkdirAll(sshDir, 0700)
	require.NoError(t, err)

	mockKeyPath := filepath.Join(sshDir, "id_rsa")
	err = os.WriteFile(mockKeyPath, []byte("fake-ssh-key"), 0600)
	require.NoError(t, err)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()
	os.Setenv("HOME", tempDir)

	auth, err := ResolveAuth(ctx, "git@github.com:user/repo.git")
	require.NoError(t, err)

	sshAuth, ok := auth.(SSHAuth)
	assert.True(t, ok)
	assert.Equal(t, mockKeyPath, sshAuth.PrivateKeyPath)
}

func TestResolveAuth_NoAuth(t *testing.T) {
	ctx := context.Background()

	// Clear all auth environment variables
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")
	os.Setenv("HOME", "/nonexistent")

	// Use non-GitHub URL to test NoAuth fallback without gh CLI interference
	auth, err := ResolveAuth(ctx, "https://gitlab.com/user/repo")
	require.NoError(t, err)

	_, ok := auth.(NoAuth)
	assert.True(t, ok)
}

func TestResolveAuth_SSHURLWithoutKeys(t *testing.T) {
	ctx := context.Background()

	// Clear all auth
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")
	os.Setenv("HOME", "/nonexistent")

	auth, err := ResolveAuth(ctx, "git@github.com:user/repo.git")
	require.NoError(t, err)

	// Should fall back to NoAuth if no SSH keys found
	_, ok := auth.(NoAuth)
	assert.True(t, ok)
}

func TestFindSSHKey(t *testing.T) {
	t.Run("finds id_rsa", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		keyPath := filepath.Join(sshDir, "id_rsa")
		err = os.WriteFile(keyPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, keyPath, found)
	})

	t.Run("finds id_ed25519", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		keyPath := filepath.Join(sshDir, "id_ed25519")
		err = os.WriteFile(keyPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, keyPath, found)
	})

	t.Run("prefers id_ed25519 over id_rsa", func(t *testing.T) {
		tempDir := t.TempDir()
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		rsaPath := filepath.Join(sshDir, "id_rsa")
		err = os.WriteFile(rsaPath, []byte("fake-key"), 0600)
		require.NoError(t, err)

		ed25519Path := filepath.Join(sshDir, "id_ed25519")
		err = os.WriteFile(ed25519Path, []byte("fake-key"), 0600)
		require.NoError(t, err)

		found := findSSHKey(tempDir)
		assert.Equal(t, ed25519Path, found)
	})

	t.Run("returns empty when no keys found", func(t *testing.T) {
		tempDir := t.TempDir()
		found := findSSHKey(tempDir)
		assert.Empty(t, found)
	})
}

func TestIsSSHURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"git@github.com:user/repo.git", true},
		{"git@gitlab.com:user/repo.git", true},
		{"ssh://git@github.com/user/repo.git", true},
		{"https://github.com/user/repo", false},
		{"http://github.com/user/repo", false},
		{"file:///path/to/repo", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isSSHURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveAuth_WithGitHubCLI(t *testing.T) {
	ctx := context.Background()

	// Clear environment tokens
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		} else {
			os.Unsetenv("GIT_TOKEN")
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")

	// Test with GitHub URL
	// Note: This test may pass or fail depending on local gh CLI authentication
	// In CI, this should gracefully return NoAuth if gh is not authenticated
	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	// Should either be TokenAuth (if gh is authenticated) or NoAuth (if not)
	_, isToken := auth.(TokenAuth)
	_, isNoAuth := auth.(NoAuth)
	assert.True(t, isToken || isNoAuth, "Expected TokenAuth or NoAuth")
}

func TestResolveAuth_GitHubCLI_WithEnvVarOverride(t *testing.T) {
	ctx := context.Background()

	// Set GITHUB_TOKEN environment variable
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	os.Setenv("GITHUB_TOKEN", "env_token_123")

	auth, err := ResolveAuth(ctx, "https://github.com/user/repo")
	require.NoError(t, err)

	tokenAuth, ok := auth.(TokenAuth)
	assert.True(t, ok)
	// Environment variable should take precedence over gh CLI
	assert.Equal(t, "env_token_123", tokenAuth.Token)
}

func TestResolveAuth_GitHubCLI_NotGitHub(t *testing.T) {
	ctx := context.Background()

	// Clear environment tokens
	originalGitHubToken := os.Getenv("GITHUB_TOKEN")
	originalGitToken := os.Getenv("GIT_TOKEN")
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalGitHubToken != "" {
			os.Setenv("GITHUB_TOKEN", originalGitHubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if originalGitToken != "" {
			os.Setenv("GIT_TOKEN", originalGitToken)
		} else {
			os.Unsetenv("GIT_TOKEN")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GIT_TOKEN")
	os.Setenv("HOME", "/nonexistent")

	// Test with non-GitHub URL (GitLab)
	auth, err := ResolveAuth(ctx, "https://gitlab.com/user/repo")
	require.NoError(t, err)

	// Should be NoAuth since it's not GitHub and no other auth available
	_, ok := auth.(NoAuth)
	assert.True(t, ok)
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "HTTPS GitHub URL",
			url:      "https://github.com/user/repo",
			expected: true,
		},
		{
			name:     "HTTPS GitHub URL with .git",
			url:      "https://github.com/user/repo.git",
			expected: true,
		},
		{
			name:     "SSH GitHub URL",
			url:      "git@github.com:user/repo.git",
			expected: true,
		},
		{
			name:     "SSH GitHub URL without .git",
			url:      "git@github.com:user/repo",
			expected: true,
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/user/repo",
			expected: false,
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:user/repo.git",
			expected: false,
		},
		{
			name:     "Bitbucket URL",
			url:      "https://bitbucket.org/user/repo",
			expected: false,
		},
		{
			name:     "Generic git URL",
			url:      "https://git.example.com/repo",
			expected: false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "GitHub Enterprise (contains github.com)",
			url:      "https://github.company.com/user/repo",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitHubURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
