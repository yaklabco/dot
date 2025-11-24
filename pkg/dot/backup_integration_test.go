package dot

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
)

// testEnv is a simple test environment helper to avoid import cycles with testutil
type testEnv struct {
	PackageDir string
	TargetDir  string
	ctx        context.Context
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	tmpDir := t.TempDir()
	packageDir := filepath.Join(tmpDir, "packages")
	targetDir := filepath.Join(tmpDir, "target")

	require.NoError(t, os.MkdirAll(packageDir, 0755))
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	return &testEnv{
		PackageDir: packageDir,
		TargetDir:  targetDir,
		ctx:        context.Background(),
	}
}

func (e *testEnv) Context() context.Context {
	return e.ctx
}

func (e *testEnv) CreatePackage(name string, files map[string]string) {
	pkgPath := filepath.Join(e.PackageDir, name)
	if err := os.MkdirAll(pkgPath, 0755); err != nil {
		panic(err)
	}
	for filename, content := range files {
		filePath := filepath.Join(pkgPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			panic(err)
		}
	}
}

// TestManageService_BackupIntegrityFlow tests the full workflow of backup creation
func TestManageService_BackupIntegrityFlow(t *testing.T) {
	env := newTestEnv(t)
	ctx := env.Context()

	backupDir := filepath.Join(t.TempDir(), "backup")
	require.NoError(t, os.MkdirAll(backupDir, 0755))

	testCases := []struct {
		name              string
		packageName       string
		packageFiles      map[string]string // relative path -> content
		conflictingFiles  map[string]string // target path -> content
		expectBackupCount int
	}{
		{
			name:        "single file backup",
			packageName: "vim",
			packageFiles: map[string]string{
				"dot-vimrc": "new vimrc configuration",
			},
			conflictingFiles: map[string]string{
				".vimrc": "existing vimrc content",
			},
			expectBackupCount: 1,
		},
		{
			name:        "multiple files backup",
			packageName: "bash",
			packageFiles: map[string]string{
				"dot-bashrc":  "new bashrc",
				"dot-profile": "new profile",
			},
			conflictingFiles: map[string]string{
				".bashrc":  "existing bashrc",
				".profile": "existing profile",
			},
			expectBackupCount: 2,
		},
		{
			name:        "unicode content",
			packageName: "config",
			packageFiles: map[string]string{
				"dot-unicode": "Hello 世界 🌍",
			},
			conflictingFiles: map[string]string{
				".unicode": "Original 世界",
			},
			expectBackupCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create package with helper
			env.CreatePackage(tc.packageName, tc.packageFiles)

			// Create conflicting target files with checksums
			conflictChecksums := make(map[string][32]byte)
			for name, content := range tc.conflictingFiles {
				targetPath := filepath.Join(env.TargetDir, name)
				require.NoError(t, os.WriteFile(targetPath, []byte(content), 0644))
				conflictChecksums[name] = sha256.Sum256([]byte(content))
			}

			// Configure client with backup enabled
			cfg := Config{
				PackageDir:         env.PackageDir,
				TargetDir:          env.TargetDir,
				BackupDir:          backupDir,
				ManifestDir:        env.TargetDir, // Store manifest in target directory
				Backup:             true,
				Overwrite:          false,
				DryRun:             false,
				Verbosity:          0,
				PackageNameMapping: false, // Use legacy behavior for tests
				FS:                 adapters.NewOSFilesystem(),
				Logger:             adapters.NewNoopLogger(),
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)

			// Execute manage
			err = client.Manage(ctx, tc.packageName)
			require.NoError(t, err, "Manage should succeed")

			// Load manifest using manifest service (same config as client)
			targetPathResult := NewTargetPath(env.TargetDir)
			require.True(t, targetPathResult.IsOk())
			manifestStore := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), env.TargetDir)
			manifestService := newManifestService(
				adapters.NewOSFilesystem(),
				adapters.NewNoopLogger(),
				manifestStore,
			)
			manifestResult := manifestService.Load(ctx, targetPathResult.Unwrap())
			require.True(t, manifestResult.IsOk(), "should load manifest")
			m := manifestResult.Unwrap()

			// Verify package was installed
			pkg, exists := m.GetPackage(tc.packageName)
			require.True(t, exists, "package should be in manifest")
			assert.Equal(t, tc.expectBackupCount, len(pkg.Backups), "should have correct number of backups")

			// Verify each backup integrity
			for name, originalContent := range tc.conflictingFiles {
				fullTargetPath := filepath.Join(env.TargetDir, name)

				// Check if backup exists in manifest
				backupPath, hasBackup := pkg.Backups[fullTargetPath]
				require.True(t, hasBackup, "backup should be tracked for %s", name)

				// Verify backup file exists
				_, err := os.Stat(backupPath)
				assert.NoError(t, err, "backup file should exist: %s", backupPath)

				// Read backup content
				backupData, err := os.ReadFile(backupPath)
				require.NoError(t, err, "should read backup file")

				// Verify byte-for-byte equality
				assert.Equal(t, []byte(originalContent), backupData,
					"backup must contain exact original content for %s", name)

				// Verify checksum
				expectedHash := conflictChecksums[name]
				actualHash := sha256.Sum256(backupData)
				assert.Equal(t, expectedHash, actualHash,
					"backup checksum must match original for %s", name)

				// Verify symlink was created
				linkTarget, err := os.Readlink(fullTargetPath)
				require.NoError(t, err, "symlink should exist at %s", fullTargetPath)
				assert.Contains(t, linkTarget, tc.packageName, "symlink should point to package")
			}

			// Cleanup for next test
			pkgPath := filepath.Join(env.PackageDir, tc.packageName)
			require.NoError(t, os.RemoveAll(pkgPath))
			for name := range tc.conflictingFiles {
				fullPath := filepath.Join(env.TargetDir, name)
				os.Remove(fullPath) // Ignore errors, might be symlink already removed
			}
			// Clean backup files
			require.NoError(t, os.RemoveAll(backupDir))
			require.NoError(t, os.MkdirAll(backupDir, 0755))
		})
	}
}

