package dot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
)

// Test extractBackupsFromOperations functionality
func TestManifestService_ExtractBackupsFromOperations(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	svc := newManifestService(fs, logger, store)

	t.Run("extracts backup from FileBackup operations", func(t *testing.T) {
		source := MustParsePath("/home/user/.vimrc")
		backup := MustParsePath("/backup/.vimrc.20060102-150405")

		backupOp := NewFileBackup("bak1", source, backup)
		ops := []Operation{backupOp}

		backups := svc.extractBackupsFromOperations(ops)

		assert.Len(t, backups, 1, "should extract 1 backup")
		assert.Equal(t, backup.String(), backups[source.String()], "should map source to backup path")
	})

	t.Run("ignores non-backup operations", func(t *testing.T) {
		sourcePath := MustParsePath("/packages/vim/.vimrc")
		targetPath := MustParseTargetPath("/home/user/.vimrc")

		linkOp := NewLinkCreate("link1", sourcePath, targetPath)
		deleteOp := NewFileDelete("del1", MustParsePath("/home/user/.bashrc"))
		ops := []Operation{linkOp, deleteOp}

		backups := svc.extractBackupsFromOperations(ops)

		assert.Empty(t, backups, "should not extract backups from non-FileBackup operations")
	})

	t.Run("handles multiple backups", func(t *testing.T) {
		source1 := MustParsePath("/home/user/.vimrc")
		backup1 := MustParsePath("/backup/.vimrc.20060102-150405")
		source2 := MustParsePath("/home/user/.bashrc")
		backup2 := MustParsePath("/backup/.bashrc.20060102-150406")

		ops := []Operation{
			NewFileBackup("bak1", source1, backup1),
			NewFileBackup("bak2", source2, backup2),
		}

		backups := svc.extractBackupsFromOperations(ops)

		assert.Len(t, backups, 2, "should extract 2 backups")
		assert.Equal(t, backup1.String(), backups[source1.String()])
		assert.Equal(t, backup2.String(), backups[source2.String()])
	})

	t.Run("handles mixed operations", func(t *testing.T) {
		source := MustParsePath("/home/user/.vimrc")
		backup := MustParsePath("/backup/.vimrc.20060102-150405")
		linkSource := MustParsePath("/packages/vim/.vimrc")
		linkTarget := MustParseTargetPath("/home/user/.vimrc")

		ops := []Operation{
			NewFileBackup("bak1", source, backup),
			NewLinkCreate("link1", linkSource, linkTarget),
			NewFileDelete("del1", source),
		}

		backups := svc.extractBackupsFromOperations(ops)

		assert.Len(t, backups, 1, "should extract only FileBackup operations")
		assert.Equal(t, backup.String(), backups[source.String()])
	})
}

// Test that manifest update tracks backups
func TestManifestService_UpdateTracksBackups(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	svc := newManifestService(fs, logger, store)

	targetDir := "/home/user"
	packageDir := "/packages"
	targetPathResult := NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	t.Run("records backups in manifest", func(t *testing.T) {
		source := MustParsePath("/home/user/.vimrc")
		backup := MustParsePath("/backup/.vimrc.20060102-150405")
		linkSource := MustParsePath("/packages/vim/.vimrc")
		linkTarget := MustParseTargetPath("/home/user/.vimrc")

		backupOp := NewFileBackup("bak1", source, backup)
		linkOp := NewLinkCreate("link1", linkSource, linkTarget)

		plan := Plan{
			Operations: []Operation{backupOp, linkOp},
			PackageOperations: map[string][]OperationID{
				"vim": {"bak1", "link1"},
			},
		}

		err := svc.Update(ctx, targetPath, packageDir, []string{"vim"}, plan)
		require.NoError(t, err)

		// Load manifest and check backups
		manifestResult := svc.Load(ctx, targetPath)
		require.True(t, manifestResult.IsOk())
		m := manifestResult.Unwrap()

		pkg, exists := m.GetPackage("vim")
		require.True(t, exists, "package should exist in manifest")

		assert.NotNil(t, pkg.Backups, "backups should be tracked")
		assert.Len(t, pkg.Backups, 1, "should have 1 backup")
		assert.Equal(t, backup.String(), pkg.Backups[source.String()], "should map source to backup path")
	})

	t.Run("handles package without backups", func(t *testing.T) {
		linkSource := MustParsePath("/packages/bash/.bashrc")
		linkTarget := MustParseTargetPath("/home/user/.bashrc")

		linkOp := NewLinkCreate("link1", linkSource, linkTarget)

		plan := Plan{
			Operations: []Operation{linkOp},
			PackageOperations: map[string][]OperationID{
				"bash": {"link1"},
			},
		}

		err := svc.Update(ctx, targetPath, packageDir, []string{"bash"}, plan)
		require.NoError(t, err)

		// Load manifest and check backups
		manifestResult := svc.Load(ctx, targetPath)
		require.True(t, manifestResult.IsOk())
		m := manifestResult.Unwrap()

		pkg, exists := m.GetPackage("bash")
		require.True(t, exists, "package should exist in manifest")

		// Backups map might be nil or empty for packages without backups
		if pkg.Backups != nil {
			assert.Empty(t, pkg.Backups, "should have no backups")
		}
	})
}

// Test backup tracking with UpdateWithSource
func TestManifestService_UpdateWithSourceTracksBackups(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	svc := newManifestService(fs, logger, store)

	targetDir := "/home/user"
	packageDir := "/packages"
	targetPathResult := NewTargetPath(targetDir)
	require.True(t, targetPathResult.IsOk())
	targetPath := targetPathResult.Unwrap()

	source := MustParsePath("/home/user/.config")
	backup := MustParsePath("/backup/.config.20060102-150405")
	linkSource := MustParsePath("/packages/config/.config")
	linkTarget := MustParseTargetPath("/home/user/.config")

	backupOp := NewFileBackup("bak1", source, backup)
	linkOp := NewLinkCreate("link1", linkSource, linkTarget)

	plan := Plan{
		Operations: []Operation{backupOp, linkOp},
		PackageOperations: map[string][]OperationID{
			"config": {"bak1", "link1"},
		},
	}

	err := svc.UpdateWithSource(ctx, targetPath, packageDir, []string{"config"}, plan, manifest.SourceAdopted)
	require.NoError(t, err)

	// Load manifest and verify
	manifestResult := svc.Load(ctx, targetPath)
	require.True(t, manifestResult.IsOk())
	m := manifestResult.Unwrap()

	pkg, exists := m.GetPackage("config")
	require.True(t, exists)

	assert.Equal(t, manifest.SourceAdopted, pkg.Source)
	assert.NotNil(t, pkg.Backups)
	assert.Equal(t, backup.String(), pkg.Backups[source.String()])
}

// Test manifest schema for backups field
func TestManifestBackupsSchema(t *testing.T) {
	pkg := manifest.PackageInfo{
		Name:        "vim",
		InstalledAt: time.Now(),
		LinkCount:   1,
		Links:       []string{".vimrc"},
		Backups: map[string]string{
			"/home/user/.vimrc": "/backup/.vimrc.20060102-150405",
		},
		Source: manifest.SourceManaged,
	}

	assert.NotNil(t, pkg.Backups)
	assert.Len(t, pkg.Backups, 1)
	assert.Contains(t, pkg.Backups, "/home/user/.vimrc")
}
