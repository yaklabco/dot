package dot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/executor"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/manifest"
	"github.com/yaklabco/dot/internal/pipeline"
	"github.com/yaklabco/dot/internal/planner"
)

func TestManageService_Manage(t *testing.T) {
	t.Run("manages package successfully", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Create dependencies
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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link created
		linkExists := fs.Exists(ctx, targetDir+"/.vimrc")
		assert.True(t, linkExists)
	})

	t.Run("dry run does not execute", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

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
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, true)

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, true)

		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify link NOT created (dry run)
		linkExists := fs.Exists(ctx, targetDir+"/.vimrc")
		assert.False(t, linkExists)
	})
}

func TestManageService_PlanManage(t *testing.T) {
	t.Run("creates execution plan", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{FS: fs, Logger: adapters.NewNoopLogger()})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		plan, err := svc.PlanManage(ctx, "test-pkg")
		require.NoError(t, err)
		assert.Greater(t, len(plan.Operations), 0)
	})
}

func TestManageService_Remanage(t *testing.T) {
	t.Run("returns ErrNoChanges for unchanged packages", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Initial manage
		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Remanage without changes should return ErrNoChanges
		err = svc.Remanage(ctx, "test-pkg")
		require.Error(t, err)
		var noChanges ErrNoChanges
		assert.ErrorAs(t, err, &noChanges)
	})

	t.Run("detects symlink replaced by regular file", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Initial manage creates symlink
		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Verify symlink was created
		isLink, err := fs.IsSymlink(ctx, targetDir+"/.vimrc")
		require.NoError(t, err)
		assert.True(t, isLink, "expected .vimrc to be a symlink after manage")

		// Replace the symlink with a regular file (simulates external modification)
		require.NoError(t, fs.Remove(ctx, targetDir+"/.vimrc"))
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("replaced content"), 0644))

		// Verify it is now a regular file, not a symlink
		isLink, err = fs.IsSymlink(ctx, targetDir+"/.vimrc")
		require.NoError(t, err)
		assert.False(t, isLink, "expected .vimrc to be a regular file after replacement")

		// Remanage should detect the replaced symlink and recreate it
		err = svc.Remanage(ctx, "test-pkg")
		require.NoError(t, err, "remanage should succeed when symlink was replaced by regular file")

		// Verify the symlink was recreated
		isLink, err = fs.IsSymlink(ctx, targetDir+"/.vimrc")
		require.NoError(t, err)
		assert.True(t, isLink, "expected .vimrc to be a symlink after remanage")
	})

	t.Run("remanages adopted package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup adopted package structure
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/dot-ssh", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/dot-ssh/config", []byte("ssh config"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/dot-ssh/known_hosts", []byte("hosts"), 0644))

		// Create manifest with adopted package
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)

		targetPathResult := NewTargetPath(targetDir)
		require.True(t, targetPathResult.IsOk())

		m := manifest.New()
		pkgInfo := manifest.PackageInfo{
			Name:        "dot-ssh",
			InstalledAt: time.Now(),
			LinkCount:   1,
			Links:       []string{".ssh"},
			Source:      manifest.SourceAdopted,
		}
		m.AddPackage(pkgInfo)
		err := manifestSvc.Save(ctx, targetPathResult.Unwrap(), m)
		require.NoError(t, err)

		// Create symlink to simulate adopted state
		require.NoError(t, fs.Symlink(ctx, packageDir+"/dot-ssh", targetDir+"/.ssh"))

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
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)
		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// Remanage adopted package - recreates the symlink (produces operations)
		err = svc.Remanage(ctx, "dot-ssh")
		require.NoError(t, err)

		// Verify symlink still exists
		linkExists := fs.Exists(ctx, targetDir+"/.ssh")
		assert.True(t, linkExists)
	})
}

func TestManageService_Manage_CorruptManifest(t *testing.T) {
	t.Run("returns error when manifest is corrupt", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Write corrupt manifest directly in target dir (default manifest location)
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.dot-manifest.json", []byte("{invalid json"), 0644))

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := svc.Manage(ctx, "test-pkg")
		require.Error(t, err, "manage should return error when manifest is corrupt")
		assert.Contains(t, err.Error(), "manifest")
	})
}