// TestManageService_MultipleBackupsIntegrity tests multiple packages with conflicts
func TestManageService_MultipleBackupsIntegrity(t *testing.T) {
	env := newTestEnv(t)
	ctx := env.Context()

	backupDir := filepath.Join(t.TempDir(), "backup")
	require.NoError(t, os.MkdirAll(backupDir, 0755))

	// Create 7 packages with unique random content
	numPackages := 7
	packages := make([]struct {
		name            string
		conflictContent []byte
		checksum        [32]byte
	}, numPackages)

	for i := 0; i < numPackages; i++ {
		pkgName := fmt.Sprintf("pkg%d", i)
		packages[i].name = pkgName

		// Generate random content for conflict
		content := make([]byte, 2048)
		_, err := rand.Read(content)
		require.NoError(t, err)
		packages[i].conflictContent = content
		packages[i].checksum = sha256.Sum256(content)

		// Create package with helper
		env.CreatePackage(pkgName, map[string]string{
			fmt.Sprintf("dot-config%d", i): fmt.Sprintf("package %d content", i),
		})

		// Create conflicting target file
		targetFile := filepath.Join(env.TargetDir, fmt.Sprintf(".config%d", i))
		require.NoError(t, os.WriteFile(targetFile, content, 0644))
	}

	// Configure client with backup enabled
	cfg := Config{
		PackageDir:         env.PackageDir,
		TargetDir:          env.TargetDir,
		BackupDir:          backupDir,
		ManifestDir:        env.TargetDir, // Store manifest in target directory
		Backup:             true,
		Overwrite:          false,
		DryRun:             false,
		Verbosity:          0,
		PackageNameMapping: false,
		FS:                 adapters.NewOSFilesystem(),
		Logger:             adapters.NewNoopLogger(),
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Install all packages
	packageNames := make([]string, numPackages)
	for i := 0; i < numPackages; i++ {
		packageNames[i] = packages[i].name
	}

	err = client.Manage(ctx, packageNames...)
	require.NoError(t, err, "should install all packages with backups")

	// Verify each package's backup integrity
	targetPathResult := NewTargetPath(env.TargetDir)
	require.True(t, targetPathResult.IsOk())
	manifestStore := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), env.TargetDir)
	manifestService := newManifestService(
		adapters.NewOSFilesystem(),
		adapters.NewNoopLogger(),
		manifestStore,
	)
	manifestResult := manifestService.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	for i, pkg := range packages {
		pkgInfo, exists := m.GetPackage(pkg.name)
		require.True(t, exists, "package %s should exist in manifest", pkg.name)

		targetFile := filepath.Join(env.TargetDir, fmt.Sprintf(".config%d", i))
		backupPath, hasBackup := pkgInfo.Backups[targetFile]
		require.True(t, hasBackup, "package %s should have backup", pkg.name)

		// Verify backup exists
		_, err := os.Stat(backupPath)
		assert.NoError(t, err, "backup should exist for %s", pkg.name)

		// Verify content integrity (no cross-contamination)
		backupData, err := os.ReadFile(backupPath)
		require.NoError(t, err)

		backupHash := sha256.Sum256(backupData)
		assert.Equal(t, pkg.checksum, backupHash,
			"package %s backup must contain correct content (no cross-contamination)", pkg.name)

		// Verify exact byte equality
		assert.Equal(t, pkg.conflictContent, backupData,
			"package %s backup must match original exactly", pkg.name)
	}
}

