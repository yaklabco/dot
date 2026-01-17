package dot

import (
	"fmt"
	"io"
	"os"
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

	// IgnorePatterns contains additional ignore patterns beyond defaults.
	// Supports glob patterns and negation with ! prefix.
	IgnorePatterns []string

	// UseDefaultIgnorePatterns controls whether default patterns are applied.
	// Default: true (.git, .DS_Store, etc.)
	UseDefaultIgnorePatterns bool

	// PerPackageIgnore enables reading .dotignore files from packages.
	// Default: true
	PerPackageIgnore bool

	// MaxFileSize is the maximum file size to include in bytes (0 = no limit).
	MaxFileSize int64

	// InteractiveLargeFiles enables prompting for large files in TTY mode.
	// Default: true
	InteractiveLargeFiles bool

	// Stdin is the input reader for interactive prompts.
	// Defaults to os.Stdin if nil.
	Stdin io.Reader

	// Stdout is the output writer for interactive prompts.
	// Defaults to os.Stdout if nil.
	Stdout io.Writer

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

	// Ignore configuration defaults
	// Note: UseDefaultIgnorePatterns zero value is false, but we want true as default
	// Since we can't distinguish between unset and explicitly set to false in the struct,
	// the caller should set this explicitly. For WithDefaults, we don't override.

	return cfg
}

// GetStdin returns the configured stdin or os.Stdin.
func (c *Config) GetStdin() io.Reader {
	if c.Stdin != nil {
		return c.Stdin
	}
	return os.Stdin
}

// GetStdout returns the configured stdout or os.Stdout.
func (c *Config) GetStdout() io.Writer {
	if c.Stdout != nil {
		return c.Stdout
	}
	return os.Stdout
}

// ConfigBuilder provides a fluent interface for constructing Config objects.
// It tracks which optional boolean fields have been explicitly set,
// allowing distinction between unset (use default) and explicitly set to false.
type ConfigBuilder struct {
	config Config

	// Track which optional bool fields were explicitly set
	foldingSet                  bool
	dryRunSet                   bool
	backupSet                   bool
	overwriteSet                bool
	packageNameMappingSet       bool
	useDefaultIgnorePatternsSet bool
	perPackageIgnoreSet         bool
	interactiveLargeFilesSet    bool
}

// NewConfigBuilder creates a new ConfigBuilder with zero values.
// Use the With* methods to set fields, then call Build() to get the Config.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// WithPackageDir sets the package directory.
func (b *ConfigBuilder) WithPackageDir(dir string) *ConfigBuilder {
	b.config.PackageDir = dir
	return b
}

// WithTargetDir sets the target directory.
func (b *ConfigBuilder) WithTargetDir(dir string) *ConfigBuilder {
	b.config.TargetDir = dir
	return b
}

// WithLinkMode sets the symlink creation mode.
func (b *ConfigBuilder) WithLinkMode(mode LinkMode) *ConfigBuilder {
	b.config.LinkMode = mode
	return b
}

// WithFolding sets whether directory folding is enabled.
func (b *ConfigBuilder) WithFolding(v bool) *ConfigBuilder {
	b.config.Folding = v
	b.foldingSet = true
	return b
}

// WithDryRun sets whether dry run mode is enabled.
func (b *ConfigBuilder) WithDryRun(v bool) *ConfigBuilder {
	b.config.DryRun = v
	b.dryRunSet = true
	return b
}

// WithVerbosity sets the verbosity level.
func (b *ConfigBuilder) WithVerbosity(v int) *ConfigBuilder {
	b.config.Verbosity = v
	return b
}

// WithBackupDir sets the backup directory.
func (b *ConfigBuilder) WithBackupDir(dir string) *ConfigBuilder {
	b.config.BackupDir = dir
	return b
}

// WithBackup sets whether backup is enabled.
func (b *ConfigBuilder) WithBackup(v bool) *ConfigBuilder {
	b.config.Backup = v
	b.backupSet = true
	return b
}

// WithOverwrite sets whether overwrite is enabled.
func (b *ConfigBuilder) WithOverwrite(v bool) *ConfigBuilder {
	b.config.Overwrite = v
	b.overwriteSet = true
	return b
}

// WithManifestDir sets the manifest directory.
func (b *ConfigBuilder) WithManifestDir(dir string) *ConfigBuilder {
	b.config.ManifestDir = dir
	return b
}

// WithConcurrency sets the concurrency limit.
func (b *ConfigBuilder) WithConcurrency(n int) *ConfigBuilder {
	b.config.Concurrency = n
	return b
}

// WithPackageNameMapping sets whether package name mapping is enabled.
// Default is true when not explicitly set.
func (b *ConfigBuilder) WithPackageNameMapping(v bool) *ConfigBuilder {
	b.config.PackageNameMapping = v
	b.packageNameMappingSet = true
	return b
}

