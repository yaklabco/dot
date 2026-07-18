package dot

import (
	"context"
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

// registrationTestEnv bundles the services needed to exercise manage and
// remanage against a MemFS with a two-file package.
type registrationTestEnv struct {
	fs          *adapters.MemFS
	svc         *ManageService
	manifestSvc *ManifestService
	managePipe  *pipeline.ManagePipeline
	exec        *executor.Executor
	unmanageSvc *UnmanageService
	packageDir  string
	targetDir   string
}

func newRegistrationTestEnv(t *testing.T) *registrationTestEnv {
	t.Helper()
	fs := adapters.NewMemFS()
	ctx := context.Background()
	packageDir := "/test/packages"
	targetDir := "/test/target"

	require.NoError(t, fs.MkdirAll(ctx, packageDir+"/test-pkg", 0o755))
	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0o755))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-vimrc", []byte("vim"), 0o644))
	require.NoError(t, fs.WriteFile(ctx, packageDir+"/test-pkg/dot-bashrc", []byte("bash"), 0o644))

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

	return &registrationTestEnv{
		fs:          fs,
		svc:         svc,
		manifestSvc: manifestSvc,
		managePipe:  managePipe,
		exec:        exec,
		unmanageSvc: unmanageSvc,
		packageDir:  packageDir,
		targetDir:   targetDir,
	}
}

// manifestLinks loads the manifest and returns the recorded links for test-pkg.
func (e *registrationTestEnv) manifestLinks(t *testing.T) []string {
	t.Helper()
	ctx := context.Background()
	targetPath := NewTargetPath(e.targetDir).Unwrap()
	result := e.manifestSvc.Load(ctx, targetPath)
	require.True(t, result.IsOk(), "manifest must load")
	m := result.Unwrap()
	pkg, exists := m.GetPackage("test-pkg")
	require.True(t, exists, "test-pkg must be in manifest")
	return pkg.Links
}

// seedManifest writes a manifest entry for test-pkg with the given links and
// a current content hash, simulating a stale manifest that predates some links.
func (e *registrationTestEnv) seedManifest(t *testing.T, links []string) {
	t.Helper()
	ctx := context.Background()
	targetPath := NewTargetPath(e.targetDir).Unwrap()
	m := manifest.New()
	m.AddPackage(manifest.PackageInfo{
		Name:        "test-pkg",
		InstalledAt: time.Now(),
		LinkCount:   len(links),
		Links:       links,
		Source:      manifest.SourceManaged,
		TargetDir:   e.targetDir,
		PackageDir:  e.packageDir + "/test-pkg",
	})
	hasher := manifest.NewContentHasher(e.fs)
	pkgPath := NewPackagePath(e.packageDir + "/test-pkg").Unwrap()
	hash, err := hasher.HashPackage(ctx, pkgPath)
	require.NoError(t, err)
	m.SetHash("test-pkg", hash)
	require.NoError(t, e.manifestSvc.Save(ctx, targetPath, m))
}

func TestManage_RecordsPreExistingCorrectLinks(t *testing.T) {
	env := newRegistrationTestEnv(t)
	ctx := context.Background()

	// One link already exists and points at the correct package file.
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-bashrc", env.targetDir+"/.bashrc"))

	require.NoError(t, env.svc.Manage(ctx, "test-pkg"))

	assert.ElementsMatch(t, []string{".vimrc", ".bashrc"}, env.manifestLinks(t),
		"manifest must record the pre-existing correct link, not only the newly created one")
}

func TestManage_HealsManifestMissingLinksWhenAllLinksExist(t *testing.T) {
	env := newRegistrationTestEnv(t)
	ctx := context.Background()

	// Both links exist correctly on disk.
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-vimrc", env.targetDir+"/.vimrc"))
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-bashrc", env.targetDir+"/.bashrc"))

	// Manifest predates .bashrc: only .vimrc recorded.
	env.seedManifest(t, []string{".vimrc"})

	err := env.svc.Manage(ctx, "test-pkg")
	require.NoError(t, err, "healing the manifest is a change, not a no-op")

	assert.ElementsMatch(t, []string{".vimrc", ".bashrc"}, env.manifestLinks(t),
		"manage must adopt already-correct links into the manifest")
}

func TestRemanage_RecreatesLinkMissingFromDiskAndManifest(t *testing.T) {
	env := newRegistrationTestEnv(t)
	ctx := context.Background()

	// Only .vimrc is linked and recorded; .bashrc is neither on disk nor in the
	// manifest, and the stored package hash matches current content, so the
	// incremental fast path would previously report "no changes".
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-vimrc", env.targetDir+"/.vimrc"))
	env.seedManifest(t, []string{".vimrc"})

	require.NoError(t, env.svc.Remanage(ctx, "test-pkg"))

	isLink, err := env.fs.IsSymlink(ctx, env.targetDir+"/.bashrc")
	require.NoError(t, err)
	assert.True(t, isLink, "remanage must recreate the missing link")
	assert.ElementsMatch(t, []string{".vimrc", ".bashrc"}, env.manifestLinks(t))
}

func TestRemanage_HealsManifestMissingLinksWhenAllLinksExist(t *testing.T) {
	env := newRegistrationTestEnv(t)
	ctx := context.Background()

	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-vimrc", env.targetDir+"/.vimrc"))
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-bashrc", env.targetDir+"/.bashrc"))
	env.seedManifest(t, []string{".vimrc"})

	err := env.svc.Remanage(ctx, "test-pkg")
	require.NoError(t, err, "healing the manifest is a change, not a no-op")

	assert.ElementsMatch(t, []string{".vimrc", ".bashrc"}, env.manifestLinks(t),
		"remanage must adopt already-correct links into the manifest")
}

func TestManage_DryRunDoesNotReconcileManifest(t *testing.T) {
	env := newRegistrationTestEnv(t)
	ctx := context.Background()

	// All links exist correctly on disk but the package is not in the manifest.
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-vimrc", env.targetDir+"/.vimrc"))
	require.NoError(t, env.fs.Symlink(ctx, env.packageDir+"/test-pkg/dot-bashrc", env.targetDir+"/.bashrc"))

	dryService := newManageService(env.fs, adapters.NewNoopLogger(), env.managePipe, env.exec, env.manifestSvc, env.unmanageSvc, env.packageDir, env.targetDir, true)

	_ = dryService.Manage(ctx, "test-pkg")

	targetPath := NewTargetPath(env.targetDir).Unwrap()
	result := env.manifestSvc.Load(ctx, targetPath)
	require.True(t, result.IsOk())
	m := result.Unwrap()
	_, exists := m.GetPackage("test-pkg")
	assert.False(t, exists, "dry-run must not write manifest reconciliation")
}
