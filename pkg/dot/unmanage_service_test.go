package dot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
	"github.com/yaklabco/dot/internal/planner"
)

func TestUnmanageService_Unmanage(t *testing.T) {
	t.Run("unmanages package successfully", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup and manage package first
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Manage first
		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link created
		assert.True(t, fs.Exists(ctx, targetDir+"/.vimrc"))

		// Now unmanage
		err = unmanageSvc.Unmanage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link removed
		assert.False(t, fs.Exists(ctx, targetDir+"/.vimrc"))
	})

	t.Run("errors on non-existent package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		// Create a managed package so the manifest exists
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/real-pkg", 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/real-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage a real package so the manifest exists
		err := manageSvc.Manage(ctx, "real-pkg")
		require.NoError(t, err)

		// Unmanaging a non-existent package should return an error
		err = unmanageSvc.Unmanage(ctx, "non-existent")
		require.Error(t, err)

		var notFound ErrPackageNotFound
		require.ErrorAs(t, err, &notFound)
		assert.Equal(t, "non-existent", notFound.Package)
	})
}

func TestUnmanageService_Unmanage_CleansEmptyDirectories(t *testing.T) {
	t.Run("removes empty parent directories after unmanage", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package with nested directory structure
		// dot-config/nvim/init.lua translates to target: dot-config/nvim/init.lua
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg/dot-config/nvim", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-config/nvim/init.lua", []byte("lua"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage the package
		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify nested structure created (dot-config translates to .config in target)
		assert.True(t, fs.Exists(ctx, targetDir+"/.config/nvim/init.lua"))
		assert.True(t, fs.Exists(ctx, targetDir+"/.config/nvim"))
		assert.True(t, fs.Exists(ctx, targetDir+"/.config"))

		// Unmanage
		err = unmanageSvc.Unmanage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link removed
		assert.False(t, fs.Exists(ctx, targetDir+"/.config/nvim/init.lua"))

		// Verify empty parent directories are cleaned up
		assert.False(t, fs.Exists(ctx, targetDir+"/.config/nvim"), "empty nvim dir should be removed")
		assert.False(t, fs.Exists(ctx, targetDir+"/.config"), "empty .config dir should be removed")

		// Target dir itself should still exist
		assert.True(t, fs.Exists(ctx, targetDir))
	})

	t.Run("preserves non-empty parent directories", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package with nested dir
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg/dot-config/nvim", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-config/nvim/init.lua", []byte("lua"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Create another file in .config AFTER manage so it's not empty after unmanage
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.config/other-file.txt", []byte("keep me"), 0644))

		// Unmanage
		err = unmanageSvc.Unmanage(ctx, "test-pkg")
		require.NoError(t, err)

		// nvim subdir should be cleaned up (empty after link removal)
		assert.False(t, fs.Exists(ctx, targetDir+"/.config/nvim"), "empty nvim dir should be removed")

		// .config should still exist (has other-file.txt)
		assert.True(t, fs.Exists(ctx, targetDir+"/.config"), ".config should remain (non-empty)")
		assert.True(t, fs.Exists(ctx, targetDir+"/.config/other-file.txt"))
	})
}

func TestUnmanageService_PlanUnmanage(t *testing.T) {
	t.Run("creates delete operations for installed package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup and manage package first
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Plan unmanage
		plan, err := unmanageSvc.PlanUnmanage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Greater(t, len(plan.Operations), 0)
	})

	t.Run("handles multiple packages", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup two packages
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg1", 0755))
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg2", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg1/dot-file1", []byte("content1"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg2/dot-file2", []byte("content2"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage both
		require.NoError(t, manageSvc.Manage(ctx, "pkg1", "pkg2"))

		// Plan unmanage both
		plan, err := unmanageSvc.PlanUnmanage(ctx, "pkg1", "pkg2")
		require.NoError(t, err)
		assert.Greater(t, len(plan.Operations), 1)
	})

	t.Run("errors when no manifest exists", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)

		svc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		_, err := svc.PlanUnmanage(ctx, "test-pkg")
		require.Error(t, err)
		var notFound ErrPackageNotFound
		require.ErrorAs(t, err, &notFound)
	})
}

