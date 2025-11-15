package dot

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDoctorService_buildIgnoreSet tests ignore set building from manifest.
func TestDoctorService_buildIgnoreSet(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, store)

	svc := newDoctorService(fs, logger, manifestSvc, "/packages", "/home")

	t.Run("empty manifest", func(t *testing.T) {
		m := manifest.New()
		m.Doctor = nil

		ignoreSet := svc.buildIgnoreSet(&m)
		assert.NotNil(t, ignoreSet)

		// Should not ignore anything
		assert.False(t, ignoreSet.ShouldIgnore(".cache/file"))
	})

	t.Run("with patterns", func(t *testing.T) {
		m := manifest.New()
		m.AddIgnoredPattern(".cache/**")
		m.AddIgnoredPattern(".tmp/**")

		ignoreSet := svc.buildIgnoreSet(&m)
		assert.NotNil(t, ignoreSet)

		// Should ignore matching paths
		assert.True(t, ignoreSet.ShouldIgnore(".cache/file"))
		assert.True(t, ignoreSet.ShouldIgnore(".tmp/other"))

		// Should not ignore non-matching
		assert.False(t, ignoreSet.ShouldIgnore(".config"))
	})
}

// TestDoctorService_calculateDepth tests depth calculation.
func TestDoctorService_calculateDepth(t *testing.T) {
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, store)

	svc := newDoctorService(fs, logger, manifestSvc, "/packages", "/home")

	tests := []struct {
		name     string
		path     string
		root     string
		expected int
	}{
		{"same path", "/home", "/home", 0},
		{"one level deep", "/home/user", "/home", 1},
		{"two levels deep", "/home/user/docs", "/home", 2},
		{"three levels deep", "/home/user/docs/files", "/home", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth := svc.calculateDepth(tt.path, tt.root)
			assert.Equal(t, tt.expected, depth)
		})
	}
}

// TestDoctorService_shouldSkipDirectory tests directory skip logic.
func TestDoctorService_shouldSkipDirectory(t *testing.T) {
	ctx := context.Background()
	fs := adapters.NewMemFS()
	logger := adapters.NewNoopLogger()
	store := manifest.NewFSManifestStore(fs)
	manifestSvc := newManifestService(fs, logger, store)
	targetDir := "/home"

	require.NoError(t, fs.MkdirAll(ctx, targetDir, 0755))

	svc := newDoctorService(fs, logger, manifestSvc, "/packages", targetDir)

	m := manifest.New()
	managedDirs := map[string]bool{
		"/home/managed": true,
	}

	t.Run("skip hidden directories", func(t *testing.T) {
		skip := svc.shouldSkipDirectory(".git", managedDirs, &m)
		assert.True(t, skip, "should skip .git")

		skip = svc.shouldSkipDirectory(".svn", managedDirs, &m)
		assert.True(t, skip, "should skip .svn")
	})

	t.Run("skip managed directories", func(t *testing.T) {
		skip := svc.shouldSkipDirectory("/home/managed", managedDirs, &m)
		assert.True(t, skip, "should skip managed directory")
	})

	t.Run("do not skip regular directories", func(t *testing.T) {
		skip := svc.shouldSkipDirectory("/home/documents", managedDirs, &m)
		assert.False(t, skip, "should not skip regular directory")
	})
}

