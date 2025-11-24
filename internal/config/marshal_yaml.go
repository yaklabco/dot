package config

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// YAMLStrategy implements Strategy for YAML format.
type YAMLStrategy struct{}

// NewYAMLStrategy creates a new YAML marshaling strategy.
func NewYAMLStrategy() *YAMLStrategy {
	return &YAMLStrategy{}
}

// Name returns "yaml".
func (s *YAMLStrategy) Name() string {
	return "yaml"
}

// Marshal converts configuration to YAML bytes.
func (s *YAMLStrategy) Marshal(cfg *ExtendedConfig, opts MarshalOptions) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("cannot marshal nil config")
	}

	if opts.IncludeComments {
		return s.marshalWithComments(cfg)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}

	return data, nil
}

// Unmarshal converts YAML bytes to configuration.
func (s *YAMLStrategy) Unmarshal(data []byte) (*ExtendedConfig, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot unmarshal empty data")
	}

	var cfg ExtendedConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return &cfg, nil
}

// marshalWithComments creates YAML with helpful comments.
func (s *YAMLStrategy) marshalWithComments(cfg *ExtendedConfig) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("# Dot Configuration File\n")
	buf.WriteString("# Documentation: https://github.com/yaklabco/dot/docs/configuration.md\n\n")

	buf.WriteString("# Core Directories\n")
	buf.WriteString("directories:\n")
	buf.WriteString("  # Package directory containing packages\n")
	buf.WriteString(fmt.Sprintf("  package: %s\n", cfg.Directories.Package))
	buf.WriteString("  # Target directory for symlinks\n")
	buf.WriteString(fmt.Sprintf("  target: %s\n", cfg.Directories.Target))
	buf.WriteString("  # Manifest directory for tracking\n")
	buf.WriteString(fmt.Sprintf("  manifest: %s\n\n", cfg.Directories.Manifest))

	buf.WriteString("# Logging Configuration\n")
	buf.WriteString("logging:\n")
	buf.WriteString("  # Log level: DEBUG, INFO, WARN, ERROR\n")
	buf.WriteString(fmt.Sprintf("  level: %s\n", cfg.Logging.Level))
	buf.WriteString("  # Log format: text, json\n")
	buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Logging.Format))
	buf.WriteString("  # Log destination: stderr, stdout, file\n")
	buf.WriteString(fmt.Sprintf("  destination: %s\n", cfg.Logging.Destination))
	buf.WriteString("  # Log file path (only used if destination is file)\n")
	buf.WriteString(fmt.Sprintf("  file: %s\n\n", cfg.Logging.File))

	buf.WriteString("# Symlink Behavior\n")
	buf.WriteString("symlinks:\n")
	buf.WriteString("  # Link mode: relative, absolute\n")
	buf.WriteString(fmt.Sprintf("  mode: %s\n", cfg.Symlinks.Mode))
	buf.WriteString("  # Enable directory folding optimization\n")
	buf.WriteString(fmt.Sprintf("  folding: %t\n", cfg.Symlinks.Folding))
	buf.WriteString("  # Overwrite existing files when conflicts occur\n")
	buf.WriteString(fmt.Sprintf("  overwrite: %t\n", cfg.Symlinks.Overwrite))
	buf.WriteString("  # Create backup of overwritten files\n")
	buf.WriteString(fmt.Sprintf("  backup: %t\n", cfg.Symlinks.Backup))
	buf.WriteString("  # Backup suffix when backups enabled\n")
	buf.WriteString(fmt.Sprintf("  backup_suffix: %s\n", cfg.Symlinks.BackupSuffix))
	buf.WriteString("  # Directory for backup files\n")
	if cfg.Symlinks.BackupDir == "" {
		buf.WriteString("  backup_dir:\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("  backup_dir: %s\n\n", cfg.Symlinks.BackupDir))
	}

	buf.WriteString("# Ignore Patterns\n")
	buf.WriteString("ignore:\n")
	buf.WriteString("  # Use default ignore patterns\n")
	buf.WriteString(fmt.Sprintf("  use_defaults: %t\n", cfg.Ignore.UseDefaults))
	buf.WriteString("  # Additional patterns to ignore (glob format)\n")
	s.writeYAMLList(&buf, "patterns", cfg.Ignore.Patterns, 2)
	buf.WriteString("  # Patterns to override (force include even if ignored)\n")
	s.writeYAMLList(&buf, "overrides", cfg.Ignore.Overrides, 2)
	buf.WriteString("\n")

	buf.WriteString("# Dotfile Translation\n")
	buf.WriteString("dotfile:\n")
	buf.WriteString("  # Enable dot- to . translation\n")
	buf.WriteString(fmt.Sprintf("  translate: %t\n", cfg.Dotfile.Translate))
	buf.WriteString("  # Prefix for dotfile translation\n")
	buf.WriteString(fmt.Sprintf("  prefix: %s\n\n", cfg.Dotfile.Prefix))

	buf.WriteString("# Output Configuration\n")
	buf.WriteString("output:\n")
	buf.WriteString("  # Default output format: text, json, yaml, table\n")
	buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Output.Format))
	buf.WriteString("  # Enable colored output: auto, always, never\n")
	buf.WriteString(fmt.Sprintf("  color: %s\n", cfg.Output.Color))
	buf.WriteString("  # Show progress indicators\n")
	buf.WriteString(fmt.Sprintf("  progress: %t\n", cfg.Output.Progress))
	buf.WriteString("  # Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)\n")
	buf.WriteString(fmt.Sprintf("  verbosity: %d\n", cfg.Output.Verbosity))
	buf.WriteString("  # Terminal width for text wrapping (0 = auto-detect)\n")
	buf.WriteString(fmt.Sprintf("  width: %d\n\n", cfg.Output.Width))

	buf.WriteString("# Operation Defaults\n")
	buf.WriteString("operations:\n")
	buf.WriteString("  # Enable dry-run mode by default\n")
	buf.WriteString(fmt.Sprintf("  dry_run: %t\n", cfg.Operations.DryRun))
	buf.WriteString("  # Enable atomic operations with rollback\n")
	buf.WriteString(fmt.Sprintf("  atomic: %t\n", cfg.Operations.Atomic))
	buf.WriteString("  # Maximum number of parallel operations (0 = auto)\n")
	buf.WriteString(fmt.Sprintf("  max_parallel: %d\n\n", cfg.Operations.MaxParallel))

	buf.WriteString("# Package Management\n")
	buf.WriteString("packages:\n")
	buf.WriteString("  # Default sort order: name, links, date\n")
	buf.WriteString(fmt.Sprintf("  sort_by: %s\n", cfg.Packages.SortBy))
	buf.WriteString("  # Automatically scan for new packages\n")
	buf.WriteString(fmt.Sprintf("  auto_discover: %t\n", cfg.Packages.AutoDiscover))
	buf.WriteString("  # Package naming convention validation\n")
	buf.WriteString(fmt.Sprintf("  validate_names: %t\n\n", cfg.Packages.ValidateNames))

	buf.WriteString("# Doctor Configuration\n")
	buf.WriteString("doctor:\n")
	buf.WriteString("  # Auto-fix issues when possible\n")
	buf.WriteString(fmt.Sprintf("  auto_fix: %t\n", cfg.Doctor.AutoFix))
	buf.WriteString("  # Check manifest integrity\n")
	buf.WriteString(fmt.Sprintf("  check_manifest: %t\n", cfg.Doctor.CheckManifest))
	buf.WriteString("  # Check for broken symlinks\n")
	buf.WriteString(fmt.Sprintf("  check_broken_links: %t\n", cfg.Doctor.CheckBrokenLinks))
	buf.WriteString("  # Check for orphaned links\n")
	buf.WriteString(fmt.Sprintf("  check_orphaned: %t\n", cfg.Doctor.CheckOrphaned))
	buf.WriteString("  # Check file permissions\n")
	buf.WriteString(fmt.Sprintf("  check_permissions: %t\n\n", cfg.Doctor.CheckPermissions))

	buf.WriteString("# Experimental Features\n")
	buf.WriteString("experimental:\n")
	buf.WriteString("  # Enable parallel operations\n")
	buf.WriteString(fmt.Sprintf("  parallel: %t\n", cfg.Experimental.Parallel))
	buf.WriteString("  # Enable performance profiling\n")
	buf.WriteString(fmt.Sprintf("  profiling: %t\n", cfg.Experimental.Profiling))

	return buf.Bytes(), nil
}

// writeYAMLList writes a YAML list with proper indentation.
func (s *YAMLStrategy) writeYAMLList(buf *bytes.Buffer, key string, items []string, indent int) {
	spaces := make([]byte, indent)
	for i := range spaces {
		spaces[i] = ' '
	}
	prefix := string(spaces)

	if len(items) == 0 {
		buf.WriteString(fmt.Sprintf("%s%s: []\n", prefix, key))
		return
	}

	buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("%s  - %s\n", prefix, item))
	}
}
