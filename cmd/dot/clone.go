package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaklabco/dot/pkg/dot"
)

// newCloneCommand creates the clone command.
func newCloneCommand() *cobra.Command {
	var (
		cloneProfile     string
		cloneInteractive bool
		cloneForce       bool
		cloneBranch      string
	)

	cmd := &cobra.Command{
		Use:   "clone <repository-url>",
		Short: "Clone dotfiles repository and install packages",
		Long: `Clone a dotfiles repository and install packages.

Like git clone, the repository is cloned into a subdirectory named after the
repository. Use --dir to specify a different target directory.

WORKFLOW:
  1. Clone repository to target directory
  2. Load optional .dotbootstrap.yaml configuration
  3. Select packages (via profile, interactive, or all)
  4. Filter by current platform
  5. Install selected packages
  6. Track repository in manifest

REPOSITORY CONFIGURATION:
  If repository contains .config/dot/config.yaml, it will be
  automatically used for all subsequent dot commands.

AUTHENTICATION:
  Automatic resolution order:
  1. GITHUB_TOKEN environment variable
  2. GIT_TOKEN environment variable
  3. SSH keys (~/.ssh/)
  4. GitHub CLI (gh) authenticated session
  5. No authentication (public repos)

BOOTSTRAP CONFIGURATION:
  Optional .dotbootstrap.yaml defines installation profiles,
  platform requirements, and package metadata.

Examples:
  # Clone and install all packages (creates ./dotfiles directory)
  dot clone https://github.com/user/dotfiles

  # Clone creates ./my-dotfiles directory
  dot clone https://github.com/user/my-dotfiles

  # Clone specific branch
  dot clone https://github.com/user/dotfiles --branch develop

  # Use named profile from bootstrap config
  dot clone https://github.com/user/dotfiles --profile minimal

  # Force interactive selection
  dot clone https://github.com/user/dotfiles --interactive

  # Clone to specific directory (overrides default)
  dot clone --dir ~/packages https://github.com/user/dotfiles

  # Overwrite existing package directory
  dot clone --force https://github.com/user/dotfiles

  # Clone via SSH
  dot clone git@github.com:user/dotfiles.git`,
		Args: argsWithUsage(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(cmd, args, cloneProfile, cloneInteractive, cloneForce, cloneBranch)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVar(&cloneProfile, "profile", "", "installation profile from bootstrap config")
	cmd.Flags().BoolVar(&cloneInteractive, "interactive", false, "interactively select packages")
	cmd.Flags().BoolVar(&cloneForce, "force", false, "overwrite package directory if exists")
	cmd.Flags().StringVar(&cloneBranch, "branch", "", "branch to clone (defaults to repository default)")

	// Add bootstrap subcommand
	cmd.AddCommand(newCloneBootstrapCommand())

	return cmd
}

// runClone handles the clone command execution.
func runClone(cmd *cobra.Command, args []string, profile string, interactive bool, force bool, branch string) error {
	repoURL := args[0]

	// Check if --dir flag was explicitly provided
	dirFlag := cmd.Flags().Lookup("dir")
	dirExplicitlySet := dirFlag != nil && dirFlag.Changed

	// If --dir was NOT explicitly set, derive directory from repo URL
	// This makes dot clone behave like git clone
	if !dirExplicitlySet {
		repoName := extractRepoName(repoURL)
		cwd, err := os.Getwd()
		if err != nil {
			return formatError(fmt.Errorf("get current directory: %w", err))
		}
		// Override the CLI flags packageDir with repo-based directory in current working directory
		cliFlags.packageDir = filepath.Join(cwd, repoName)
	}

	// Build config (will use the modified packageDir if set above)
	cfg, err := buildConfigWithCmd(cmd)
	if err != nil {
		return formatError(err)
	}

	// Create client
	client, err := dot.NewClient(cfg)
	if err != nil {
		return formatError(err)
	}

	// Get context
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Build clone options
	opts := dot.CloneOptions{
		Profile:     profile,
		Interactive: interactive,
		Force:       force,
		Branch:      branch,
	}

	// Execute clone
	if err := client.Clone(ctx, repoURL, opts); err != nil {
		return formatCloneError(err)
	}

	return nil
}

// formatCloneError formats clone-specific errors with helpful messages.
func formatCloneError(err error) error {
	var packageDirNotEmpty dot.ErrPackageDirNotEmpty
	if errors.As(err, &packageDirNotEmpty) {
		return fmt.Errorf("%w\n\nUse --force to overwrite the existing directory", packageDirNotEmpty)
	}

	var bootstrapNotFound dot.ErrBootstrapNotFound
	if errors.As(err, &bootstrapNotFound) {
		return fmt.Errorf("%w\n\nThe repository may not have been properly cloned", bootstrapNotFound)
	}

	var invalidBootstrap dot.ErrInvalidBootstrap
	if errors.As(err, &invalidBootstrap) {
		return fmt.Errorf("%w\n\nCheck the .dotbootstrap.yaml syntax and validation rules", invalidBootstrap)
	}

	var authFailed dot.ErrAuthFailed
	if errors.As(err, &authFailed) {
		return fmt.Errorf("%w\n\nTry:\n  - Setting GITHUB_TOKEN environment variable\n  - Setting GIT_TOKEN environment variable\n  - Configuring SSH keys in ~/.ssh/", authFailed)
	}

	var cloneFailed dot.ErrCloneFailed
	if errors.As(err, &cloneFailed) {
		return fmt.Errorf("%w\n\nEnsure:\n  - URL is correct\n  - Repository is accessible\n  - Network connection is available\n  - Authentication is configured (for private repos)", cloneFailed)
	}

	var profileNotFound dot.ErrProfileNotFound
	if errors.As(err, &profileNotFound) {
		return fmt.Errorf("%w\n\nCheck available profiles in .dotbootstrap.yaml", profileNotFound)
	}

	return err
}

// extractRepoName extracts the repository name from a URL.
// It handles various URL formats:
//   - HTTPS: https://github.com/user/my-dotfiles.git -> my-dotfiles
//   - SSH: git@github.com:user/my-dotfiles.git -> my-dotfiles
//   - Simple paths: my-dotfiles -> my-dotfiles
//
// Returns "dotfiles" as a fallback if extraction fails.
func extractRepoName(repoURL string) string {
	if repoURL == "" {
		return "dotfiles"
	}

	// Remove query parameters if present (e.g., ?ref=main)
	if idx := strings.Index(repoURL, "?"); idx != -1 {
		repoURL = repoURL[:idx]
	}

	// Handle SSH URLs: git@github.com:user/repo -> user/repo
	if strings.Contains(repoURL, ":") && strings.Contains(repoURL, "@") {
		parts := strings.Split(repoURL, ":")
		if len(parts) >= 2 {
			repoURL = parts[len(parts)-1]
		}
	}

	// Extract last path component
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		if name != "" {
			// Remove .git suffix if present
			name = strings.TrimSuffix(name, ".git")
			return name
		}
	}

	return "dotfiles" // fallback
}
