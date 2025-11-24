package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/config"
)

func TestExtendedConfig_Default(t *testing.T) {
	cfg := config.DefaultExtended()

	require.NotNil(t, cfg)

	// Directories
	assert.NotEmpty(t, cfg.Directories.Package)
	assert.NotEmpty(t, cfg.Directories.Target)
	assert.NotEmpty(t, cfg.Directories.Manifest)

	// Logging
	assert.Equal(t, "INFO", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.Equal(t, "stderr", cfg.Logging.Destination)

	// Symlinks
	assert.Equal(t, "relative", cfg.Symlinks.Mode)
	assert.True(t, cfg.Symlinks.Folding)
	assert.False(t, cfg.Symlinks.Overwrite)
	assert.False(t, cfg.Symlinks.Backup)
	assert.Equal(t, ".bak", cfg.Symlinks.BackupSuffix)

	// Ignore
	assert.True(t, cfg.Ignore.UseDefaults)
	assert.Empty(t, cfg.Ignore.Patterns)
	assert.Empty(t, cfg.Ignore.Overrides)

	// Dotfile
	assert.True(t, cfg.Dotfile.Translate)
	assert.Equal(t, "dot-", cfg.Dotfile.Prefix)
	assert.True(t, cfg.Dotfile.PackageNameMapping)

	// Output
	assert.Equal(t, "text", cfg.Output.Format)
	assert.Equal(t, "auto", cfg.Output.Color)
	assert.True(t, cfg.Output.Progress)
	assert.Equal(t, 1, cfg.Output.Verbosity)
	assert.Equal(t, 0, cfg.Output.Width)

	// Operations
	assert.False(t, cfg.Operations.DryRun)
	assert.True(t, cfg.Operations.Atomic)
	assert.Equal(t, 0, cfg.Operations.MaxParallel)

	// Packages
	assert.Equal(t, "name", cfg.Packages.SortBy)
	assert.True(t, cfg.Packages.AutoDiscover)
	assert.True(t, cfg.Packages.ValidateNames)

	// Doctor
	assert.False(t, cfg.Doctor.AutoFix)
	assert.True(t, cfg.Doctor.CheckManifest)
	assert.True(t, cfg.Doctor.CheckBrokenLinks)
	assert.True(t, cfg.Doctor.CheckOrphaned)
	assert.True(t, cfg.Doctor.CheckPermissions)

	// Update
	assert.True(t, cfg.Update.CheckOnStartup)
	assert.Equal(t, 24, cfg.Update.CheckFrequency)
	assert.Equal(t, "auto", cfg.Update.PackageManager)
	assert.Equal(t, "yaklabco/dot", cfg.Update.Repository)
	assert.False(t, cfg.Update.IncludePrerelease)

	// Experimental
	assert.False(t, cfg.Experimental.Parallel)
	assert.False(t, cfg.Experimental.Profiling)
}

func TestExtendedConfig_LoadFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
directories:
  package: /home/user/dotfiles
  target: /home/user
  manifest: /home/user/.local/share/dot/manifest

logging:
  level: DEBUG
  format: json
  destination: stderr

symlinks:
  mode: absolute
  folding: false
  overwrite: true
  backup: true
  backup_suffix: .backup

ignore:
  use_defaults: false
  patterns:
    - "*.swp"
    - "*.tmp"
  overrides:
    - ".gitignore"

dotfile:
  translate: false
  prefix: dot_
  package_name_mapping: false

output:
  format: table
  color: always
  progress: false
  verbosity: 2
  width: 120

operations:
  dry_run: true
  atomic: false
  max_parallel: 4

packages:
  sort_by: links
  auto_discover: false
  validate_names: false

doctor:
  auto_fix: true
  check_manifest: false
  check_broken_links: false
  check_orphaned: false
  check_permissions: false

experimental:
  parallel: true
  profiling: true
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	cfg, err := config.LoadExtendedFromFile(configFile)
	require.NoError(t, err)

	// Verify all values
	assert.Equal(t, "/home/user/dotfiles", cfg.Directories.Package)
	assert.Equal(t, "/home/user", cfg.Directories.Target)
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "absolute", cfg.Symlinks.Mode)
	assert.False(t, cfg.Symlinks.Folding)
	assert.True(t, cfg.Symlinks.Overwrite)
	assert.Equal(t, []string{"*.swp", "*.tmp"}, cfg.Ignore.Patterns)
	assert.False(t, cfg.Dotfile.Translate)
	assert.Equal(t, "dot_", cfg.Dotfile.Prefix)
	assert.False(t, cfg.Dotfile.PackageNameMapping)
	assert.Equal(t, "table", cfg.Output.Format)
	assert.Equal(t, 2, cfg.Output.Verbosity)
	assert.True(t, cfg.Operations.DryRun)
	assert.Equal(t, 4, cfg.Operations.MaxParallel)
	assert.Equal(t, "links", cfg.Packages.SortBy)
	assert.True(t, cfg.Doctor.AutoFix)
	assert.True(t, cfg.Experimental.Parallel)
}

