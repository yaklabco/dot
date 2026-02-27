// Package adapters provides interfaces and implementations for external dependencies.
package adapters

import (
	"context"
	"io"
)

// GitCloner defines the interface for cloning git repositories.
type GitCloner interface {
	// Clone clones a repository from the specified URL to the target path.
	//
	// Returns an error if:
	//   - URL is invalid
	//   - Target path already exists
	//   - Authentication fails
	//   - Network errors occur
	//   - Repository is not accessible
	Clone(ctx context.Context, url string, path string, opts CloneOptions) error
}

// CloneOptions configures repository cloning behavior.
type CloneOptions struct {
	// Auth specifies the authentication method.
	// If nil, no authentication is used (public repos only).
	Auth AuthMethod

	// Branch specifies which branch to clone.
	// If empty, the default branch is cloned.
	Branch string

	// Depth specifies how many commits to fetch.
	// If 0, full history is cloned.
	// If 1, only the latest commit is fetched (shallow clone).
	Depth int

	// Progress is an optional writer for clone progress output.
	// If nil, no progress is reported.
	Progress io.Writer
}

// AuthMethod represents a git authentication method.
//
// This is a sealed interface implemented only by:
//   - NoAuth
//   - TokenAuth
//   - SSHAuth
type AuthMethod interface {
	// isAuthMethod is unexported to seal the interface.
	isAuthMethod()
}

// NoAuth represents no authentication (public repositories).
type NoAuth struct{}

func (NoAuth) isAuthMethod() {}

// TokenAuth represents token-based authentication (HTTPS).
//
// The token is transmitted using HTTP Basic Authentication with "git" as the
// username and the token as the password. This format is compatible with:
//   - GitHub personal access tokens
//   - GitHub fine-grained tokens
//   - GitLab personal access tokens
//   - Gitea tokens
//   - Azure DevOps personal access tokens
type TokenAuth struct {
	// Token is the authentication token.
	Token string
}

func (TokenAuth) isAuthMethod() {}

// SSHAuth represents SSH key-based authentication.
type SSHAuth struct {
	// PrivateKeyPath is the filesystem path to the SSH private key.
	PrivateKeyPath string

	// Passphrase is an optional passphrase for encrypted keys.
	Passphrase string
}

func (SSHAuth) isAuthMethod() {}
