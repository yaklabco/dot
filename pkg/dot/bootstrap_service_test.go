package dot

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

func TestBootstrapService_GenerateBootstrap(t *testing.T) {
	tests := []struct {
		name        string
		setupFS     func(fs FS) error
		opts        GenerateBootstrapOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "basic generation with packages",
			setupFS: func(fs FS) error {
				ctx := context.Background()
				// Create package directories
				if err := fs.MkdirAll(ctx, "/tmp/packages/vim", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/zsh", 0755); err != nil {
					return err
				}
				// Create manifest with one installed package
				manifest := `{
					"version": "1.0",
					"updated_at": "2025-10-10T00:00:00Z",
					"packages": {
						"vim": {
							"name": "vim",
							"installed_at": "2025-10-10T00:00:00Z",
							"link_count": 1,
							"links": [".vimrc"]
						}
					},
					"hashes": {}
				}`
				return fs.WriteFile(ctx, "/tmp/target/.dot-manifest.json", []byte(manifest), 0644)
			},
			opts: GenerateBootstrapOptions{
				ConflictPolicy: "backup",
			},
			wantErr: false,
		},
		{
			name: "no packages found",
			setupFS: func(fs FS) error {
				// Empty package directory
				return nil
			},
			opts:        GenerateBootstrapOptions{},
			wantErr:     true,
			errContains: "no packages",
		},
		{
			name: "from manifest only",
			setupFS: func(fs FS) error {
				ctx := context.Background()
				// Create multiple packages
				if err := fs.MkdirAll(ctx, "/tmp/packages/vim", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/zsh", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/git", 0755); err != nil {
					return err
				}
				// Create manifest with only vim and zsh installed
				manifestData := `{
  "version": "1",
  "updated_at": "2024-01-01T00:00:00Z",
  "packages": {
    "vim": {
      "name": "vim",
      "installed_at": "2024-01-01T00:00:00Z",
      "link_count": 1,
      "links": ["/home/user/.vimrc"]
    },
    "zsh": {
      "name": "zsh",
      "installed_at": "2024-01-01T00:00:00Z",
      "link_count": 1,
      "links": ["/home/user/.zshrc"]
    }
  },
  "hashes": {}
}`
				return fs.WriteFile(ctx, "/tmp/target/.dot-manifest.json", []byte(manifestData), 0644)
			},
			opts: GenerateBootstrapOptions{
				FromManifest: true,
			},
			wantErr: false,
		},
		{
			name: "from manifest filters correctly",
			setupFS: func(fs FS) error {
				ctx := context.Background()
				// Create multiple packages
				if err := fs.MkdirAll(ctx, "/tmp/packages/vim", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/zsh", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/git", 0755); err != nil {
					return err
				}
				if err := fs.MkdirAll(ctx, "/tmp/packages/tmux", 0755); err != nil {
					return err
				}
				// Create manifest with only vim installed (should filter to just vim)
				manifestData := `{
  "version": "1",
  "updated_at": "2024-01-01T00:00:00Z",
  "packages": {
    "vim": {
      "name": "vim",
      "installed_at": "2024-01-01T00:00:00Z",
      "link_count": 1,
      "links": ["/home/user/.vimrc"]
    }
  },
  "hashes": {}
}`
				return fs.WriteFile(ctx, "/tmp/target/.dot-manifest.json", []byte(manifestData), 0644)
			},
			opts: GenerateBootstrapOptions{
				FromManifest: true,
			},
			wantErr: false,
		},
		{
			name: "from manifest with no installed packages fails",
			setupFS: func(fs FS) error {
				ctx := context.Background()
				// Create packages
				if err := fs.MkdirAll(ctx, "/tmp/packages/vim", 0755); err != nil {
					return err
				}
				// Create empty manifest
				manifestData := `{
  "version": "1",
  "updated_at": "2024-01-01T00:00:00Z",
  "packages": {},
  "hashes": {}
}`
				return fs.WriteFile(ctx, "/tmp/target/.dot-manifest.json", []byte(manifestData), 0644)
			},
			opts: GenerateBootstrapOptions{
				FromManifest: true,
			},
			wantErr:     true,
			errContains: "no packages to include",
		},
		{
			name: "invalid conflict policy",
			setupFS: func(fs FS) error {
				ctx := context.Background()
				return fs.MkdirAll(ctx, "/tmp/packages/vim", 0755)
			},
			opts: GenerateBootstrapOptions{
				ConflictPolicy: "invalid",
			},
			wantErr:     true,
			errContains: "invalid conflict policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup filesystem
			fs := adapters.NewMemFS()
			ctx := context.Background()
			require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages", 0755))
			require.NoError(t, fs.MkdirAll(ctx, "/tmp/target", 0755))

			if tt.setupFS != nil {
				require.NoError(t, tt.setupFS(fs))
			}

			// Create logger
			logger := adapters.NewNoopLogger()

			// Create service
			svc := newBootstrapService(fs, logger, "/tmp/packages", "/tmp/target")

			// Execute
			result, err := svc.GenerateBootstrap(ctx, tt.opts)

			// Verify
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result.Config)
			assert.NotEmpty(t, result.YAML)
			assert.Equal(t, "1.0", result.Config.Version)
			assert.NotEmpty(t, result.Config.Packages)

			// Test-specific assertions for manifest filtering
			verifyManifestFiltering(t, tt.name, result)
		})
	}
}

