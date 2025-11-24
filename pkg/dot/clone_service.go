package dot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/bootstrap"
	"github.com/yaklabco/dot/internal/cli/selector"
	"github.com/yaklabco/dot/internal/cli/terminal"
	"github.com/yaklabco/dot/internal/config"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/scanner"
)

// CloneService handles repository cloning and package installation.
type CloneService struct {
	fs         FS
	logger     Logger
	manageSvc  *ManageService
	cloner     adapters.GitCloner
	selector   selector.PackageSelector
	packageDir string
	targetDir  string
	dryRun     bool
}

// newCloneService creates a new clone service.
func newCloneService(
	fs FS,
	logger Logger,
	manageSvc *ManageService,
	cloner adapters.GitCloner,
	sel selector.PackageSelector,
	packageDir string,
	targetDir string,
	dryRun bool,
) *CloneService {
	return &CloneService{
		fs:         fs,
		logger:     logger,
		manageSvc:  manageSvc,
		cloner:     cloner,
		selector:   sel,
		packageDir: packageDir,
		targetDir:  targetDir,
		dryRun:     dryRun,
	}
}

// CloneOptions configures repository cloning behavior.
type CloneOptions struct {
	// Profile specifies which bootstrap profile to use.
	// If empty, uses default profile or interactive selection.
	Profile string

	// Interactive forces interactive package selection.
	// If false, uses profile or installs all packages.
	Interactive bool

	// Force allows cloning into non-empty packageDir.
	Force bool

	// Branch specifies which branch to clone.
	// If empty, clones default branch.
	Branch string
}

// Clone clones a repository and installs packages.
//
// Workflow:
//  1. Validate packageDir is empty (unless Force=true)
//  2. Resolve authentication from environment
//  3. Clone repository to packageDir
//  4. Load bootstrap config if present
//  5. Select packages (profile, interactive, or all)
//  6. Filter packages by current platform
//  7. Install selected packages via ManageService
//  8. Update manifest with repository information
func (s *CloneService) Clone(ctx context.Context, repoURL string, opts CloneOptions) error {
	s.logger.Info(ctx, "clone_operation_started", "url", repoURL, "package_dir", s.packageDir)

	// Validate package directory
	s.logger.Debug(ctx, "validating_package_directory", "path", s.packageDir, "force", opts.Force)
	if err := validatePackageDir(ctx, s.fs, s.packageDir, opts.Force); err != nil {
		s.logger.Error(ctx, "package_directory_validation_failed", "error", err)
		return err
	}
	s.logger.Debug(ctx, "package_directory_validated")

	// Resolve authentication
	s.logger.Debug(ctx, "resolving_authentication", "url", repoURL)
	auth, err := adapters.ResolveAuth(ctx, repoURL)
	if err != nil {
		s.logger.Error(ctx, "authentication_resolution_failed", "error", err)
		return ErrAuthFailed{Cause: err}
	}
	s.logger.Debug(ctx, "authentication_resolved", "method", getAuthMethodName(auth))

	s.logger.Info(ctx, "cloning_repository", "url", repoURL, "destination", s.packageDir)

	// Clone repository
	cloneOpts := adapters.CloneOptions{
		Auth:   auth,
		Branch: opts.Branch,
		Depth:  1, // Shallow clone for faster cloning
	}

	s.logger.Debug(ctx, "initiating_git_clone", "branch", opts.Branch, "depth", 1)
	if err := s.cloner.Clone(ctx, repoURL, s.packageDir, cloneOpts); err != nil {
		s.logger.Error(ctx, "git_clone_failed", "error", err)
		return ErrCloneFailed{URL: repoURL, Cause: err}
	}

	s.logger.Info(ctx, "repository_cloned_successfully", "path", s.packageDir)

	// Load bootstrap configuration if present
	s.logger.Debug(ctx, "checking_for_bootstrap_config")
	bootstrapConfig, hasBootstrap, err := loadBootstrapConfig(ctx, s.fs, s.packageDir)
	if err != nil {
		s.logger.Error(ctx, "bootstrap_config_load_failed", "error", err)
		return err
	}

	if hasBootstrap {
		s.logger.Info(ctx, "bootstrap_config_found", "packages", len(bootstrapConfig.Packages), "profiles", len(bootstrapConfig.Profiles))
	} else {
		s.logger.Debug(ctx, "no_bootstrap_config_found")
	}

	// Select packages to install
	s.logger.Info(ctx, "selecting_packages", "has_bootstrap", hasBootstrap, "profile", opts.Profile, "interactive", opts.Interactive)
	var packagesToInstall []string
	if hasBootstrap {
		packagesToInstall, err = s.selectPackagesWithBootstrap(ctx, bootstrapConfig, opts)
	} else {
		packagesToInstall, err = s.selectPackagesWithoutBootstrap(ctx, opts)
	}
	if err != nil {
		s.logger.Error(ctx, "package_selection_failed", "error", err)
		return err
	}

	if len(packagesToInstall) == 0 {
		s.logger.Info(ctx, "no_packages_selected")
		fmt.Fprintln(os.Stderr, "Warning: No packages selected for installation")
		fmt.Fprintln(os.Stderr, "Repository cloned successfully, but no symlinks were created")
		return nil
	}

	s.logger.Info(ctx, "packages_selected", "count", len(packagesToInstall), "packages", packagesToInstall)

	// Install packages
	if s.dryRun {
		s.logger.Info(ctx, "dry_run_mode", "would_install", packagesToInstall)
		return nil
	}

	s.logger.Info(ctx, "installing_packages", "count", len(packagesToInstall))
	if err := s.manageSvc.Manage(ctx, packagesToInstall...); err != nil {
		s.logger.Error(ctx, "package_installation_failed", "error", err)
		return fmt.Errorf("install packages: %w", err)
	}
	s.logger.Info(ctx, "packages_installed_successfully", "count", len(packagesToInstall))

	// Update manifest with repository information
	s.updateRepoManifest(ctx, repoURL, opts.Branch)

	s.logger.Info(ctx, "clone_complete", "packages_installed", len(packagesToInstall))

	// Offer to persist package directory to config
	if !s.dryRun {
		if err := s.offerToPersistPackageDirectory(ctx, s.packageDir); err != nil {
			s.logger.Warn(ctx, "failed_to_persist_package_directory", "error", err)
		}
	}

	return nil
}

