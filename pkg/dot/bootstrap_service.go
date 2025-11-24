package dot

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/yaklabco/dot/internal/bootstrap"
	"github.com/yaklabco/dot/internal/manifest"
)

// BootstrapService handles bootstrap configuration generation.
type BootstrapService struct {
	fs         FS
	logger     Logger
	packageDir string
	targetDir  string
}

// newBootstrapService creates a new bootstrap service.
func newBootstrapService(
	fs FS,
	logger Logger,
	packageDir string,
	targetDir string,
) *BootstrapService {
	return &BootstrapService{
		fs:         fs,
		logger:     logger,
		packageDir: packageDir,
		targetDir:  targetDir,
	}
}

// GenerateBootstrapOptions configures bootstrap generation.
type GenerateBootstrapOptions struct {
	// FromManifest only includes packages from manifest
	FromManifest bool

	// ConflictPolicy sets default conflict resolution policy
	ConflictPolicy string

	// IncludeComments adds helpful comments to output
	IncludeComments bool

	// Force allows overwriting existing bootstrap files
	Force bool
}

// BootstrapResult contains the generated bootstrap configuration.
type BootstrapResult struct {
	// Config is the generated bootstrap configuration
	Config bootstrap.Config

	// YAML is the marshaled configuration
	YAML []byte

	// PackageCount is the number of packages included
	PackageCount int

	// InstalledCount is the number of packages currently installed
	InstalledCount int
}

// GenerateBootstrap creates a bootstrap configuration from the current installation.
//
// The method:
//  1. Scans package directory for available packages
//  2. Reads manifest to identify installed packages
//  3. Generates bootstrap configuration with appropriate defaults
//  4. Marshals configuration to YAML
//
// Returns an error if:
//   - Package directory cannot be scanned
//   - No packages are found
//   - Configuration generation fails
//   - YAML marshaling fails
func (s *BootstrapService) GenerateBootstrap(ctx context.Context, opts GenerateBootstrapOptions) (BootstrapResult, error) {
	// Discover packages
	packageNames, err := s.discoverPackages(ctx)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("discover packages: %w", err)
	}

	if len(packageNames) == 0 {
		return BootstrapResult{}, fmt.Errorf("no packages found in %s", s.packageDir)
	}

	// Read manifest to find installed packages
	installed, err := s.getInstalledPackages(ctx)
	if err != nil {
		s.logger.Warn(ctx, "failed to read manifest", "error", err)
		// Continue with empty installed list
		installed = []string{}
	}

	// Generate configuration
	gen := bootstrap.NewGenerator()
	genOpts := bootstrap.GenerateOptions{
		FromManifest:    opts.FromManifest,
		ConflictPolicy:  opts.ConflictPolicy,
		IncludeComments: opts.IncludeComments,
	}

	cfg, err := gen.Generate(packageNames, installed, genOpts)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("generate config: %w", err)
	}

	// Marshal to YAML
	var yamlData []byte
	if opts.IncludeComments {
		yamlData, err = gen.MarshalYAMLWithComments(cfg, installed)
	} else {
		yamlData, err = gen.MarshalYAML(cfg)
	}
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("marshal YAML: %w", err)
	}

	return BootstrapResult{
		Config:         cfg,
		YAML:           yamlData,
		PackageCount:   len(cfg.Packages),
		InstalledCount: len(installed),
	}, nil
}

// WriteBootstrap writes the bootstrap configuration to a file.
//
// Returns an error if:
//   - File already exists (unless Force option was set during generation)
//   - File cannot be written
func (s *BootstrapService) WriteBootstrap(ctx context.Context, data []byte, outputPath string) error {
	// Check if file exists
	if s.fs.Exists(ctx, outputPath) {
		// Note: Force checking happens in the CLI layer where the option is available
		// For now, we return an error if file exists
		return ErrBootstrapExists{Path: outputPath}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if err := s.fs.MkdirAll(ctx, dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := s.fs.WriteFile(ctx, outputPath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// getInstalledPackages retrieves the list of installed packages from manifest.
func (s *BootstrapService) getInstalledPackages(ctx context.Context) ([]string, error) {
	// Read manifest file
	manifestPath := filepath.Join(s.targetDir, ".dot-manifest.json")
	if !s.fs.Exists(ctx, manifestPath) {
		return []string{}, nil // No manifest is okay
	}

	data, err := s.fs.ReadFile(ctx, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest file: %w", err)
	}

	var m manifest.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest JSON: %w", err)
	}

	installedPackages := make([]string, 0, len(m.Packages))
	for _, pkg := range m.Packages {
		installedPackages = append(installedPackages, pkg.Name)
	}

	return installedPackages, nil
}

// discoverPackages discovers package directories in the package directory.
func (s *BootstrapService) discoverPackages(ctx context.Context) ([]string, error) {
	entries, err := s.fs.ReadDir(ctx, s.packageDir)
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
