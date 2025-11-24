package dot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/pkg/dot"
)

// Test that Doctor properly propagates non-not-found manifest errors

type errorFS struct {
	*adapters.MemFS
	readFileError error
}

func (e *errorFS) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if e.readFileError != nil && path == "/test/target/.dot-manifest.json" {
		return nil, e.readFileError
	}
	return e.MemFS.ReadFile(ctx, path)
}

func TestClient_Doctor_ManifestReadError(t *testing.T) {
	// Create FS that will fail with permission error
	memFS := adapters.NewMemFS()
	ctx := context.Background()

	require.NoError(t, memFS.MkdirAll(ctx, "/test/packages", 0755))
	require.NoError(t, memFS.MkdirAll(ctx, "/test/target", 0755))

	// Use error FS that returns permission error (not not-found)
	errFS := &errorFS{
		MemFS:         memFS,
		readFileError: errors.New("permission denied"),
	}

	cfg := dot.Config{
		PackageDir: "/test/packages",
		TargetDir:  "/test/target",
		FS:         errFS,
		Logger:     adapters.NewNoopLogger(),
	}

	client, err := dot.NewClient(cfg)
	require.NoError(t, err)

	// Doctor should propagate the permission error, not return OK
	report, err := client.Doctor(ctx)
	require.Error(t, err, "Expected error to be propagated")
	require.Contains(t, err.Error(), "permission denied")
	require.Equal(t, dot.DiagnosticReport{}, report)
}