func TestUnmanageService_UnmanageAll(t *testing.T) {
	t.Run("unmanages all packages successfully", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup two packages
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg1", 0755))
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/pkg2", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg1/dot-file1", []byte("content1"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/pkg2/dot-file2", []byte("content2"), 0644))

		// Setup services
		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage both packages
		require.NoError(t, manageSvc.Manage(ctx, "pkg1", "pkg2"))

		// Verify links created
		assert.True(t, fs.Exists(ctx, targetDir+"/.file1"))
		assert.True(t, fs.Exists(ctx, targetDir+"/.file2"))

		// Unmanage all
		count, err := unmanageSvc.UnmanageAll(ctx, DefaultUnmanageOptions())
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Verify all links removed
		assert.False(t, fs.Exists(ctx, targetDir+"/.file1"))
		assert.False(t, fs.Exists(ctx, targetDir+"/.file2"))
	})

	t.Run("returns zero when no packages installed", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"
		require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)

		count, err := unmanageSvc.UnmanageAll(ctx, DefaultUnmanageOptions())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("dry run mode returns count without removing links", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-file", []byte("content"), 0644))

		// Setup services with dry-run enabled
		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, true) // dry-run=true
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage package
		require.NoError(t, manageSvc.Manage(ctx, "test-pkg"))
		assert.True(t, fs.Exists(ctx, targetDir+"/.file"))

		// Dry run unmanage all
		count, err := unmanageSvc.UnmanageAll(ctx, DefaultUnmanageOptions())
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify link still exists
		assert.True(t, fs.Exists(ctx, targetDir+"/.file"))
	})

	t.Run("handles purge and no-restore options", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-file", []byte("content"), 0644))

		// Setup services
		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Manage package first
		require.NoError(t, manageSvc.Manage(ctx, "test-pkg"))

		// Unmanage all with custom options
		opts := UnmanageOptions{
			Purge:   true,
			Restore: false,
			Cleanup: false,
		}
		count, err := unmanageSvc.UnmanageAll(ctx, opts)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("error on invalid target directory", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "" // Invalid empty target dir

		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: adapters.NewNoopLogger(),
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)

		// Should error on invalid target
		count, err := unmanageSvc.UnmanageAll(ctx, DefaultUnmanageOptions())
		require.Error(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestUnmanageService_MultiPackageManifestBatch(t *testing.T) {
	t.Run("unmanaging multiple packages removes all from manifest", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup: create 3 packages
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		for _, pkg := range []string{"bash", "vim", "git"} {
			require.NoError(t, fs.MkdirAll(ctx, packageDir+"/"+pkg, 0755))
		}
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/bash/dot-bashrc", []byte("bash"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/vim/dot-vimrc", []byte("vim"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/git/dot-gitconfig", []byte("git"), 0644))

		logger := adapters.NewNoopLogger()
		exec := executor.New(executor.Opts{
			FS:     fs,
			Logger: logger,
			Tracer: adapters.NewNoopTracer(),
		})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, logger, manifestStore)

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		unmanageSvc := newUnmanageService(fs, logger, exec, manifestSvc, packageDir, targetDir, false)
		manageSvc := newManageService(fs, logger, managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := manageSvc.Manage(ctx, "bash", "vim", "git")
		require.NoError(t, err)

		// Verify all 3 are in manifest
		targetPath := NewTargetPath(targetDir)
		require.True(t, targetPath.IsOk())
		mResult := manifestSvc.Load(ctx, targetPath.Unwrap())
		require.True(t, mResult.IsOk())
		m := mResult.Unwrap()
		assert.Len(t, m.Packages, 3)

		// Unmanage all 3 at once
		err = unmanageSvc.UnmanageWithOptions(ctx, DefaultUnmanageOptions(), "bash", "vim", "git")
		require.NoError(t, err)

		// Verify manifest has no packages left
		mResult = manifestSvc.Load(ctx, targetPath.Unwrap())
		require.True(t, mResult.IsOk())
		m = mResult.Unwrap()
		assert.Empty(t, m.Packages, "all packages should be removed from manifest after multi-package unmanage")
	})
}
