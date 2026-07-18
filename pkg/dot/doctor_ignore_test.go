package dot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/manifest"
)

func newIgnoreTestService(t *testing.T) (*DoctorService, *adapters.MemFS) {
	t.Helper()
	fs := adapters.NewMemFS()
	ctx := context.Background()
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0o755))
	require.NoError(t, fs.MkdirAll(ctx, "/packages", 0o755))
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, adapters.NewNoopLogger(), store)
	svc := newDoctorService(fs, adapters.NewNoopLogger(), manifestSvc, "/packages", "/home")
	return svc, fs
}

func TestDoctorService_IgnoreAndUnignorePattern(t *testing.T) {
	svc, _ := newIgnoreTestService(t)
	ctx := context.Background()

	require.NoError(t, svc.IgnorePattern(ctx, "Code/*"))

	_, patterns, err := svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"Code/*"}, patterns)

	// Adding the same pattern again must not duplicate it.
	require.NoError(t, svc.IgnorePattern(ctx, "Code/*"))
	_, patterns, err = svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"Code/*"}, patterns)

	require.NoError(t, svc.UnignorePattern(ctx, "Code/*"))
	_, patterns, err = svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestDoctorService_UnignorePattern_NotFound(t *testing.T) {
	svc, _ := newIgnoreTestService(t)
	ctx := context.Background()

	err := svc.UnignorePattern(ctx, "nope/*")
	assert.Error(t, err)
}

func TestDoctorService_IgnoreLinkRoundTrip(t *testing.T) {
	svc, fs := newIgnoreTestService(t)
	ctx := context.Background()

	require.NoError(t, fs.Symlink(ctx, "/nix/store/abc/profile", "/home/.nix-profile"))

	require.NoError(t, svc.IgnoreLink(ctx, ".nix-profile", "nix managed"))

	links, _, err := svc.ListIgnored(ctx)
	require.NoError(t, err)
	require.Contains(t, links, ".nix-profile")
	assert.Equal(t, "/nix/store/abc/profile", links[".nix-profile"].Target)
	assert.Equal(t, "nix managed", links[".nix-profile"].Reason)

	require.NoError(t, svc.UnignoreLink(ctx, ".nix-profile"))
	links, _, err = svc.ListIgnored(ctx)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestClient_DoctorIgnoreSurface(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()
	require.NoError(t, fs.MkdirAll(ctx, "/home", 0o755))
	require.NoError(t, fs.MkdirAll(ctx, "/packages", 0o755))

	client, err := NewClient(Config{
		PackageDir: "/packages",
		TargetDir:  "/home",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	})
	require.NoError(t, err)

	require.NoError(t, client.DoctorIgnorePattern(ctx, "Bootstrap/*"))
	links, patterns, err := client.DoctorListIgnored(ctx)
	require.NoError(t, err)
	assert.Empty(t, links)
	assert.Equal(t, []string{"Bootstrap/*"}, patterns)

	require.NoError(t, client.DoctorUnignorePattern(ctx, "Bootstrap/*"))
	_, patterns, err = client.DoctorListIgnored(ctx)
	require.NoError(t, err)
	assert.Empty(t, patterns)
}
