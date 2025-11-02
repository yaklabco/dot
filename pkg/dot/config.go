package dot

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// Config holds configuration for the dot Client.
type Config struct {
	// PackageDir is the source directory containing packages.
	// Must be an absolute path.
	PackageDir string

	// TargetDir is the destination directory for symlinks.
	// Must be an absolute path.
	TargetDir string

	// LinkMode specifies whether to create relative or absolute symlinks.
	LinkMode LinkMode

	// Folding enables directory-level linking when all contents
	// belong to a single package.
	Folding bool

	// DryRun enables preview mode without applying changes.
	DryRun bool

	// Verbosity controls logging detail (0=quiet, 1=info, 2=debug, 3=trace).
	Verbosity int

	// BackupDir specifies where to store backup files.
	// If empty, backups go to <TargetDir>/.dot-backup/
	BackupDir string

	// Backup enables automatic backup of conflicting files.
	// When true, conflicting files are backed up before being replaced.
	Backup bool

	// Overwrite enables automatic overwriting of conflicting files.
	// When true, conflicting files are deleted before creating symlinks.
	// Takes precedence over Backup if both are true.
	Overwrite bool

	// ManifestDir specifies where to store the manifest file.
	// If empty, manifest is stored in TargetDir for backward compatibility.
	ManifestDir string

	// Concurrency limits parallel operation execution.
	// If zero, defaults to runtime.NumCPU().
	Concurrency int

	// PackageNameMapping enables package name to target directory mapping.
	// When enabled, package "dot-gnupg" targets ~/.gnupg/ instead of ~/.
	// Default: true (project is pre-1.0, breaking change acceptable)
	PackageNameMapping bool

	// Infrastructure dependencies (required)
	FS      FS
	Logger  Logger
	Tracer  Tracer
	Metrics Metrics
}

// LinkMode specifies symlink creation strategy.
type LinkMode int

const (
	// LinkRelative creates relative symlinks (default).
	LinkRelative LinkMode = iota
	// LinkAbsolute creates absolute symlinks.
	LinkAbsolute
)

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.PackageDir == "" {
		return fmt.Errorf("packageDir is required")
	}
	if !filepath.IsAbs(c.PackageDir) {
		return fmt.Errorf("packageDir must be absolute path: %s", c.PackageDir)
	}

	if c.TargetDir == "" {
		return fmt.Errorf("targetDir is required")
	}
	if !filepath.IsAbs(c.TargetDir) {
		return fmt.Errorf("targetDir must be absolute path: %s", c.TargetDir)
	}

	if c.FS == nil {
		return fmt.Errorf("FS is required")
	}

	if c.Logger == nil {
		return fmt.Errorf("Logger is required")
	}

	if c.Verbosity < 0 {
		return fmt.Errorf("verbosity cannot be negative")
	}

	if c.Concurrency < 0 {
		return fmt.Errorf("concurrency cannot be negative")
	}

	return nil
}

// WithDefaults returns a copy of the config with defaults applied.
func (c Config) WithDefaults() Config {
	cfg := c

	if cfg.Tracer == nil {
		cfg.Tracer = NewNoopTracer()
	}

	if cfg.Metrics == nil {
		cfg.Metrics = NewNoopMetrics()
	}

	if cfg.BackupDir == "" {
		cfg.BackupDir = filepath.Join(cfg.TargetDir, ".dot-backup")
	}

	if cfg.Concurrency == 0 {
		cfg.Concurrency = runtime.NumCPU()
	}

	return cfg
}
