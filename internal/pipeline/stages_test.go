package pipeline

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/jamesainslie/dot/internal/ignore"
	"github.com/jamesainslie/dot/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before start", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		scanStage := ScanStage()
		input := ScanInput{
			PackageDir: domain.NewPackagePath("/packages").Unwrap(),
			TargetDir:  domain.NewTargetPath("/target").Unwrap(),
			Packages:   []string{"vim"},
			IgnoreSet:  ignore.NewIgnoreSet(),
			FS:         adapters.NewOSFilesystem(),
		}

		result := scanStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})

	t.Run("empty package list with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		scanStage := ScanStage()
		input := ScanInput{
			PackageDir: domain.NewPackagePath("/packages").Unwrap(),
			TargetDir:  domain.NewTargetPath("/target").Unwrap(),
			Packages:   []string{}, // Empty list
			IgnoreSet:  ignore.NewIgnoreSet(),
			FS:         adapters.NewOSFilesystem(),
		}

		result := scanStage(ctx, input)

		// Should catch cancellation at early check
		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestPlanStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before planning", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		planStage := PlanStage()
		input := PlanInput{
			Packages:  []domain.Package{},
			TargetDir: domain.NewTargetPath("/target").Unwrap(),
		}

		result := planStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestResolveStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before resolution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		resolveStage := ResolveStage()
		input := ResolveInput{
			Desired: planner.DesiredState{
				Links: make(map[string]planner.LinkSpec),
				Dirs:  make(map[string]planner.DirSpec),
			},
			FS:        adapters.NewOSFilesystem(),
			Policies:  planner.DefaultPolicies(),
			BackupDir: "",
		}

		result := resolveStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestSortStage_ContextCancellation(t *testing.T) {
	t.Run("cancelled before sorting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		sortStage := SortStage()
		input := SortInput{
			Operations: []domain.Operation{},
		}

		result := sortStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})

	t.Run("cancelled with operations", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		sortStage := SortStage()

		source := domain.NewFilePath("/packages/vim/vimrc").Unwrap()
		target := domain.NewTargetPath("/home/user/.vimrc").Unwrap()

		input := SortInput{
			Operations: []domain.Operation{
				domain.NewLinkCreate("link1", source, target),
			},
		}

		result := sortStage(ctx, input)

		require.False(t, result.IsOk())
		assert.Equal(t, context.Canceled, result.UnwrapErr())
	})
}

func TestStages_ValidContextPropagation(t *testing.T) {
	t.Run("all stages respect valid context", func(t *testing.T) {
		ctx := context.Background()

		// Scan stage with empty packages
		scanResult := ScanStage()(ctx, ScanInput{
			PackageDir: domain.NewPackagePath("/packages").Unwrap(),
			TargetDir:  domain.NewTargetPath("/target").Unwrap(),
			Packages:   []string{},
			IgnoreSet:  ignore.NewIgnoreSet(),
			FS:         adapters.NewOSFilesystem(),
		})
		require.True(t, scanResult.IsOk())

		// Plan stage with empty packages
		planResult := PlanStage()(ctx, PlanInput{
			Packages:  []domain.Package{},
			TargetDir: domain.NewTargetPath("/target").Unwrap(),
		})
		require.True(t, planResult.IsOk())

		// Resolve stage with empty state
		resolveResult := ResolveStage()(ctx, ResolveInput{
			Desired: planner.DesiredState{
				Links: make(map[string]planner.LinkSpec),
				Dirs:  make(map[string]planner.DirSpec),
			},
			FS:        adapters.NewOSFilesystem(),
			Policies:  planner.DefaultPolicies(),
			BackupDir: "",
		})
		require.True(t, resolveResult.IsOk())

		// Sort stage with empty operations
		sortResult := SortStage()(ctx, SortInput{
			Operations: []domain.Operation{},
		})
		require.True(t, sortResult.IsOk())
	})
}

func TestScanCurrentState_NonExistentDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/nonexistent").Unwrap()

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Empty(t, result.Files, "should have no files for nonexistent directory")
	assert.Empty(t, result.Links, "should have no links for nonexistent directory")
	assert.Empty(t, result.Dirs, "should have no dirs for nonexistent directory")
}

func TestScanCurrentState_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create empty target directory
	require.NoError(t, fs.MkdirAll(ctx, targetDir.String(), 0o755))

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Empty(t, result.Files, "should have no files in empty directory")
	assert.Empty(t, result.Links, "should have no links in empty directory")
	assert.Empty(t, result.Dirs, "should have no dirs in empty directory")
}

func TestScanCurrentState_FilesOnly(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create target directory with files
	require.NoError(t, fs.MkdirAll(ctx, targetDir.String(), 0o755))
	require.NoError(t, fs.WriteFile(ctx, "/target/.vimrc", []byte("vim config"), 0o644))
	require.NoError(t, fs.WriteFile(ctx, "/target/.bashrc", []byte("bash config"), 0o644))

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Len(t, result.Files, 2, "should detect 2 files")
	assert.Contains(t, result.Files, "/target/.vimrc")
	assert.Contains(t, result.Files, "/target/.bashrc")
	assert.Equal(t, int64(10), result.Files["/target/.vimrc"].Size)
	assert.Equal(t, int64(11), result.Files["/target/.bashrc"].Size)
	assert.Empty(t, result.Links, "should have no links")
	assert.Empty(t, result.Dirs, "should have no dirs")
}

