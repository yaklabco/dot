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

	t.Run("handles non-existent package", func(t *testing.T) {
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
		err := svc.Unmanage(ctx, "non-existent")
		require.NoError(t, err) // Should not error, just no-op
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

	t.Run("returns empty plan when no manifest exists", func(t *testing.T) {
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
		plan, err := svc.PlanUnmanage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Len(t, plan.Operations, 0)
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