// TestManageService_BackupWithOverwrite tests that Overwrite=true skips backups
func TestManageService_BackupWithOverwrite(t *testing.T) {
	env := newTestEnv(t)
	ctx := env.Context()

	backupDir := filepath.Join(t.TempDir(), "backup")
	require.NoError(t, os.MkdirAll(backupDir, 0755))

	// Create package with helper
	env.CreatePackage("vim", map[string]string{
		"dot-vimrc": "new",
	})

	// Create conflicting file
	conflictPath := filepath.Join(env.TargetDir, ".vimrc")
	require.NoError(t, os.WriteFile(conflictPath, []byte("old"), 0644))

	// Configure with Overwrite=true (should skip backup)
	cfg := Config{
		PackageDir:         env.PackageDir,
		TargetDir:          env.TargetDir,
		BackupDir:          backupDir,
		ManifestDir:        env.TargetDir, // Store manifest in target directory
		Backup:             false,
		Overwrite:          true, // Overwrite takes precedence
		DryRun:             false,
		Verbosity:          0,
		PackageNameMapping: false,
		FS:                 adapters.NewOSFilesystem(),
		Logger:             adapters.NewNoopLogger(),
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	err = client.Manage(ctx, "vim")
	require.NoError(t, err)

	// Load manifest and verify NO backups were created
	targetPathResult := NewTargetPath(env.TargetDir)
	require.True(t, targetPathResult.IsOk())
	manifestStore := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), env.TargetDir)
	manifestService := newManifestService(
		adapters.NewOSFilesystem(),
		adapters.NewNoopLogger(),
		manifestStore,
	)
	manifestResult := manifestService.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	pkg, exists := m.GetPackage("vim")
	require.True(t, exists)

	// With Overwrite policy, no backups should be created
	assert.Empty(t, pkg.Backups, "Overwrite policy should not create backups")

	// But symlink should exist
	linkTarget, err := os.Readlink(conflictPath)
	require.NoError(t, err)
	assert.Contains(t, linkTarget, "vim", "symlink should point to package")
}

// TestBackupIntegrity_PermissionPreservation tests that permissions are preserved in backups
func TestBackupIntegrity_PermissionPreservation(t *testing.T) {
	env := newTestEnv(t)
	ctx := env.Context()

	backupDir := filepath.Join(t.TempDir(), "backup")
	require.NoError(t, os.MkdirAll(backupDir, 0755))

	// Create package with helper
	env.CreatePackage("secure", map[string]string{
		"dot-ssh-config": "new ssh config",
	})

	// Create conflicting file with restrictive permissions
	conflictPath := filepath.Join(env.TargetDir, ".ssh-config")
	require.NoError(t, os.WriteFile(conflictPath, []byte("old ssh config"), 0644))
	require.NoError(t, os.Chmod(conflictPath, 0600))

	// Get original permissions
	origInfo, err := os.Stat(conflictPath)
	require.NoError(t, err)

	cfg := Config{
		PackageDir:         env.PackageDir,
		TargetDir:          env.TargetDir,
		BackupDir:          backupDir,
		ManifestDir:        env.TargetDir, // Store manifest in target directory
		Backup:             true,
		Overwrite:          false,
		DryRun:             false,
		Verbosity:          0,
		PackageNameMapping: false,
		FS:                 adapters.NewOSFilesystem(),
		Logger:             adapters.NewNoopLogger(),
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	err = client.Manage(ctx, "secure")
	require.NoError(t, err)

	// Find backup path from manifest
	targetPathResult := NewTargetPath(env.TargetDir)
	require.True(t, targetPathResult.IsOk())
	manifestStore := manifest.NewFSManifestStoreWithDir(adapters.NewOSFilesystem(), env.TargetDir)
	manifestService := newManifestService(
		adapters.NewOSFilesystem(),
		adapters.NewNoopLogger(),
		manifestStore,
	)
	manifestResult := manifestService.Load(ctx, targetPathResult.Unwrap())
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	pkg, exists := m.GetPackage("secure")
	require.True(t, exists)

	backupPath, hasBackup := pkg.Backups[conflictPath]
	require.True(t, hasBackup)

	// Verify backup has same permissions
	backupInfo, err := os.Stat(backupPath)
	require.NoError(t, err)

	assert.Equal(t, origInfo.Mode(), backupInfo.Mode(),
		"backup must preserve original file permissions")
}