func TestManageService_Manage_CorruptManifest_AlreadyManaged(t *testing.T) {
	t.Run("returns error when re-managing with corrupt manifest", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicySkip},
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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// First manage succeeds (creates symlink + manifest)
		err := svc.Manage(ctx, "test-pkg")
		require.NoError(t, err)

		// Corrupt the manifest
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.dot-manifest.json", []byte("{invalid json"), 0644))

		// Re-manage same package with corrupt manifest should return error,
		// not silently report "no changes detected"
		err = svc.Manage(ctx, "test-pkg")
		require.Error(t, err, "manage should return error when manifest is corrupt, even for already-managed packages")
		assert.Contains(t, err.Error(), "manifest")
	})
}

func TestManageService_Remanage_CorruptManifest(t *testing.T) {
	t.Run("returns error when manifest is corrupt", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0644))

		// Write corrupt manifest before remanage
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.dot-manifest.json", []byte("{invalid json"), 0644))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicySkip},
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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		// PlanRemanage currently falls back to PlanManage on corrupt manifest
		// instead of returning an error. This should be fixed.
		err := svc.Remanage(ctx, "test-pkg")
		require.Error(t, err, "remanage should return error when manifest is corrupt")
		assert.Contains(t, err.Error(), "manifest")
	})
}

func TestManageService_PlanManage_ReservedName(t *testing.T) {
	t.Run("returns specific error for single reserved package", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{FS: fs, Logger: adapters.NewNoopLogger()})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		_, err := svc.PlanManage(ctx, "dot")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reserved")
		assert.Contains(t, err.Error(), "dot")
	})

	t.Run("returns specific error for all reserved packages", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		require.NoError(t, fs.MkdirAll(ctx, packageDir, 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

		managePipe := pipeline.NewManagePipeline(pipeline.ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignore.NewDefaultIgnoreSet(),
			Policies:           planner.ResolutionPolicies{OnFileExists: planner.PolicyFail},
			PackageNameMapping: false,
		})
		exec := executor.New(executor.Opts{FS: fs, Logger: adapters.NewNoopLogger()})
		manifestStore := manifest.NewFSManifestStore(fs)
		manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), manifestStore)
		unmanageSvc := newUnmanageService(fs, adapters.NewNoopLogger(), exec, manifestSvc, packageDir, targetDir, false)

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		_, err := svc.PlanManage(ctx, "dot", ".dot", "dot-config")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reserved")
	})
}

func TestManageService_ConflictReturnsTypedError(t *testing.T) {
	t.Run("returns typed ErrConflict when conflicts detected", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package with a file
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim config"), 0644))

		// Create existing file at target to cause conflict
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("existing file"), 0644))

		// Use PolicyFail to ensure conflicts are detected
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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := svc.Manage(ctx, "test-pkg")

		// Verify error occurred
		require.Error(t, err)

		// Verify error is typed as ErrConflict using errors.As
		var conflictErr ErrConflict
		require.True(t, errors.As(err, &conflictErr), "expected error to be ErrConflict, got %T: %v", err, err)

		// Verify the error message contains conflict details
		assert.Contains(t, err.Error(), "conflict")
		assert.Contains(t, err.Error(), ".vimrc")
	})

	t.Run("ErrConflict contains first conflict path", func(t *testing.T) {
		fs := adapters.NewMemFS()
		ctx := context.Background()
		packageDir := "/test/packages"
		targetDir := "/test/target"

		// Setup package with multiple files
		require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0755))
		require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-bashrc", []byte("bash config"), 0644))
		require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim config"), 0644))

		// Create existing files at target to cause multiple conflicts
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.bashrc", []byte("existing bashrc"), 0644))
		require.NoError(t, fs.WriteFile(ctx, targetDir+"/.vimrc", []byte("existing vimrc"), 0644))

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

		svc := newManageService(fs, adapters.NewNoopLogger(), managePipe, exec, manifestSvc, unmanageSvc, packageDir, targetDir, false)

		err := svc.Manage(ctx, "test-pkg")

		require.Error(t, err)

		var conflictErr ErrConflict
		require.True(t, errors.As(err, &conflictErr), "expected error to be ErrConflict, got %T", err)

		// The Path field should be set to the first conflict's path
		assert.NotEmpty(t, conflictErr.Path, "ErrConflict.Path should not be empty")

		// The Reason field should contain the full error message with all conflicts
		assert.Contains(t, conflictErr.Reason, "conflict")
	})
}
