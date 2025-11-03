package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal config",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim", Required: true},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with profiles",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim", Required: true},
					{Name: "dot-zsh", Required: false},
				},
				Profiles: map[string]Profile{
					"minimal": {
						Description: "Minimal setup",
						Packages:    []string{"dot-vim"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with platform filtering",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{
						Name:     "dot-vim",
						Required: true,
						Platform: []string{"linux", "darwin"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with conflict policy",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{
						Name:           "dot-vim",
						Required:       true,
						ConflictPolicy: "backup",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim", Required: true},
				},
				Defaults: Defaults{
					ConflictPolicy: "skip",
					Profile:        "minimal",
				},
				Profiles: map[string]Profile{
					"minimal": {
						Description: "Minimal setup",
						Packages:    []string{"dot-vim"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing version",
			config:  Config{},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "empty version",
			config: Config{
				Version: "",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "no packages allowed",
			config: Config{
				Version:  "1.0",
				Packages: []PackageSpec{},
			},
			wantErr: false,
		},
		{
			name: "invalid platform",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{
						Name:     "dot-vim",
						Platform: []string{"invalid-os"},
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid platform",
		},
		{
			name: "invalid conflict policy in package",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{
						Name:           "dot-vim",
						ConflictPolicy: "invalid-policy",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid conflict policy",
		},
		{
			name: "invalid conflict policy in defaults",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim"},
				},
				Defaults: Defaults{
					ConflictPolicy: "invalid-policy",
				},
			},
			wantErr: true,
			errMsg:  "invalid conflict policy",
		},
		{
			name: "profile references non-existent package",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim"},
				},
				Profiles: map[string]Profile{
					"full": {
						Description: "Full setup",
						Packages:    []string{"dot-vim", "non-existent"},
					},
				},
			},
			wantErr: true,
			errMsg:  "unknown package",
		},
		{
			name: "default profile does not exist",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim"},
				},
				Defaults: Defaults{
					Profile: "non-existent",
				},
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name: "package with empty name",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: ""},
				},
			},
			wantErr: true,
			errMsg:  "package name cannot be empty",
		},
		{
			name: "duplicate package names",
			config: Config{
				Version: "1.0",
				Packages: []PackageSpec{
					{Name: "dot-vim"},
					{Name: "dot-vim"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate package name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
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

func TestValidPlatforms(t *testing.T) {
	validPlatforms := []string{"linux", "darwin", "windows", "freebsd"}
	for _, platform := range validPlatforms {
		t.Run(platform, func(t *testing.T) {
			assert.True(t, isValidPlatform(platform), "platform %s should be valid", platform)
		})
	}

	invalidPlatforms := []string{"macos", "osx", "win32", "ubuntu", ""}
	for _, platform := range invalidPlatforms {
		t.Run("invalid_"+platform, func(t *testing.T) {
			assert.False(t, isValidPlatform(platform), "platform %s should be invalid", platform)
		})
	}
}

func TestValidConflictPolicies(t *testing.T) {
	validPolicies := []string{"fail", "backup", "overwrite", "skip"}
	for _, policy := range validPolicies {
		t.Run(policy, func(t *testing.T) {
			assert.True(t, isValidConflictPolicy(policy), "policy %s should be valid", policy)
		})
	}

	invalidPolicies := []string{"ignore", "abort", "delete", ""}
	for _, policy := range invalidPolicies {
		t.Run("invalid_"+policy, func(t *testing.T) {
			assert.False(t, isValidConflictPolicy(policy), "policy %s should be invalid", policy)
		})
	}
}
