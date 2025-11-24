package dot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestClient_Doctor_BrokenLinks(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/broken", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/broken/dot-config", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage
	err = client.Manage(ctx, "broken")
	require.NoError(t, err)

	// Break the link by removing source
	require.NoError(t, fs.Remove(ctx, "/test/packages/broken/dot-config"))

	// Doctor should detect broken link
	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
	assert.True(t, report.Statistics.BrokenLinks > 0)

	// Should have broken link issue
	hasBroken := false
	for _, issue := range report.Issues {
		if issue.Type == dot.IssueBrokenLink {
			hasBroken = true
			break
		}
	}
	assert.True(t, hasBroken)
}

func TestClient_Doctor_WrongLinkType(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Replace symlink with regular file
	linkPath := "/test/target/.file"
	require.NoError(t, fs.Remove(ctx, linkPath))
	require.NoError(t, fs.WriteFile(ctx, linkPath, []byte("not a link"), 0644))

	// Doctor should detect wrong type
	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	assert.Equal(t, dot.HealthErrors, report.OverallHealth)
}