// selectPackagesWithBootstrap selects packages using bootstrap configuration.
func (s *CloneService) selectPackagesWithBootstrap(ctx context.Context, config bootstrap.Config, opts CloneOptions) ([]string, error) {
	// Filter packages by platform
	s.logger.Debug(ctx, "filtering_packages_by_platform", "platform", runtime.GOOS, "total_packages", len(config.Packages))
	filtered := bootstrap.FilterPackagesByPlatform(config.Packages, runtime.GOOS)
	allPackages := extractPackageNames(filtered)
	s.logger.Debug(ctx, "platform_filtered_packages", "count", len(allPackages), "packages", allPackages)

	// Filter out reserved packages with warning
	validPackages := make([]string, 0, len(allPackages))
	skipped := []string{}

	for _, pkg := range allPackages {
		if scanner.IsReservedPackageName(pkg) {
			skipped = append(skipped, pkg)
			continue
		}
		validPackages = append(validPackages, pkg)
	}

	// Show warning if any were skipped
	if len(skipped) > 0 {
		s.logger.Warn(ctx, "skipped_reserved_packages", "packages", skipped)
		fmt.Fprintf(os.Stderr,
			"\nWarning: Skipped %d reserved package(s): %s\n"+
				"Dot cannot manage its own configuration and state files.\n\n",
			len(skipped), strings.Join(skipped, ", "))
	}

	allPackages = validPackages

	// If profile specified, use it
	if opts.Profile != "" {
		s.logger.Info(ctx, "using_specified_profile", "profile", opts.Profile)
		profilePackages, err := selectPackagesFromProfile(config, opts.Profile)
		if err != nil {
			s.logger.Error(ctx, "profile_selection_failed", "profile", opts.Profile, "error", err)
			return nil, err
		}
		result := intersectPackages(profilePackages, allPackages)
		s.logger.Debug(ctx, "profile_packages_selected", "count", len(result))
		return result, nil
	}

	// If interactive flag explicitly set, prompt user
	if opts.Interactive {
		s.logger.Info(ctx, "interactive_mode", "available_packages", len(allPackages))
		return s.selector.Select(ctx, allPackages)
	}

	// Use default profile if configured
	if config.Defaults.Profile != "" {
		s.logger.Info(ctx, "using_default_profile", "profile", config.Defaults.Profile)
		profilePackages, err := selectPackagesFromProfile(config, config.Defaults.Profile)
		if err != nil {
			s.logger.Error(ctx, "default_profile_selection_failed", "profile", config.Defaults.Profile, "error", err)
			return nil, err
		}
		result := intersectPackages(profilePackages, allPackages)
		s.logger.Debug(ctx, "default_profile_packages_selected", "count", len(result))
		return result, nil
	}

	// If terminal is interactive (and no default profile), prompt user
	if terminal.IsInteractive() {
		s.logger.Info(ctx, "terminal_interactive_detected", "prompting_user", true)
		return s.selector.Select(ctx, allPackages)
	}

	// Install all packages (non-interactive mode with no profile)
	s.logger.Info(ctx, "non_interactive_mode", "installing_all_packages", len(allPackages))
	return allPackages, nil
}

