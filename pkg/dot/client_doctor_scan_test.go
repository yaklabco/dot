package dot_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Doctor_OrphanedLinkDetection(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup managed package
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Create orphaned symlink
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/.orphaned"))

	// Test scoped scan (should detect orphan)
	report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
	require.NoError(t, err)
	assert.True(t, report.Statistics.OrphanedLinks >= 1, "Expected to detect orphaned link")

	// Verify issues reported - broken orphaned link is reported as broken link (error)
	hasOrphanOrBrokenIssue := false
	for _, issue := range report.Issues {
		if issue.Type == dot.IssueOrphanedLink || issue.Type == dot.IssueBrokenLink {
			hasOrphanOrBrokenIssue = true
			break
		}
	}
	assert.True(t, hasOrphanOrBrokenIssue, "Expected orphaned or broken link issue")
}

func TestClient_Doctor_NestedDirectories(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup with nested structure
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/deep", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/subdir/nested", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/deep/dot-file", []byte("x"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Test deep scan with depth limit
	scanCfg := dot.DeepScanConfig(10)
	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

func TestClient_Doctor_SkipPatterns(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/.git", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/node_modules", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("x"), 0644))

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

	// Create links in directories that should be skipped
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/.git/link"))
	require.NoError(t, fs.Symlink(ctx, "/nowhere", "/test/target/node_modules/link"))

	// Deep scan should skip these directories
	report, err := client.DoctorWithScan(ctx, dot.DeepScanConfig(5))
	require.NoError(t, err)

	// Links in skipped directories should not be reported as orphans
	assert.NotNil(t, report)
}

func TestClient_Doctor_OrphanedLinkBrokenTarget(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Create orphaned symlink with broken target
	require.NoError(t, fs.Symlink(ctx, "/nonexistent", "/test/target/.orphaned-broken"))

	// Create orphaned symlink with valid target
	require.NoError(t, fs.MkdirAll(ctx, "/test/valid-target", 0755))
	require.NoError(t, fs.Symlink(ctx, "/test/valid-target", "/test/target/.orphaned-ok"))

	// Test scoped scan
	report, err := client.DoctorWithScan(ctx, dot.ScopedScanConfig())
	require.NoError(t, err)

	// Should detect 2 orphaned links
	assert.Equal(t, 2, report.Statistics.OrphanedLinks, "Expected 2 orphaned links")

	// Should detect 1 broken link
	assert.Equal(t, 1, report.Statistics.BrokenLinks, "Expected 1 broken link")

	// Verify issue types
	// Both orphaned links should be classified as IssueOrphanedLink
	// (even the one with a broken target, since they're both unmanaged)
	orphanedCount := 0
	orphanedWithBrokenTarget := 0
	orphanedWithValidTarget := 0
	
	for _, issue := range report.Issues {
		if issue.Type == dot.IssueOrphanedLink {
			orphanedCount++
			if issue.Severity == dot.SeverityError {
				orphanedWithBrokenTarget++
				assert.Contains(t, issue.Path, "orphaned-broken", "Error orphaned issue should reference broken symlink")
			} else {
				orphanedWithValidTarget++
				assert.Equal(t, dot.SeverityWarning, issue.Severity, "Orphaned link with valid target should be warning")
				assert.Contains(t, issue.Path, "orphaned-ok", "Warning orphaned issue should reference valid symlink")
			}
		}
	}

	assert.Equal(t, 2, orphanedCount, "Expected 2 orphaned link issues")
	assert.Equal(t, 1, orphanedWithBrokenTarget, "Expected 1 orphaned link with broken target (error)")
	assert.Equal(t, 1, orphanedWithValidTarget, "Expected 1 orphaned link with valid target (warning)")
}

func TestClient_Doctor_DefaultScanMode(t *testing.T) {
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

	// Manage package
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Create orphaned symlink
	require.NoError(t, fs.WriteFile(ctx, "/test/target-file", []byte("x"), 0644))
	require.NoError(t, fs.Symlink(ctx, "/test/target-file", "/test/target/.orphan"))

	// Test default doctor (should use scoped scanning)
	report, err := client.Doctor(ctx)
	require.NoError(t, err)

	// Should detect orphaned link with default scoped scanning
	assert.Equal(t, 1, report.Statistics.OrphanedLinks, "Expected default scoped scan to detect orphan")
}
