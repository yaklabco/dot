package bootstrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, cfg Config)
	}{
		{
			name: "valid minimal config",
			content: `version: "1.0"
packages:
  - name: dot-vim
    required: true
`,
			wantErr: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "1.0", cfg.Version)
				assert.Len(t, cfg.Packages, 1)
				assert.Equal(t, "dot-vim", cfg.Packages[0].Name)
				assert.True(t, cfg.Packages[0].Required)
			},
		},
		{
			name: "valid config with profiles",
			content: `version: "1.0"
packages:
  - name: dot-vim
    required: true
  - name: dot-zsh
    required: false
profiles:
  minimal:
    description: Minimal setup
    packages:
      - dot-vim
  full:
    description: Full setup
    packages:
      - dot-vim
      - dot-zsh
`,
			wantErr: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "1.0", cfg.Version)
				assert.Len(t, cfg.Packages, 2)
				assert.Len(t, cfg.Profiles, 2)
				assert.Contains(t, cfg.Profiles, "minimal")
				assert.Contains(t, cfg.Profiles, "full")
				assert.Equal(t, "Minimal setup", cfg.Profiles["minimal"].Description)
				assert.Equal(t, []string{"dot-vim"}, cfg.Profiles["minimal"].Packages)
			},
		},
		{
			name: "valid config with platform filtering",
			content: `version: "1.0"
packages:
  - name: dot-vim
    required: true
    platform:
      - linux
      - darwin
  - name: dot-wsl
    required: false
    platform:
      - linux
`,
			wantErr: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Len(t, cfg.Packages, 2)
				assert.Equal(t, []string{"linux", "darwin"}, cfg.Packages[0].Platform)
				assert.Equal(t, []string{"linux"}, cfg.Packages[1].Platform)
			},
		},
		{
			name: "valid config with conflict policies",
			content: `version: "1.0"
packages:
  - name: dot-vim
    required: true
    on_conflict: backup
  - name: dot-zsh
    on_conflict: skip
defaults:
  on_conflict: fail
`,
			wantErr: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "backup", cfg.Packages[0].ConflictPolicy)
				assert.Equal(t, "skip", cfg.Packages[1].ConflictPolicy)
				assert.Equal(t, "fail", cfg.Defaults.ConflictPolicy)
			},
		},
		{
			name: "valid config with defaults",
			content: `version: "1.0"
packages:
  - name: dot-vim
profiles:
  minimal:
    description: Minimal
    packages: [dot-vim]
defaults:
  profile: minimal
  on_conflict: backup
`,
			wantErr: false,
			validate: func(t *testing.T, cfg Config) {
				assert.Equal(t, "minimal", cfg.Defaults.Profile)
				assert.Equal(t, "backup", cfg.Defaults.ConflictPolicy)
			},
		},
		{
			name: "invalid YAML syntax",
			content: `version: "1.0"
packages:
  - name: dot-vim
    invalid: [unclosed
`,
			wantErr: true,
			errMsg:  "parse YAML",
		},
		{
			name: "invalid config - missing version",
			content: `packages:
  - name: dot-vim
`,
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "valid config - no packages",
			content: `version: "1.0"
packages: []
`,
			wantErr: false,
		},
		{
			name: "invalid config - invalid platform",
			content: `version: "1.0"
packages:
  - name: dot-vim
    platform:
      - invalid-os
`,
			wantErr: true,
			errMsg:  "invalid platform",
		},
		{
			name: "invalid config - profile references unknown package",
			content: `version: "1.0"
packages:
  - name: dot-vim
profiles:
  full:
    packages:
      - dot-vim
      - non-existent
`,
			wantErr: true,
			errMsg:  "unknown package",
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
			errMsg:  "version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fs := adapters.NewMemFS()

			// Write test config to filesystem
			configPath := "/.dotbootstrap.yaml"
			err := fs.WriteFile(ctx, configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := Load(ctx, fs, configPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	cfg, err := Load(ctx, fs, "/non-existent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read config file")
	assert.Equal(t, Config{}, cfg)
}

func TestFilterPackagesByPlatform(t *testing.T) {
	packages := []PackageSpec{
		{Name: "all-platforms"},
		{Name: "linux-only", Platform: []string{"linux"}},
		{Name: "darwin-only", Platform: []string{"darwin"}},
		{Name: "linux-darwin", Platform: []string{"linux", "darwin"}},
		{Name: "windows-only", Platform: []string{"windows"}},
	}

	tests := []struct {
		name     string
		platform string
		expected []string
	}{
		{
			name:     "linux",
			platform: "linux",
			expected: []string{"all-platforms", "linux-only", "linux-darwin"},
		},
		{
			name:     "darwin",
			platform: "darwin",
			expected: []string{"all-platforms", "darwin-only", "linux-darwin"},
		},
		{
			name:     "windows",
			platform: "windows",
			expected: []string{"all-platforms", "windows-only"},
		},
		{
			name:     "freebsd",
			platform: "freebsd",
			expected: []string{"all-platforms"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterPackagesByPlatform(packages, tt.platform)
			var names []string
			for _, pkg := range filtered {
				names = append(names, pkg.Name)
			}
			assert.Equal(t, tt.expected, names)
		})
	}
}

func TestGetPackageNames(t *testing.T) {
	config := Config{
		Version: "1.0",
		Packages: []PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
	}

	names := GetPackageNames(config)
	assert.Equal(t, []string{"dot-vim", "dot-zsh", "dot-tmux"}, names)
}

func TestGetProfile(t *testing.T) {
	config := Config{
		Version: "1.0",
		Packages: []PackageSpec{
			{Name: "dot-vim"},
			{Name: "dot-zsh"},
			{Name: "dot-tmux"},
		},
		Profiles: map[string]Profile{
			"minimal": {
				Description: "Minimal setup",
				Packages:    []string{"dot-vim"},
			},
			"full": {
				Description: "Full setup",
				Packages:    []string{"dot-vim", "dot-zsh", "dot-tmux"},
			},
		},
	}

	t.Run("existing profile", func(t *testing.T) {
		packages, err := GetProfile(config, "minimal")
		assert.NoError(t, err)
		assert.Equal(t, []string{"dot-vim"}, packages)
	})

	t.Run("non-existent profile", func(t *testing.T) {
		packages, err := GetProfile(config, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile not found")
		assert.Nil(t, packages)
	})
}
