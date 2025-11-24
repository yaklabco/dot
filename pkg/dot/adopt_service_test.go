package dot

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/manifest"
)

func TestAdoptService_ResolveAdoptPath(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	exec := executor.New(executor.Opts{
		FS:     fs,
		Logger: logger,
		Tracer: adapters.NewNoopTracer(),
	})
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	targetDir := "/home/user"
	packageDir := "/home/user/dotfiles"

	svc := newAdoptService(fs, logger, exec, manifestSvc, packageDir, targetDir, false)

	// Store original cwd and restore after test
	originalCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalCwd)
		require.NoError(t, err)
	}()

	tests := []struct {
		name     string
		file     string
		cwd      string
		expected string
		wantErr  bool
	}{
		{
			name:     "absolute path",
			file:     "/etc/config",
			cwd:      "/tmp",
			expected: "/etc/config",
		},
		{
			name:     "tilde path",
			file:     "~/.vimrc",
			cwd:      "/tmp",
			expected: filepath.Join(os.Getenv("HOME"), ".vimrc"),
		},
		{
			name:     "tilde only",
			file:     "~",
			cwd:      "/tmp",
			expected: os.Getenv("HOME"),
		},
		{
			name:     "relative with dot-slash from pwd",
			file:     "./ado-cli",
			cwd:      "/home/user/.config",
			expected: "", // Will be set dynamically in test
		},
		{
			name:     "relative with parent from pwd",
			file:     "../.bashrc",
			cwd:      "/home/user/.config",
			expected: "", // Will be set dynamically in test
		},
		{
			name:     "bare path from target - backward compatible",
			file:     ".config/nvim",
			cwd:      "/tmp",
			expected: "/home/user/.config/nvim",
		},
		{
			name:     "bare filename from target",
			file:     ".vimrc",
			cwd:      "/anywhere",
			expected: "/home/user/.vimrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory and change to it
			if tt.cwd != "" {
				testDir := t.TempDir()
				cwdPath := filepath.Join(testDir, strings.TrimPrefix(tt.cwd, "/"))
				require.NoError(t, os.MkdirAll(cwdPath, 0755))
				require.NoError(t, os.Chdir(cwdPath))

				// For relative paths starting with ./ or ../, compute expected dynamically
				// using the actual resolved cwd to handle symlinks (e.g., /var -> /private/var on macOS)
				if tt.expected == "" && (strings.HasPrefix(tt.file, "./") || strings.HasPrefix(tt.file, "../")) {
					actualCwd, err := os.Getwd()
					require.NoError(t, err)
					tt.expected = filepath.Join(actualCwd, tt.file)
					tt.expected, err = filepath.Abs(tt.expected)
					require.NoError(t, err)
				}
			}

			result, err := svc.resolveAdoptPath(ctx, tt.file)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdoptService_GetManagedPaths_Empty(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	exec := executor.New(executor.Opts{
		FS:     fs,
		Logger: logger,
		Tracer: adapters.NewNoopTracer(),
	})
	manifestStore := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, manifestStore)

	targetDir := "/home/user"
	packageDir := "/home/user/dotfiles"

	svc := newAdoptService(fs, logger, exec, manifestSvc, packageDir, targetDir, false)

	// No manifest exists yet
	managedPaths, err := svc.GetManagedPaths(ctx)
	require.NoError(t, err)
	assert.Empty(t, managedPaths)
}

func TestAdoptService_GetManagedPaths_WithPackages(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	targetDir := "/home/user"
	packageDir := "/home/user/dotfiles"

	// Create directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, "bash"), 0755))
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, "vim"), 0755))

	// Create package files
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "bash", "dot-bashrc"), []byte("bashrc"), 0644))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "vim", "dot-vimrc"), []byte("vimrc"), 0644))

	// Create client and manage packages
	cfg := Config{
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
		PackageDir: packageDir,
		TargetDir:  targetDir,
	}
	cfg = cfg.WithDefaults()
	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Manage packages
	err = client.Manage(ctx, "bash", "vim")
	require.NoError(t, err)

	// Now test GetManagedPaths
	managedPaths, err := client.adoptSvc.GetManagedPaths(ctx)
	require.NoError(t, err)

	// Should have .bashrc and .vimrc as managed
	assert.True(t, managedPaths[filepath.Join(targetDir, ".bashrc")])
	assert.True(t, managedPaths[filepath.Join(targetDir, ".vimrc")])
	assert.Len(t, managedPaths, 2)
}

func TestAdoptService_GetManagedPaths_MultipleLinks(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()

	targetDir := "/home/user"
	packageDir := "/home/user/dotfiles"

	// Create directories
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
	require.NoError(t, fs.MkdirAll(ctx, filepath.Join(packageDir, "bash"), 0755))

	// Create multiple files in bash package
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "bash", "dot-bashrc"), []byte("bashrc"), 0644))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "bash", "dot-bash_profile"), []byte("profile"), 0644))
	require.NoError(t, fs.WriteFile(ctx, filepath.Join(packageDir, "bash", "dot-bash_aliases"), []byte("aliases"), 0644))

	// Create client and manage package
	cfg := Config{
		FS:         fs,
		PackageDir: packageDir,
		TargetDir:  targetDir,
		Logger:     adapters.NewNoopLogger(),
	}
	cfg = cfg.WithDefaults()
	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Manage bash package
	err = client.Manage(ctx, "bash")
	require.NoError(t, err)

	// Test GetManagedPaths
	managedPaths, err := client.adoptSvc.GetManagedPaths(ctx)
	require.NoError(t, err)

	// Should have all three bash files
	assert.True(t, managedPaths[filepath.Join(targetDir, ".bashrc")])
	assert.True(t, managedPaths[filepath.Join(targetDir, ".bash_profile")])
	assert.True(t, managedPaths[filepath.Join(targetDir, ".bash_aliases")])
	assert.Len(t, managedPaths, 3)
}