// selectPackagesWithoutBootstrap selects packages when no bootstrap config exists.
func (s *CloneService) selectPackagesWithoutBootstrap(ctx context.Context, opts CloneOptions) ([]string, error) {
	// Discover packages in directory
	s.logger.Debug(ctx, "discovering_packages", "directory", s.packageDir)
	packages, err := discoverPackages(ctx, s.fs, s.packageDir)
	if err != nil {
		s.logger.Error(ctx, "package_discovery_failed", "error", err)
		return nil, fmt.Errorf("discover packages: %w", err)
	}

	s.logger.Debug(ctx, "packages_discovered", "count", len(packages), "packages", packages)
	if len(packages) == 0 {
		s.logger.Warn(ctx, "no_packages_found", "packageDir", s.packageDir)
		return []string{}, nil
	}

	// If interactive flag or terminal is interactive, prompt user
	if opts.Interactive || terminal.IsInteractive() {
		s.logger.Info(ctx, "interactive_selection", "available_packages", len(packages))
		return s.selector.Select(ctx, packages)
	}

	// Install all discovered packages
	s.logger.Info(ctx, "auto_selecting_all_packages", "count", len(packages))
	return packages, nil
}

// updateManifestRepository updates the manifest with repository information.
func (s *CloneService) updateManifestRepository(ctx context.Context, info manifest.RepositoryInfo) error {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return targetPathResult.UnwrapErr()
	}

	// Load existing manifest
	manifestStore := manifest.NewFSManifestStore(s.fs)
	manifestResult := manifestStore.Load(ctx, targetPathResult.Unwrap())
	if !manifestResult.IsOk() {
		return manifestResult.UnwrapErr()
	}

	// Update repository info
	m := manifestResult.Unwrap()
	m.SetRepository(info)

	// Save manifest
	return manifestStore.Save(ctx, targetPathResult.Unwrap(), m)
}

// validatePackageDir checks if the package directory is suitable for cloning.
func validatePackageDir(ctx context.Context, fs FS, path string, force bool) error {
	// Check if directory exists
	exists := fs.Exists(ctx, path)
	if !exists {
		return nil // Non-existent directory is fine
	}

	// Check if it's a directory
	isDir, err := fs.IsDir(ctx, path)
	if err != nil {
		return fmt.Errorf("check packageDir: %w", err)
	}
	if !isDir {
		return ErrPackageDirNotEmpty{Path: path, Cause: fmt.Errorf("path exists but is not a directory")}
	}

	// If force flag is set, allow non-empty directory
	if force {
		return nil
	}

	// Check if directory is empty
	entries, err := fs.ReadDir(ctx, path)
	if err != nil {
		return fmt.Errorf("read packageDir: %w", err)
	}

	if len(entries) > 0 {
		return ErrPackageDirNotEmpty{Path: path}
	}

	return nil
}