// WithIgnorePatterns sets additional ignore patterns.
func (b *ConfigBuilder) WithIgnorePatterns(patterns []string) *ConfigBuilder {
	b.config.IgnorePatterns = patterns
	return b
}

// WithUseDefaultIgnorePatterns sets whether default ignore patterns are used.
// Default is true when not explicitly set.
func (b *ConfigBuilder) WithUseDefaultIgnorePatterns(v bool) *ConfigBuilder {
	b.config.UseDefaultIgnorePatterns = v
	b.useDefaultIgnorePatternsSet = true
	return b
}

// WithPerPackageIgnore sets whether per-package .dotignore files are read.
// Default is true when not explicitly set.
func (b *ConfigBuilder) WithPerPackageIgnore(v bool) *ConfigBuilder {
	b.config.PerPackageIgnore = v
	b.perPackageIgnoreSet = true
	return b
}

// WithMaxFileSize sets the maximum file size.
func (b *ConfigBuilder) WithMaxFileSize(size int64) *ConfigBuilder {
	b.config.MaxFileSize = size
	return b
}

// WithInteractiveLargeFiles sets whether to prompt for large files.
// Default is true when not explicitly set.
func (b *ConfigBuilder) WithInteractiveLargeFiles(v bool) *ConfigBuilder {
	b.config.InteractiveLargeFiles = v
	b.interactiveLargeFilesSet = true
	return b
}

// WithStdin sets the input reader.
func (b *ConfigBuilder) WithStdin(r io.Reader) *ConfigBuilder {
	b.config.Stdin = r
	return b
}

// WithStdout sets the output writer.
func (b *ConfigBuilder) WithStdout(w io.Writer) *ConfigBuilder {
	b.config.Stdout = w
	return b
}

// WithFS sets the filesystem implementation.
func (b *ConfigBuilder) WithFS(fs FS) *ConfigBuilder {
	b.config.FS = fs
	return b
}

// WithLogger sets the logger implementation.
func (b *ConfigBuilder) WithLogger(logger Logger) *ConfigBuilder {
	b.config.Logger = logger
	return b
}

// WithTracer sets the tracer implementation.
func (b *ConfigBuilder) WithTracer(tracer Tracer) *ConfigBuilder {
	b.config.Tracer = tracer
	return b
}

// WithMetrics sets the metrics implementation.
func (b *ConfigBuilder) WithMetrics(metrics Metrics) *ConfigBuilder {
	b.config.Metrics = metrics
	return b
}

// IsFoldingSet returns whether Folding was explicitly set.
func (b *ConfigBuilder) IsFoldingSet() bool {
	return b.foldingSet
}

// IsDryRunSet returns whether DryRun was explicitly set.
func (b *ConfigBuilder) IsDryRunSet() bool {
	return b.dryRunSet
}

// IsBackupSet returns whether Backup was explicitly set.
func (b *ConfigBuilder) IsBackupSet() bool {
	return b.backupSet
}

// IsOverwriteSet returns whether Overwrite was explicitly set.
func (b *ConfigBuilder) IsOverwriteSet() bool {
	return b.overwriteSet
}

// IsPackageNameMappingSet returns whether PackageNameMapping was explicitly set.
func (b *ConfigBuilder) IsPackageNameMappingSet() bool {
	return b.packageNameMappingSet
}

// IsUseDefaultIgnorePatternsSet returns whether UseDefaultIgnorePatterns was explicitly set.
func (b *ConfigBuilder) IsUseDefaultIgnorePatternsSet() bool {
	return b.useDefaultIgnorePatternsSet
}

// IsPerPackageIgnoreSet returns whether PerPackageIgnore was explicitly set.
func (b *ConfigBuilder) IsPerPackageIgnoreSet() bool {
	return b.perPackageIgnoreSet
}

// IsInteractiveLargeFilesSet returns whether InteractiveLargeFiles was explicitly set.
func (b *ConfigBuilder) IsInteractiveLargeFilesSet() bool {
	return b.interactiveLargeFilesSet
}

// Build returns the constructed Config.
// This applies defaults for optional bool fields that were not explicitly set.
func (b *ConfigBuilder) Build() Config {
	cfg := b.config

	// Apply defaults for optional bools that were not explicitly set
	// These fields default to true
	if !b.packageNameMappingSet {
		cfg.PackageNameMapping = true
	}
	if !b.useDefaultIgnorePatternsSet {
		cfg.UseDefaultIgnorePatterns = true
	}
	if !b.perPackageIgnoreSet {
		cfg.PerPackageIgnore = true
	}
	if !b.interactiveLargeFilesSet {
		cfg.InteractiveLargeFiles = true
	}

	return cfg
}

// BuildRaw returns the constructed Config without applying defaults.
// Use this when you need to inspect exactly what was set.
func (b *ConfigBuilder) BuildRaw() Config {
	return b.config
}
