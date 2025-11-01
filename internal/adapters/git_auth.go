package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/pkg/auth"
)

// ResolveAuth determines the appropriate authentication method for a repository URL.
//
// Resolution priority:
//  1. GITHUB_TOKEN environment variable → TokenAuth
//  2. GIT_TOKEN environment variable → TokenAuth
//  3. SSH keys in ~/.ssh/ → SSHAuth (for SSH URLs)
//  4. GitHub CLI (gh) authenticated token → TokenAuth (for HTTPS GitHub URLs)
//  5. NoAuth (public repositories)
//
// The function inspects the URL to determine authentication needs.
// For SSH URLs (git@... or ssh://...), SSH key auth is preferred.
// For GitHub HTTPS URLs, checks gh CLI if environment tokens not set.
// This ensures SSH URLs use SSH keys as users expect.
func ResolveAuth(ctx context.Context, repoURL string) (AuthMethod, error) {
	// Priority 1: Check for token in environment variables
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return TokenAuth{Token: token}, nil
	}

	if token := os.Getenv("GIT_TOKEN"); token != "" {
		return TokenAuth{Token: token}, nil
	}

	// Priority 2: For SSH URLs, try to find SSH keys
	// SSH URLs explicitly request SSH auth, so honor that before trying tokens
	if isSSHURL(repoURL) {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			if keyPath := findSSHKey(homeDir); keyPath != "" {
				return SSHAuth{PrivateKeyPath: keyPath}, nil
			}
		}
	}

	// Priority 3: GitHub CLI for HTTPS GitHub URLs
	if isGitHubURL(repoURL) && !isSSHURL(repoURL) {
		if token := getGitHubCLIToken(); token != "" {
			return TokenAuth{Token: token}, nil
		}
	}

	// Priority 4: Fall back to no authentication (public repos)
	return NoAuth{}, nil
}

// isSSHURL checks if a URL uses SSH protocol.
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://")
}

// findSSHKey searches for common SSH private keys in the user's home directory.
//
// Checks for keys in this order:
//  1. ~/.ssh/id_ed25519 (modern, preferred)
//  2. ~/.ssh/id_rsa (older, common)
//
// Returns the path to the first key found, or empty string if none exist.
func findSSHKey(homeDir string) string {
	sshDir := filepath.Join(homeDir, ".ssh")

	// Check for Ed25519 key (preferred)
	ed25519Key := filepath.Join(sshDir, "id_ed25519")
	if _, err := os.Stat(ed25519Key); err == nil {
		return ed25519Key
	}

	// Check for RSA key (fallback)
	rsaKey := filepath.Join(sshDir, "id_rsa")
	if _, err := os.Stat(rsaKey); err == nil {
		return rsaKey
	}

	return ""
}

// getGitHubCLIToken attempts to retrieve the GitHub token from gh CLI.
// Uses the official go-gh library to access gh's stored credentials.
// Returns empty string if gh CLI is not authenticated or not installed.
func getGitHubCLIToken() string {
	token, _ := auth.TokenForHost("github.com")
	// TokenForHost returns ("", "default") if no token is found
	return token
}

// isGitHubURL checks if a URL is a GitHub URL.
// Supports both HTTPS and SSH GitHub URLs.
func isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com")
}
