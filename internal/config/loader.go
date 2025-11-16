package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles loading configuration from multiple sources.
type Loader struct {
	appName    string
	configPath string
}

// NewLoader creates a configuration loader.
func NewLoader(appName string, configPath string) *Loader {
	return &Loader{
		appName:    appName,
		configPath: configPath,
	}
}

// Load loads configuration from file with proper precedence.
// Precedence: file > defaults
func (l *Loader) Load() (*ExtendedConfig, error) {
	// Load from config file if it exists
	if fileExists(l.configPath) {
		fileCfg, err := LoadExtendedFromFile(l.configPath)
		if err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		// Use file config directly to preserve explicit false values
		return fileCfg, nil
	}

	// No file - return defaults
	cfg := DefaultExtended()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithEnv loads configuration from file and applies environment variable overrides.
// Precedence: env > file > defaults
func (l *Loader) LoadWithEnv() (*ExtendedConfig, error) {
	// Start with file load
	cfg, err := l.Load()
	if err != nil {
		return nil, err
	}

	// Load from environment (sparse config with only env-set values)
	envCfg := l.loadFromEnv()
	// Use simple merge for env (only strings, no booleans unless tracked)
	cfg = mergeConfigs(cfg, envCfg)

	// Validate merged configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithFlags loads configuration and applies flag overrides.
// Precedence: flags > env > file > defaults
func (l *Loader) LoadWithFlags(flags map[string]interface{}) (*ExtendedConfig, error) {
	// Load with env
	cfg, err := l.LoadWithEnv()
	if err != nil {
		return nil, err
	}

	// Apply flag overrides
	flagCfg, verbositySet := l.configFromFlags(flags)
	cfg = mergeConfigsWithVerbosity(cfg, flagCfg, verbositySet)

	// Validate again after flag overrides
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables.
// Returns a sparse config with only explicitly set environment values.
func (l *Loader) loadFromEnv() *ExtendedConfig {
	v := viper.New()

	// Set up environment variable handling
	v.SetEnvPrefix(strings.ToUpper(l.appName))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind all configuration keys
	l.bindEnvKeys(v)

	// Create sparse config
	cfg := createSparseConfig()

	// Load each section
	loadDirectoriesFromEnv(v, &cfg.Directories)
	loadLoggingFromEnv(v, &cfg.Logging)
	loadSymlinksFromEnv(v, &cfg.Symlinks)
	loadIgnoreFromEnv(v, &cfg.Ignore)
	loadDotfileFromEnv(v, &cfg.Dotfile)
	loadOutputFromEnv(v, &cfg.Output)
	loadOperationsFromEnv(v, &cfg.Operations)
	loadPackagesFromEnv(v, &cfg.Packages)
	loadDoctorFromEnv(v, &cfg.Doctor)
	loadExperimentalFromEnv(v, &cfg.Experimental)

	return cfg
}

func loadDirectoriesFromEnv(v *viper.Viper, cfg *DirectoriesConfig) {
	if v.IsSet("directories.package") {
		cfg.Package = v.GetString("directories.package")
	}
	if v.IsSet("directories.target") {
		cfg.Target = v.GetString("directories.target")
	}
	if v.IsSet("directories.manifest") {
		cfg.Manifest = v.GetString("directories.manifest")
	}
}

func loadLoggingFromEnv(v *viper.Viper, cfg *LoggingConfig) {
	if v.IsSet("logging.level") {
		cfg.Level = v.GetString("logging.level")
	}
	if v.IsSet("logging.format") {
		cfg.Format = v.GetString("logging.format")
	}
	if v.IsSet("logging.destination") {
		cfg.Destination = v.GetString("logging.destination")
	}
	if v.IsSet("logging.file") {
		cfg.File = v.GetString("logging.file")
	}
}

func loadSymlinksFromEnv(v *viper.Viper, cfg *SymlinksConfig) {
	if v.IsSet("symlinks.mode") {
		cfg.Mode = v.GetString("symlinks.mode")
	}
	if v.IsSet("symlinks.folding") {
		cfg.Folding = v.GetBool("symlinks.folding")
	}
	if v.IsSet("symlinks.overwrite") {
		cfg.Overwrite = v.GetBool("symlinks.overwrite")
	}
	if v.IsSet("symlinks.backup") {
		cfg.Backup = v.GetBool("symlinks.backup")
	}
	if v.IsSet("symlinks.backup_suffix") {
		cfg.BackupSuffix = v.GetString("symlinks.backup_suffix")
	}
}

func loadIgnoreFromEnv(v *viper.Viper, cfg *IgnoreConfig) {
	if v.IsSet("ignore.use_defaults") {
		cfg.UseDefaults = v.GetBool("ignore.use_defaults")
	}
	if v.IsSet("ignore.patterns") {
		cfg.Patterns = v.GetStringSlice("ignore.patterns")
	}
	if v.IsSet("ignore.overrides") {
		cfg.Overrides = v.GetStringSlice("ignore.overrides")
	}
	if v.IsSet("ignore.per_package_ignore") {
		cfg.PerPackageIgnore = v.GetBool("ignore.per_package_ignore")
	}
	if v.IsSet("ignore.max_file_size") {
		cfg.MaxFileSize = v.GetInt64("ignore.max_file_size")
	}
	if v.IsSet("ignore.interactive_large_files") {
		cfg.InteractiveLargeFiles = v.GetBool("ignore.interactive_large_files")
	}
}

func loadDotfileFromEnv(v *viper.Viper, cfg *DotfileConfig) {
	if v.IsSet("dotfile.translate") {
		cfg.Translate = v.GetBool("dotfile.translate")
	}
	if v.IsSet("dotfile.prefix") {
		cfg.Prefix = v.GetString("dotfile.prefix")
	}
}

func loadOutputFromEnv(v *viper.Viper, cfg *OutputConfig) {
	if v.IsSet("output.format") {
		cfg.Format = v.GetString("output.format")
	}
	if v.IsSet("output.color") {
		cfg.Color = v.GetString("output.color")
	}
	if v.IsSet("output.progress") {
		cfg.Progress = v.GetBool("output.progress")
	}
	if v.IsSet("output.verbosity") {
		cfg.Verbosity = v.GetInt("output.verbosity")
	}
	if v.IsSet("output.width") {
		cfg.Width = v.GetInt("output.width")
	}
}

func loadOperationsFromEnv(v *viper.Viper, cfg *OperationsConfig) {
	if v.IsSet("operations.dry_run") {
		cfg.DryRun = v.GetBool("operations.dry_run")
	}
	if v.IsSet("operations.atomic") {
		cfg.Atomic = v.GetBool("operations.atomic")
	}
	if v.IsSet("operations.max_parallel") {
		cfg.MaxParallel = v.GetInt("operations.max_parallel")
	}
}

func loadPackagesFromEnv(v *viper.Viper, cfg *PackagesConfig) {
	if v.IsSet("packages.sort_by") {
		cfg.SortBy = v.GetString("packages.sort_by")
	}
	if v.IsSet("packages.auto_discover") {
		cfg.AutoDiscover = v.GetBool("packages.auto_discover")
	}
	if v.IsSet("packages.validate_names") {
		cfg.ValidateNames = v.GetBool("packages.validate_names")
	}
}

func loadDoctorFromEnv(v *viper.Viper, cfg *DoctorConfig) {
	if v.IsSet("doctor.auto_fix") {
		cfg.AutoFix = v.GetBool("doctor.auto_fix")
	}
	if v.IsSet("doctor.check_manifest") {
		cfg.CheckManifest = v.GetBool("doctor.check_manifest")
	}
	if v.IsSet("doctor.check_broken_links") {
		cfg.CheckBrokenLinks = v.GetBool("doctor.check_broken_links")
	}
	if v.IsSet("doctor.check_orphaned") {
		cfg.CheckOrphaned = v.GetBool("doctor.check_orphaned")
	}
	if v.IsSet("doctor.check_permissions") {
		cfg.CheckPermissions = v.GetBool("doctor.check_permissions")
	}
}

func loadExperimentalFromEnv(v *viper.Viper, cfg *ExperimentalConfig) {
	if v.IsSet("experimental.parallel") {
		cfg.Parallel = v.GetBool("experimental.parallel")
	}
	if v.IsSet("experimental.profiling") {
		cfg.Profiling = v.GetBool("experimental.profiling")
	}
}

// getEnvWithPrefix gets an environment variable with the given prefix.
func getEnvWithPrefix(prefix, key string) string {
	return os.Getenv(prefix + key)
}

// bindEnvKeys binds all configuration keys to environment variables.
func (l *Loader) bindEnvKeys(v *viper.Viper) {
	v.BindEnv("directories.package")
	v.BindEnv("directories.target")
	v.BindEnv("directories.manifest")

	v.BindEnv("logging.level")
	v.BindEnv("logging.format")
	v.BindEnv("logging.destination")
	v.BindEnv("logging.file")

	v.BindEnv("symlinks.mode")
	v.BindEnv("symlinks.folding")
	v.BindEnv("symlinks.overwrite")
	v.BindEnv("symlinks.backup")
	v.BindEnv("symlinks.backup_suffix")

	v.BindEnv("ignore.use_defaults")
	v.BindEnv("ignore.patterns")
	v.BindEnv("ignore.overrides")
	v.BindEnv("ignore.per_package_ignore")
	v.BindEnv("ignore.max_file_size")
	v.BindEnv("ignore.interactive_large_files")

	v.BindEnv("dotfile.translate")
	v.BindEnv("dotfile.prefix")

	v.BindEnv("output.format")
	v.BindEnv("output.color")
	v.BindEnv("output.progress")
	v.BindEnv("output.verbosity")
	v.BindEnv("output.width")

	v.BindEnv("operations.dry_run")
	v.BindEnv("operations.atomic")
	v.BindEnv("operations.max_parallel")

	v.BindEnv("packages.sort_by")
	v.BindEnv("packages.auto_discover")
	v.BindEnv("packages.validate_names")

	v.BindEnv("doctor.auto_fix")
	v.BindEnv("doctor.check_manifest")
	v.BindEnv("doctor.check_broken_links")
	v.BindEnv("doctor.check_orphaned")
	v.BindEnv("doctor.check_permissions")

	v.BindEnv("experimental.parallel")
	v.BindEnv("experimental.profiling")
}

// configFromFlags creates partial config from flag map.
func (l *Loader) configFromFlags(flags map[string]interface{}) (*ExtendedConfig, bool) {
	cfg := createSparseConfig()

	verbositySet := applyFlagsToConfig(cfg, flags)

	return cfg, verbositySet
}

// createSparseConfig creates an empty config for flag/env merging.
func createSparseConfig() *ExtendedConfig {
	return &ExtendedConfig{
		Directories:  DirectoriesConfig{},
		Logging:      LoggingConfig{},
		Symlinks:     SymlinksConfig{},
		Ignore:       IgnoreConfig{},
		Dotfile:      DotfileConfig{},
		Output:       OutputConfig{Verbosity: -1}, // Use -1 as sentinel for "not set"
		Operations:   OperationsConfig{},
		Packages:     PackagesConfig{},
		Doctor:       DoctorConfig{},
		Experimental: ExperimentalConfig{},
	}
}

// applyFlagsToConfig maps command-line flags to configuration fields.
func applyFlagsToConfig(cfg *ExtendedConfig, flags map[string]interface{}) bool {
	verbositySet := false

	applyDirectoryFlags(cfg, flags)
	applyLoggingFlags(cfg, flags)
	applyOperationFlags(cfg, flags)
	verbositySet = applyOutputFlags(cfg, flags)

	return verbositySet
}

// applyDirectoryFlags applies directory-related flags.
func applyDirectoryFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["dir"].(string); ok && val != "" {
		cfg.Directories.Package = val
	}
	if val, ok := flags["target"].(string); ok && val != "" {
		cfg.Directories.Target = val
	}
}

// applyLoggingFlags applies logging-related flags.
func applyLoggingFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["log-json"].(bool); ok && val {
		cfg.Logging.Format = "json"
	}
}

// applyOperationFlags applies operation-related flags.
func applyOperationFlags(cfg *ExtendedConfig, flags map[string]interface{}) {
	if val, ok := flags["dry-run"].(bool); ok && val {
		cfg.Operations.DryRun = val
	}
}

// applyOutputFlags applies output-related flags and returns if verbosity was set.
func applyOutputFlags(cfg *ExtendedConfig, flags map[string]interface{}) bool {
	verbositySet := false

	if val, ok := flags["verbose"].(int); ok {
		cfg.Output.Verbosity = val
		verbositySet = true
	}
	if val, ok := flags["quiet"].(bool); ok && val {
		cfg.Output.Verbosity = 0
		verbositySet = true
	}
	if val, ok := flags["color"].(string); ok && val != "" {
		cfg.Output.Color = val
	}
	if val, ok := flags["format"].(string); ok && val != "" {
		cfg.Output.Format = val
	}

	return verbositySet
}

// mergeConfigs merges two configs, with override taking precedence for non-zero values.
// Only merges fields that are explicitly set in override (non-empty strings, non-zero lists).
func mergeConfigs(base, override *ExtendedConfig) *ExtendedConfig {
	return mergeConfigsWithVerbosity(base, override, false)
}

// mergeConfigsWithVerbosity merges configs with special handling for verbosity.
func mergeConfigsWithVerbosity(base, override *ExtendedConfig, verbosityExplicit bool) *ExtendedConfig {
	merged := *base

	mergeDirectories(&merged, override)
	mergeLogging(&merged, override)
	mergeSymlinks(&merged, override)
	mergeIgnore(&merged, override)
	mergeDotfile(&merged, override)
	mergeOutput(&merged, override, verbosityExplicit)
	mergeOperations(&merged, override)
	mergePackages(&merged, override)
	mergeDoctor(&merged, override)
	mergeExperimental(&merged, override)

	return &merged
}

// mergeDirectories merges directory configuration.
func mergeDirectories(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Directories.Package != "" {
		merged.Directories.Package = override.Directories.Package
	}
	if override.Directories.Target != "" {
		merged.Directories.Target = override.Directories.Target
	}
	if override.Directories.Manifest != "" {
		merged.Directories.Manifest = override.Directories.Manifest
	}
}

// mergeLogging merges logging configuration.
func mergeLogging(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Logging.Level != "" {
		merged.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		merged.Logging.Format = override.Logging.Format
	}
	if override.Logging.Destination != "" {
		merged.Logging.Destination = override.Logging.Destination
	}
	if override.Logging.File != "" {
		merged.Logging.File = override.Logging.File
	}
}

// mergeSymlinks merges symlink configuration.
func mergeSymlinks(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Symlinks.Mode != "" {
		merged.Symlinks.Mode = override.Symlinks.Mode
	}
	if override.Symlinks.BackupSuffix != "" {
		merged.Symlinks.BackupSuffix = override.Symlinks.BackupSuffix
	}
	if override.Symlinks.Overwrite {
		merged.Symlinks.Overwrite = true
	}
	if override.Symlinks.Backup {
		merged.Symlinks.Backup = true
	}
}

// mergeIgnore merges ignore pattern configuration.
func mergeIgnore(merged *ExtendedConfig, override *ExtendedConfig) {
	if len(override.Ignore.Patterns) > 0 {
		merged.Ignore.Patterns = override.Ignore.Patterns
	}
	if len(override.Ignore.Overrides) > 0 {
		merged.Ignore.Overrides = override.Ignore.Overrides
	}
}

// mergeDotfile merges dotfile translation configuration.
func mergeDotfile(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Dotfile.Prefix != "" {
		merged.Dotfile.Prefix = override.Dotfile.Prefix
	}
}

// mergeOutput merges output configuration with special verbosity handling.
func mergeOutput(merged *ExtendedConfig, override *ExtendedConfig, verbosityExplicit bool) {
	if override.Output.Format != "" {
		merged.Output.Format = override.Output.Format
	}
	if override.Output.Color != "" {
		merged.Output.Color = override.Output.Color
	}
	if verbosityExplicit || override.Output.Verbosity >= 0 {
		merged.Output.Verbosity = override.Output.Verbosity
	}
	if override.Output.Width > 0 {
		merged.Output.Width = override.Output.Width
	}
}

// mergeOperations merges operation configuration.
func mergeOperations(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Operations.DryRun {
		merged.Operations.DryRun = true
	}
	if override.Operations.MaxParallel > 0 {
		merged.Operations.MaxParallel = override.Operations.MaxParallel
	}
}

// mergePackages merges package management configuration.
func mergePackages(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Packages.SortBy != "" {
		merged.Packages.SortBy = override.Packages.SortBy
	}
}

// mergeDoctor merges doctor configuration.
func mergeDoctor(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Doctor.AutoFix {
		merged.Doctor.AutoFix = true
	}
}

// mergeExperimental merges experimental feature configuration.
func mergeExperimental(merged *ExtendedConfig, override *ExtendedConfig) {
	if override.Experimental.Parallel {
		merged.Experimental.Parallel = true
	}
	if override.Experimental.Profiling {
		merged.Experimental.Profiling = true
	}
}