func TestScanCurrentState_SymlinksOnly(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create target directory with symlinks
	require.NoError(t, fs.MkdirAll(ctx, targetDir.String(), 0o755))
	require.NoError(t, fs.Symlink(ctx, "/packages/vim/.vimrc", "/target/.vimrc"))
	require.NoError(t, fs.Symlink(ctx, "/packages/bash/.bashrc", "/target/.bashrc"))

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Len(t, result.Links, 2, "should detect 2 symlinks")
	assert.Contains(t, result.Links, "/target/.vimrc")
	assert.Contains(t, result.Links, "/target/.bashrc")
	assert.Equal(t, "/packages/vim/.vimrc", result.Links["/target/.vimrc"].Target)
	assert.Equal(t, "/packages/bash/.bashrc", result.Links["/target/.bashrc"].Target)
	assert.Empty(t, result.Files, "should have no files")
	assert.Empty(t, result.Dirs, "should have no dirs")
}

func TestScanCurrentState_NestedDirectories(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create nested directory structure
	require.NoError(t, fs.MkdirAll(ctx, "/target/.config/nvim", 0o755))
	require.NoError(t, fs.MkdirAll(ctx, "/target/.local/share", 0o755))
	require.NoError(t, fs.WriteFile(ctx, "/target/.config/nvim/init.vim", []byte("neovim"), 0o644))
	require.NoError(t, fs.WriteFile(ctx, "/target/.local/share/data.txt", []byte("data"), 0o644))

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Len(t, result.Dirs, 4, "should detect 4 directories")
	assert.Contains(t, result.Dirs, "/target/.config")
	assert.Contains(t, result.Dirs, "/target/.config/nvim")
	assert.Contains(t, result.Dirs, "/target/.local")
	assert.Contains(t, result.Dirs, "/target/.local/share")

	assert.Len(t, result.Files, 2, "should detect 2 files")
	assert.Contains(t, result.Files, "/target/.config/nvim/init.vim")
	assert.Contains(t, result.Files, "/target/.local/share/data.txt")

	assert.Empty(t, result.Links, "should have no links")
}

func TestScanCurrentState_MixedContent(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create mixed content: files, symlinks, and directories
	require.NoError(t, fs.MkdirAll(ctx, "/target/.config", 0o755))
	require.NoError(t, fs.WriteFile(ctx, "/target/.bashrc", []byte("bash"), 0o644))
	require.NoError(t, fs.Symlink(ctx, "/packages/vim/.vimrc", "/target/.vimrc"))
	require.NoError(t, fs.WriteFile(ctx, "/target/.config/settings.conf", []byte("settings"), 0o644))
	require.NoError(t, fs.Symlink(ctx, "/packages/nvim/init.vim", "/target/.config/init.vim"))

	result := scanCurrentState(ctx, fs, targetDir)

	// Check directories
	assert.Len(t, result.Dirs, 1, "should detect 1 directory")
	assert.Contains(t, result.Dirs, "/target/.config")

	// Check files
	assert.Len(t, result.Files, 2, "should detect 2 files")
	assert.Contains(t, result.Files, "/target/.bashrc")
	assert.Contains(t, result.Files, "/target/.config/settings.conf")
	assert.Equal(t, int64(4), result.Files["/target/.bashrc"].Size)
	assert.Equal(t, int64(8), result.Files["/target/.config/settings.conf"].Size)

	// Check symlinks
	assert.Len(t, result.Links, 2, "should detect 2 symlinks")
	assert.Contains(t, result.Links, "/target/.vimrc")
	assert.Contains(t, result.Links, "/target/.config/init.vim")
	assert.Equal(t, "/packages/vim/.vimrc", result.Links["/target/.vimrc"].Target)
	assert.Equal(t, "/packages/nvim/init.vim", result.Links["/target/.config/init.vim"].Target)
}

func TestScanCurrentState_DeepNesting(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	targetDir := domain.NewTargetPath("/target").Unwrap()

	// Create deeply nested structure
	deepPath := "/target/a/b/c/d/e"
	require.NoError(t, fs.MkdirAll(ctx, deepPath, 0o755))
	require.NoError(t, fs.WriteFile(ctx, deepPath+"/deep.txt", []byte("deep file"), 0o644))

	result := scanCurrentState(ctx, fs, targetDir)

	assert.Len(t, result.Dirs, 5, "should detect 5 nested directories")
	assert.Contains(t, result.Dirs, "/target/a")
	assert.Contains(t, result.Dirs, "/target/a/b")
	assert.Contains(t, result.Dirs, "/target/a/b/c")
	assert.Contains(t, result.Dirs, "/target/a/b/c/d")
	assert.Contains(t, result.Dirs, "/target/a/b/c/d/e")

	assert.Len(t, result.Files, 1, "should detect 1 file")
	assert.Contains(t, result.Files, "/target/a/b/c/d/e/deep.txt")
}
