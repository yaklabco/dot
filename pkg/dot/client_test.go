package dot_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// TestNewClient verifies Client creation.
func TestNewClient(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         adapters.NewMemFS(),
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify config is stored
	clientCfg := client.Config()
	require.Equal(t, cfg.PackageDir, clientCfg.PackageDir)
	require.Equal(t, cfg.TargetDir, clientCfg.TargetDir)
}

// TestNewClient_InvalidConfig verifies validation errors.
func TestNewClient_InvalidConfig(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "relative/path", // Invalid - not absolute
	}

	client, err := dot.NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "invalid configuration")
}

// =============================================================================
// API Contract Tests
// =============================================================================

// TestClient_MethodSignatures is a compile-time check that public method
// signatures don't change unexpectedly. If any signature changes, this test
// will fail to compile.
func TestClient_MethodSignatures(t *testing.T) {
	var client *dot.Client

	// Core management methods
	var _ func(context.Context, ...string) error = client.Manage
	var _ func(context.Context, ...string) (dot.Plan, error) = client.PlanManage
	var _ func(context.Context, ...string) error = client.Unmanage
	var _ func(context.Context, dot.UnmanageOptions, ...string) error = client.UnmanageWithOptions
	var _ func(context.Context, dot.UnmanageOptions) (int, error) = client.UnmanageAll
	var _ func(context.Context, ...string) (dot.Plan, error) = client.PlanUnmanage
	var _ func(context.Context, ...string) error = client.Remanage
	var _ func(context.Context, ...string) (dot.Plan, error) = client.PlanRemanage

	// Status methods
	var _ func(context.Context, ...string) (dot.Status, error) = client.Status
	var _ func(context.Context) ([]dot.PackageInfo, error) = client.List

	// Adoption methods
	var _ func(context.Context, []string, string) error = client.Adopt
	var _ func(context.Context, []string, string) (dot.Plan, error) = client.PlanAdopt

	// Doctor methods
	var _ func(context.Context) (dot.DiagnosticReport, error) = client.Doctor
	var _ func(context.Context, dot.ScanConfig) (dot.DiagnosticReport, error) = client.DoctorWithScan
	var _ func(context.Context, dot.DiagnosticMode, dot.ScanConfig) (dot.DiagnosticReport, error) = client.DoctorWithMode
	var _ func(context.Context, dot.ScanConfig, dot.TriageOptions) (dot.TriageResult, error) = client.Triage

	// Clone methods
	var _ func(context.Context, string, dot.CloneOptions) error = client.Clone

	// Bootstrap methods
	var _ func(context.Context, dot.GenerateBootstrapOptions) (dot.BootstrapResult, error) = client.GenerateBootstrap
	var _ func(context.Context, []byte, string) error = client.WriteBootstrap

	// Config accessor
	var _ func() dot.Config = client.Config

	// This test passes if it compiles - the type assertions above verify
	// that the method signatures match exactly what we expect
	t.Log("All public method signatures verified")
}

// TestClient_ErrorTypes_SupportErrorsIs verifies that all exported error types
// properly support errors.Is for wrapped error detection.
func TestClient_ErrorTypes_SupportErrorsIs(t *testing.T) {
	// Error types exported from pkg/dot/errors.go
	errorTypes := []struct {
		name  string
		err   error
		maker func() error
	}{
		{
			name:  "ErrPackageDirNotEmpty",
			err:   dot.ErrPackageDirNotEmpty{},
			maker: func() error { return dot.ErrPackageDirNotEmpty{Path: "/test"} },
		},
		{
			name:  "ErrBootstrapNotFound",
			err:   dot.ErrBootstrapNotFound{},
			maker: func() error { return dot.ErrBootstrapNotFound{Path: "/test"} },
		},
		{
			name:  "ErrInvalidBootstrap",
			err:   dot.ErrInvalidBootstrap{},
			maker: func() error { return dot.ErrInvalidBootstrap{Reason: "test"} },
		},
		{
			name:  "ErrAuthFailed",
			err:   dot.ErrAuthFailed{},
			maker: func() error { return dot.ErrAuthFailed{Cause: fmt.Errorf("auth error")} },
		},
		{
			name:  "ErrCloneFailed",
			err:   dot.ErrCloneFailed{},
			maker: func() error { return dot.ErrCloneFailed{URL: "git@example.com:test/repo"} },
		},
		{
			name:  "ErrProfileNotFound",
			err:   dot.ErrProfileNotFound{},
			maker: func() error { return dot.ErrProfileNotFound{Profile: "work"} },
		},
		{
			name:  "ErrBootstrapExists",
			err:   dot.ErrBootstrapExists{},
			maker: func() error { return dot.ErrBootstrapExists{Path: "/test/.bootstrap.yml"} },
		},
	}

	for _, tc := range errorTypes {
		t.Run(tc.name, func(t *testing.T) {
			// Create an instance of the error
			errInstance := tc.maker()

			// Wrap it
			wrapped := fmt.Errorf("wrapped: %w", errInstance)

			// Verify errors.Is works with the wrapped error
			assert.True(t, errors.Is(wrapped, tc.err),
				"errors.Is should match wrapped %s", tc.name)

			// Verify direct match also works
			assert.True(t, errors.Is(errInstance, tc.err),
				"errors.Is should match direct %s", tc.name)
		})
	}
}

// TestClient_DryRun_NoSideEffects verifies that operations in dry-run mode
// do not produce any filesystem side effects.
func TestClient_DryRun_NoSideEffects(t *testing.T) {
	// Create real temporary directories for this test
	tmpDir := t.TempDir()
	targetDir := t.TempDir()

	// Create a test package with a file
	pkgDir := filepath.Join(tmpDir, "dot-test")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))
	testFile := filepath.Join(pkgDir, ".testrc")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0o644))

	// Create client with DryRun enabled
	cfg := dot.Config{
		PackageDir: tmpDir,
		TargetDir:  targetDir,
		FS:         adapters.NewOSFilesystem(),
		Logger:     adapters.NewNoopLogger(),
		DryRun:     true,
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Run manage in dry-run mode
	ctx := context.Background()
	err = client.Manage(ctx, "dot-test")
	require.NoError(t, err)

	// Verify no symlinks were created in target directory
	entries, err := os.ReadDir(targetDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "dry-run should not create any files in target directory")

	// Also verify PlanManage returns a valid plan even in dry-run
	plan, err := client.PlanManage(ctx, "dot-test")
	require.NoError(t, err)
	assert.NotNil(t, plan, "PlanManage should return a plan even in dry-run mode")
}
