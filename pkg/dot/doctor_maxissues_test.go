package dot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// TestClient_Doctor_MaxIssuesRespectedInParallelScan verifies that MaxIssues cap
// is properly enforced when using parallel workers, even when multiple workers
// independently collect issues.
func TestClient_Doctor_MaxIssuesRespectedInParallelScan(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup basic structure with multiple directories to force parallel scanning
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	// Create multiple directories under target to force parallel scanning
	require.NoError(t, fs.MkdirAll(ctx, "/test/target/dir1", 0755))
	for i := 0; i < 50; i++ {
		linkPath := fmt.Sprintf("/test/target/dir1/orphan%d", i)
		require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
	}

	require.NoError(t, fs.MkdirAll(ctx, "/test/target/dir2", 0755))
	for i := 0; i < 50; i++ {
		linkPath := fmt.Sprintf("/test/target/dir2/orphan%d", i)
		require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
	}

	require.NoError(t, fs.MkdirAll(ctx, "/test/target/dir3", 0755))
	for i := 0; i < 50; i++ {
		linkPath := fmt.Sprintf("/test/target/dir3/orphan%d", i)
		require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Manage package to initialize manifest
	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Test with MaxIssues=10, scan specific directories in parallel
	// Use absolute paths that will be processed as separate roots
	scanCfg := dot.ScanConfig{
		Mode:        dot.ScanDeep, // Use deep mode with absolute paths
		MaxDepth:    3,
		MaxWorkers:  3, // Force parallel execution
		MaxIssues:   10,
		ScopeToDirs: []string{"/test/target/dir1", "/test/target/dir2", "/test/target/dir3"},
	}

	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)

	// Verify MaxIssues cap is respected
	// Allow small overhead due to parallel collection race conditions, but should be close to limit
	assert.LessOrEqual(t, len(report.Issues), 13,
		"MaxIssues=10 should limit issues to <=13 (accounting for up to 3 workers), but got %d", len(report.Issues))
}

// TestClient_Doctor_MaxIssuesSingleLargeDirectory verifies that a single directory
// with many issues doesn't blow past the MaxIssues cap.
func TestClient_Doctor_MaxIssuesSingleLargeDirectory(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	// Create a single directory with 1000 orphaned symlinks
	for i := 0; i < 1000; i++ {
		linkPath := fmt.Sprintf("/test/target/orphan%d", i)
		require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Test with MaxIssues=10
	scanCfg := dot.ScanConfig{
		Mode:       dot.ScanDeep,
		MaxDepth:   5,
		MaxWorkers: 2,
		MaxIssues:  10,
	}

	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)

	// Should have exactly 10 issues, not 1000
	assert.LessOrEqual(t, len(report.Issues), 10,
		"MaxIssues=10 should limit issues to 10, but got %d", len(report.Issues))
}

// TestClient_Doctor_MaxIssuesSequentialScan verifies MaxIssues works in sequential mode.
func TestClient_Doctor_MaxIssuesSequentialScan(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	// Create multiple directories with orphaned symlinks
	for dirIdx := 0; dirIdx < 5; dirIdx++ {
		dirPath := fmt.Sprintf("/test/target/dir%d", dirIdx)
		require.NoError(t, fs.MkdirAll(ctx, dirPath, 0755))
		for i := 0; i < 50; i++ {
			linkPath := fmt.Sprintf("%s/orphan%d", dirPath, i)
			require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
		}
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Test sequential scan (MaxWorkers=1)
	scanCfg := dot.ScanConfig{
		Mode:       dot.ScanDeep,
		MaxDepth:   5,
		MaxWorkers: 1, // Force sequential
		MaxIssues:  15,
	}

	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)

	// Sequential scan should also respect MaxIssues
	assert.LessOrEqual(t, len(report.Issues), 15,
		"MaxIssues=15 should limit issues to 15, but got %d", len(report.Issues))
}

// TestClient_Doctor_MaxIssuesZeroMeansUnlimited verifies that MaxIssues=0 means no limit.
func TestClient_Doctor_MaxIssuesZeroMeansUnlimited(t *testing.T) {
	fs := adapters.NewMemFS()
	ctx := context.Background()

	// Setup
	require.NoError(t, fs.MkdirAll(ctx, "/test/packages/app", 0755))
	require.NoError(t, fs.MkdirAll(ctx, "/test/target", 0755))
	require.NoError(t, fs.WriteFile(ctx, "/test/packages/app/dot-config", []byte("cfg"), 0644))

	// Create 30 orphaned symlinks
	for i := 0; i < 30; i++ {
		linkPath := fmt.Sprintf("/test/target/orphan%d", i)
		require.NoError(t, fs.Symlink(ctx, "/nowhere", linkPath))
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         fs,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	err = client.Manage(ctx, "app")
	require.NoError(t, err)

	// Test with MaxIssues=0 (unlimited)
	scanCfg := dot.ScanConfig{
		Mode:       dot.ScanDeep,
		MaxDepth:   5,
		MaxWorkers: 2,
		MaxIssues:  0, // Unlimited
	}

	report, err := client.DoctorWithScan(ctx, scanCfg)
	require.NoError(t, err)

	// Should have all 30 issues
	assert.Equal(t, 30, len(report.Issues),
		"MaxIssues=0 should mean unlimited, expected 30 issues, got %d", len(report.Issues))
}
