package adapters

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// GoGitCloner implements GitCloner using go-git library.
type GoGitCloner struct{}

// NewGoGitCloner creates a new go-git based cloner.
func NewGoGitCloner() *GoGitCloner {
	return &GoGitCloner{}
}

// Clone clones a git repository using go-git.
func (g *GoGitCloner) Clone(ctx context.Context, url string, path string, opts CloneOptions) error {
	// Check if target path already exists and is not empty
	if err := validateTargetPath(path); err != nil {
		return err
	}

	// Convert auth method to go-git transport auth
	auth, err := convertAuthMethod(opts.Auth)
	if err != nil {
		return fmt.Errorf("configure authentication: %w", err)
	}

	// Build clone options
	cloneOpts := &git.CloneOptions{
		URL:      url,
		Progress: opts.Progress,
		Auth:     auth,
	}

	// Set branch reference if specified
	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
	}

	// Set depth for shallow clone
	if opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	}

	// Perform clone with context
	_, err = git.PlainCloneContext(ctx, path, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("clone repository: %w", err)
	}

	return nil
}

// validateTargetPath checks if the target path is suitable for cloning.
func validateTargetPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist - this is fine
			return nil
		}
		return fmt.Errorf("check target path: %w", err)
	}

	// Path exists - check if it's empty
	if !info.IsDir() {
		return fmt.Errorf("target path exists and is not a directory: %s", path)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read target directory: %w", err)
	}

	if len(entries) > 0 {
		return fmt.Errorf("target directory already exists and is not empty: %s", path)
	}

	return nil
}

// convertAuthMethod converts our AuthMethod to go-git transport auth.
func convertAuthMethod(auth AuthMethod) (transport.AuthMethod, error) {
	if auth == nil {
		return nil, nil
	}

	switch a := auth.(type) {
	case NoAuth:
		return nil, nil

	case TokenAuth:
		// Most git providers (GitHub, GitLab, Gitea, Azure DevOps) expect the token
		// in the password field with a placeholder username
		return &http.BasicAuth{
			Username: "git",
			Password: a.Token, // Token goes in password field
		}, nil

	case SSHAuth:
		// Load SSH private key
		publicKeys, err := ssh.NewPublicKeysFromFile("git", a.PrivateKeyPath, a.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("load SSH key: %w", err)
		}
		return publicKeys, nil

	default:
		return nil, fmt.Errorf("unsupported authentication method: %T", auth)
	}
}