// loadBootstrapConfig loads the bootstrap configuration if it exists.
func loadBootstrapConfig(ctx context.Context, fs FS, packageDir string) (bootstrap.Config, bool, error) {
	bootstrapPath := filepath.Join(packageDir, ".dotbootstrap.yaml")

	// Check if bootstrap file exists
	if !fs.Exists(ctx, bootstrapPath) {
		return bootstrap.Config{}, false, nil
	}

	// Load and parse bootstrap config
	config, err := bootstrap.Load(ctx, fs, bootstrapPath)
	if err != nil {
		return bootstrap.Config{}, false, ErrInvalidBootstrap{
			Reason: "failed to parse bootstrap configuration",
			Cause:  err,
		}
	}

	return config, true, nil
}

// selectPackagesFromProfile selects packages from a named profile.
func selectPackagesFromProfile(config bootstrap.Config, profileName string) ([]string, error) {
	packages, err := bootstrap.GetProfile(config, profileName)
	if err != nil {
		return nil, ErrProfileNotFound{Profile: profileName}
	}
	return packages, nil
}

// discoverPackages discovers package directories in the package directory.
func discoverPackages(ctx context.Context, fs FS, packageDir string) ([]string, error) {
	entries, err := fs.ReadDir(ctx, packageDir)
	if err != nil {
		return nil, fmt.Errorf("read packageDir: %w", err)
	}

	packages := make([]string, 0)
	for _, entry := range entries {
		// Only include directories, skip files and hidden directories
		if entry.IsDir() && !isHiddenFile(entry.Name()) {
			packages = append(packages, entry.Name())
		}
	}

	return packages, nil
}

// isHiddenFile checks if a filename is hidden (starts with dot).
func isHiddenFile(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// extractPackageNames extracts package names from package specs.
func extractPackageNames(packages []bootstrap.PackageSpec) []string {
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	return names
}

// intersectPackages returns packages present in both lists, preserving order from first list.
func intersectPackages(packages, allowed []string) []string {
	// Build a set of allowed packages for O(1) lookup
	allowedSet := make(map[string]bool, len(allowed))
	for _, pkg := range allowed {
		allowedSet[pkg] = true
	}

	// Filter packages to only those in allowed set
	result := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if allowedSet[pkg] {
			result = append(result, pkg)
		}
	}
	return result
}

// buildRepositoryInfo constructs repository information.
func buildRepositoryInfo(url, branch, commitSHA string) manifest.RepositoryInfo {
	return manifest.RepositoryInfo{
		URL:       url,
		Branch:    branch,
		ClonedAt:  time.Now(),
		CommitSHA: commitSHA,
	}
}

// getCurrentBranch reads the current branch name from a git repository.
// Returns an error if HEAD is detached or cannot be read.
func getCurrentBranch(repoPath string) (string, error) {
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", fmt.Errorf("read HEAD file: %w", err)
	}

	headRef := strings.TrimSpace(string(headData))

	// Parse "ref: refs/heads/branch-name" format
	const refPrefix = "ref: refs/heads/"
	if !strings.HasPrefix(headRef, refPrefix) {
		// Detached HEAD or unexpected format
		return "", fmt.Errorf("detached HEAD or unexpected format")
	}

	// Extract branch name after "ref: refs/heads/"
	branch := headRef[len(refPrefix):]
	if branch == "" {
		return "", fmt.Errorf("empty branch name in HEAD")
	}

	return branch, nil
}

// getCommitSHA attempts to get the current commit SHA from a git repository.
// Returns empty string if unable to determine (best effort).
func getCommitSHA(repoPath string) (string, error) {
	// Read the HEAD file to get current ref
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}

	headRef := strings.TrimSpace(string(headData))

	// If HEAD contains a ref like "ref: refs/heads/main", extract the ref
	const refPrefix = "ref: "
	if strings.HasPrefix(headRef, refPrefix) {
		refPath := strings.TrimSpace(headRef[len(refPrefix):])
		if refPath == "" {
			return "", fmt.Errorf("empty ref path in HEAD")
		}

		// Build full path to ref file
		fullRefPath := filepath.Join(repoPath, ".git", refPath)
		shaData, err := os.ReadFile(fullRefPath)
		if err != nil {
			return "", err
		}

		sha := strings.TrimSpace(string(shaData))
		if len(sha) < 40 {
			return "", fmt.Errorf("invalid SHA length: got %d, expected 40", len(sha))
		}
		return sha[:40], nil
	}

	// HEAD directly contains SHA (detached HEAD)
	if len(headRef) < 40 {
		return "", fmt.Errorf("invalid SHA length in detached HEAD: got %d, expected 40", len(headRef))
	}

	return headRef[:40], nil
}