func TestExtendedConfig_ValidateDirectories(t *testing.T) {
	tests := []struct {
		name    string
		config  func() *config.ExtendedConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid directories",
			config: func() *config.ExtendedConfig {
				cfg := config.DefaultExtended()
				cfg.Directories.Package = "/home/user/dotfiles"
				cfg.Directories.Target = "/home/user"
				return cfg
			},
			wantErr: false,
		},
		{
			name: "empty package directory",
			config: func() *config.ExtendedConfig {
				cfg := config.DefaultExtended()
				cfg.Directories.Package = ""
				return cfg
			},
			wantErr: true,
			errMsg:  "package directory cannot be empty",
		},
		{
			name: "empty target directory",
			config: func() *config.ExtendedConfig {
				cfg := config.DefaultExtended()
				cfg.Directories.Target = ""
				return cfg
			},
			wantErr: true,
			errMsg:  "target directory cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config()
			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateLogging(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		format  string
		dest    string
		wantErr bool
	}{
		{"valid DEBUG", "DEBUG", "text", "stderr", false},
		{"valid INFO", "INFO", "json", "stdout", false},
		{"valid WARN", "WARN", "text", "file", false},
		{"valid ERROR", "ERROR", "json", "stderr", false},
		{"invalid level", "TRACE", "text", "stderr", true},
		{"invalid format", "INFO", "xml", "stderr", true},
		{"invalid destination", "INFO", "text", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Logging.Level = tt.level
			cfg.Logging.Format = tt.format
			cfg.Logging.Destination = tt.dest

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateSymlinks(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"relative mode", "relative", false},
		{"absolute mode", "absolute", false},
		{"invalid mode", "hardlink", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Symlinks.Mode = tt.mode

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateOutput(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		color     string
		verbosity int
		width     int
		wantErr   bool
	}{
		{"valid text format", "text", "auto", 1, 0, false},
		{"valid json format", "json", "always", 2, 80, false},
		{"valid yaml format", "yaml", "never", 0, 120, false},
		{"valid table format", "table", "auto", 3, 0, false},
		{"invalid format", "xml", "auto", 1, 0, true},
		{"invalid color", "text", "invalid", 1, 0, true},
		{"invalid verbosity negative", "text", "auto", -1, 0, true},
		{"invalid verbosity too high", "text", "auto", 4, 0, true},
		{"invalid width negative", "text", "auto", 1, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Output.Format = tt.format
			cfg.Output.Color = tt.color
			cfg.Output.Verbosity = tt.verbosity
			cfg.Output.Width = tt.width

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidatePackages(t *testing.T) {
	tests := []struct {
		name    string
		sortBy  string
		wantErr bool
	}{
		{"sort by name", "name", false},
		{"sort by links", "links", false},
		{"sort by date", "date", false},
		{"invalid sort field", "size", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Packages.SortBy = tt.sortBy

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateIgnorePatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{"valid patterns", []string{"*.swp", "*.tmp", "*~"}, false},
		{"empty patterns", []string{}, false},
		{"invalid glob", []string{"[invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Ignore.Patterns = tt.patterns

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateOperations(t *testing.T) {
	cfg := config.DefaultExtended()

	// Test valid max_parallel
	cfg.Operations.MaxParallel = 0
	assert.NoError(t, cfg.Validate())

	cfg.Operations.MaxParallel = 4
	assert.NoError(t, cfg.Validate())

	// Test invalid max_parallel
	cfg.Operations.MaxParallel = -1
	assert.Error(t, cfg.Validate())
}

func TestExtendedConfig_ValidateUpdate(t *testing.T) {
	tests := []struct {
		name      string
		frequency int
		pkgMgr    string
		repo      string
		wantErr   bool
	}{
		{"valid defaults", 24, "auto", "yaklabco/dot", false},
		{"valid check frequency 0", 0, "auto", "yaklabco/dot", false},
		{"valid check frequency -1 (disabled)", -1, "auto", "yaklabco/dot", false},
		{"invalid check frequency -2", -2, "auto", "yaklabco/dot", true},
		{"valid package manager brew", 24, "brew", "yaklabco/dot", false},
		{"valid package manager apt", 24, "apt", "yaklabco/dot", false},
		{"valid package manager yum", 24, "yum", "yaklabco/dot", false},
		{"valid package manager pacman", 24, "pacman", "yaklabco/dot", false},
		{"valid package manager dnf", 24, "dnf", "yaklabco/dot", false},
		{"valid package manager zypper", 24, "zypper", "yaklabco/dot", false},
		{"valid package manager manual", 24, "manual", "yaklabco/dot", false},
		{"invalid package manager", 24, "invalid-mgr", "yaklabco/dot", true},
		{"empty repository", 24, "auto", "", true},
		{"invalid repository format (no slash)", 24, "auto", "invalid", true},
		{"invalid repository format (empty owner)", 24, "auto", "/repo", true},
		{"invalid repository format (empty repo)", 24, "auto", "owner/", true},
		{"valid different repository", 24, "auto", "owner/different-repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Update.CheckFrequency = tt.frequency
			cfg.Update.PackageManager = tt.pkgMgr
			cfg.Update.Repository = tt.repo

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_ValidateNetwork(t *testing.T) {
	tests := []struct {
		name           string
		timeout        int
		connectTimeout int
		tlsTimeout     int
		wantErr        bool
		errContains    string
	}{
		{"valid defaults (all zero)", 0, 0, 0, false, ""},
		{"valid positive timeout", 30, 10, 10, false, ""},
		{"valid large values", 300, 60, 60, false, ""},
		{"negative timeout", -1, 0, 0, true, "network.timeout must be non-negative"},
		{"negative connect_timeout", 0, -5, 0, true, "network.connect_timeout must be non-negative"},
		{"negative tls_timeout", 0, 0, -10, true, "network.tls_timeout must be non-negative"},
		{"multiple negative (timeout checked first)", -1, -5, -10, true, "network.timeout must be non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultExtended()
			cfg.Network.Timeout = tt.timeout
			cfg.Network.ConnectTimeout = tt.connectTimeout
			cfg.Network.TLSTimeout = tt.tlsTimeout

			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendedConfig_MarshalYAML(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Directories.Package = "/test/dotfiles"
	cfg.Logging.Level = "DEBUG"

	// This will be used when implementing marshaling
	// For now, verify the config is valid
	assert.NoError(t, cfg.Validate())
}

func TestExtendedConfig_MarshalJSON(t *testing.T) {
	cfg := config.DefaultExtended()
	cfg.Symlinks.Mode = "absolute"
	cfg.Output.Color = "always"

	// Verify config is valid
	assert.NoError(t, cfg.Validate())
}