// verifyManifestFiltering performs test-specific assertions for manifest filtering tests.
func verifyManifestFiltering(t *testing.T, testName string, result BootstrapResult) {
	switch testName {
	case "from manifest only":
		// Manifest had vim and zsh installed out of 3 packages (vim, zsh, git)
		assert.Equal(t, 2, result.PackageCount, "expected 2 packages from manifest")
		assert.Equal(t, 2, len(result.Config.Packages), "config should have 2 packages")
		packageNames := make([]string, len(result.Config.Packages))
		for i, pkg := range result.Config.Packages {
			packageNames[i] = pkg.Name
		}
		assert.Contains(t, packageNames, "vim", "vim should be included from manifest")
		assert.Contains(t, packageNames, "zsh", "zsh should be included from manifest")
		assert.NotContains(t, packageNames, "git", "git should NOT be included (not in manifest)")

	case "from manifest filters correctly":
		// Manifest had only vim installed out of 4 packages (vim, zsh, git, tmux)
		assert.Equal(t, 1, result.PackageCount, "expected 1 package from manifest")
		assert.Equal(t, 1, len(result.Config.Packages), "config should have 1 package")
		assert.Equal(t, "vim", result.Config.Packages[0].Name, "only vim should be included")
	}
}

func TestBootstrapService_WriteBootstrap(t *testing.T) {
	// Setup
	fs := adapters.NewMemFS()
	ctx := context.Background()
	logger := adapters.NewNoopLogger()

	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages/vim", 0755))

	svc := newBootstrapService(fs, logger, "/tmp/packages", "/tmp/target")

	// Generate config
	result, err := svc.GenerateBootstrap(ctx, GenerateBootstrapOptions{})
	require.NoError(t, err)

	// Write to custom path
	outPath := "/tmp/packages/.dotbootstrap.yaml"
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	require.NoError(t, err)

	// Verify file exists
	exists := fs.Exists(ctx, outPath)
	assert.True(t, exists)

	// Verify content
	data, err := fs.ReadFile(ctx, outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's valid YAML by loading it
	_, err = bootstrap.Load(ctx, fs, outPath)
	assert.NoError(t, err)
}

func TestBootstrapService_WriteBootstrap_FileExists(t *testing.T) {
	// Setup
	fs := adapters.NewMemFS()
	ctx := context.Background()
	logger := adapters.NewNoopLogger()

	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages/vim", 0755))

	svc := newBootstrapService(fs, logger, "/tmp/packages", "/tmp/target")

	// Generate config
	result, err := svc.GenerateBootstrap(ctx, GenerateBootstrapOptions{})
	require.NoError(t, err)

	// Write file
	outPath := "/tmp/packages/.dotbootstrap.yaml"
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	require.NoError(t, err)

	// Try to write again without force
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestBootstrapService_WriteBootstrap_WithForce(t *testing.T) {
	// Force flag is handled at CLI layer, not service layer
	// Service always returns error if file exists
	// This test verifies that behavior

	// Setup
	fs := adapters.NewMemFS()
	ctx := context.Background()
	logger := adapters.NewNoopLogger()

	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/tmp/packages/vim", 0755))

	svc := newBootstrapService(fs, logger, "/tmp/packages", "/tmp/target")

	// Generate config
	result, err := svc.GenerateBootstrap(ctx, GenerateBootstrapOptions{})
	require.NoError(t, err)

	// Write file
	outPath := "/tmp/packages/.dotbootstrap.yaml"
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	require.NoError(t, err)

	// Write again should fail - force is handled by CLI layer deleting file first
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestBootstrapService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Create temp directory
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	// Setup directories
	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "vim"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(packageDir, "zsh"), 0755))

	// Create real filesystem adapter
	fs := adapters.NewOSFilesystem()
	logger := adapters.NewNoopLogger()

	// Create service
	svc := newBootstrapService(fs, logger, packageDir, targetDir)

	// Generate bootstrap
	ctx := context.Background()
	result, err := svc.GenerateBootstrap(ctx, GenerateBootstrapOptions{
		IncludeComments: true,
		ConflictPolicy:  "backup",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.YAML)

	// Write to file
	outPath := filepath.Join(packageDir, ".dotbootstrap.yaml")
	err = svc.WriteBootstrap(ctx, result.YAML, outPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(outPath)
	require.NoError(t, err)

	// Verify content is valid
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "version:")
	assert.Contains(t, string(data), "packages:")
}
