package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/bootstrap"
)

// TestCloneCommand_EndToEnd tests the complete clone workflow using the public Client API.
// Note: This is a placeholder for full end-to-end testing which requires:
// - Actual git repository (or mock git cloner)
// - Full client setup with all dependencies
// - Network access for real cloning tests
//
// The TestBootstrapConfigLoading tests below provide comprehensive integration testing
// for the bootstrap configuration aspect of the clone feature.
func TestCloneCommand_EndToEnd(t *testing.T) {
	t.Skip("placeholder for future end-to-end clone testing with real git repository")

	// Future implementation would:
	// 1. Clone a test repository (e.g., github.com/yaklabco/dot-test-fixture)
	// 2. Verify bootstrap config is loaded
	// 3. Verify packages are installed according to profile
	// 4. Verify manifest contains repository tracking info
	// 5. Verify symlinks are created correctly
}

// setupMockRepository creates a mock dotfiles repository structure.
func setupMockRepository(t *testing.T, repoDir string) error {
	t.Helper()

	// Create .dotbootstrap.yaml
	bootstrapContent := `version: "1.0"

packages:
  - name: dot-vim
    required: true

  - name: dot-zsh
    required: false

profiles:
  minimal:
    description: "Minimal setup"
    packages:
      - dot-vim

defaults:
  profile: minimal
`
	bootstrapPath := filepath.Join(repoDir, ".dotbootstrap.yaml")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(bootstrapPath, []byte(bootstrapContent), 0644); err != nil {
		return err
	}

	// Create dot-vim package
	vimDir := filepath.Join(repoDir, "dot-vim")
	if err := os.MkdirAll(vimDir, 0755); err != nil {
		return err
	}
	vimrc := filepath.Join(vimDir, "vimrc")
	if err := os.WriteFile(vimrc, []byte("\" Test vimrc\nset number\n"), 0644); err != nil {
		return err
	}

	// Create dot-zsh package
	zshDir := filepath.Join(repoDir, "dot-zsh")
	if err := os.MkdirAll(zshDir, 0755); err != nil {
		return err
	}
	zshrc := filepath.Join(zshDir, "zshrc")
	if err := os.WriteFile(zshrc, []byte("# Test zshrc\n"), 0644); err != nil {
		return err
	}

	return nil
}

// fileExistsHelper checks if a file exists.
func fileExistsHelper(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// TestBootstrapConfigLoading tests loading and parsing bootstrap configs.
func TestBootstrapConfigLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name      string
		fixture   string
		wantError bool
	}{
		{
			name:      "minimal valid config",
			fixture:   "minimal.yaml",
			wantError: false,
		},
		{
			name:      "config with profiles",
			fixture:   "with-profiles.yaml",
			wantError: false,
		},
		{
			name:      "platform-specific config",
			fixture:   "platform-specific.yaml",
			wantError: false,
		},
		{
			name:      "invalid YAML syntax",
			fixture:   "invalid-syntax.yaml",
			wantError: true,
		},
		{
			name:      "missing required version",
			fixture:   "invalid-missing-version.yaml",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "fixtures", "bootstrap-configs", tt.fixture)
			data, err := os.ReadFile(fixturePath)
			require.NoError(t, err)

			ctx := context.Background()
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ".dotbootstrap.yaml")

			err = os.WriteFile(configPath, data, 0644)
			require.NoError(t, err)

			fs := adapters.NewOSFilesystem()
			config, err := bootstrap.Load(ctx, fs, configPath)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, config.Version)
			}
		})
	}
}