// getAuthMethodName returns a human-readable name for the authentication method.
func getAuthMethodName(auth adapters.AuthMethod) string {
	if auth == nil {
		return "none"
	}

	switch auth.(type) {
	case adapters.NoAuth:
		return "none"
	case adapters.TokenAuth:
		return "token"
	case adapters.SSHAuth:
		return "ssh"
	default:
		return "unknown"
	}
}

// updateRepoManifest updates the manifest with repository information.
func (s *CloneService) updateRepoManifest(ctx context.Context, repoURL, branchOpt string) {
	s.logger.Debug(ctx, "updating_manifest_with_repository_info")
	branch := branchOpt
	if branch == "" {
		// Read actual branch from repository HEAD
		detectedBranch, err := getCurrentBranch(s.packageDir)
		if err != nil {
			// If we can't detect the branch (detached HEAD, IO error, etc.),
			// fall back to "main" as a sensible default
			s.logger.Warn(ctx, "failed_to_detect_branch", "error", err, "fallback", "main")
			branch = "main"
		} else {
			s.logger.Debug(ctx, "detected_branch", "branch", detectedBranch)
			branch = detectedBranch
		}
	}

	commitSHA, err := getCommitSHA(s.packageDir)
	if err != nil {
		s.logger.Debug(ctx, "failed_to_get_commit_sha", "error", err)
	} else {
		s.logger.Debug(ctx, "detected_commit_sha", "sha", commitSHA)
	}

	repoInfo := buildRepositoryInfo(repoURL, branch, commitSHA)

	if err := s.updateManifestRepository(ctx, repoInfo); err != nil {
		s.logger.Warn(ctx, "failed_to_update_manifest_repository", "error", err)
	} else {
		s.logger.Debug(ctx, "manifest_updated_with_repository_info")
	}
}

// offerToPersistPackageDirectory asks the user if they want to save the package directory to config.
func (s *CloneService) offerToPersistPackageDirectory(ctx context.Context, packageDir string) error {
	configPath := filepath.Join(config.GetConfigPath("dot"), "config.yaml")

	// Check if already set in config
	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err == nil && cfg != nil && cfg.Directories.Package != "" {
		existingAbs, _ := filepath.Abs(cfg.Directories.Package)
		newAbs, _ := filepath.Abs(packageDir)
		if existingAbs == newAbs {
			return nil // Already set correctly
		}
	}

	// Ask user for confirmation
	if !terminal.IsInteractive() {
		return nil // Skip in non-interactive mode
	}

	fmt.Printf("\nSave package directory to config?\n")
	fmt.Printf("  Location: %s\n", packageDir)
	fmt.Printf("  Config:   %s\n\n", configPath)
	fmt.Printf("This will make dot automatically use this directory. [Y/n] ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "" && response != "y" && response != "yes" {
		fmt.Println("Skipped. Use --dir flag or DOT_PACKAGE_DIR environment variable.")
		return nil
	}

	return persistPackageDirectory(packageDir, configPath)
}

// persistPackageDirectory saves the package directory to the config file.
func persistPackageDirectory(packageDir, configPath string) error {
	// Load or create config
	loader := config.NewLoader("dot", configPath)
	cfg, err := loader.LoadWithEnv()
	if err != nil || cfg == nil {
		cfg = config.DefaultExtended()
	}

	// Update package directory
	cfg.Directories.Package = packageDir

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Save config
	writer := config.NewWriter(configPath)
	opts := config.WriteOptions{
		Format:          "yaml",
		IncludeComments: true,
		Indent:          2,
	}
	return writer.Write(cfg, opts)
}
