package dot

import (
	"testing"

	"github.com/jamesainslie/dot/internal/adapters"
	"github.com/jamesainslie/dot/internal/manifest"
	"github.com/stretchr/testify/assert"
)

// TestDoctorService_buildIgnoreSet tests ignore set building from manifest.
func TestDoctorService_buildIgnoreSet(t *testing.T) {
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

	t.Run("with invalid pattern", func(t *testing.T) {
		m := manifest.New()
		// Add a valid pattern
		m.AddIgnoredPattern(".cache/**")
		// Add an invalid pattern (will be ignored gracefully)
		m.AddIgnoredPattern("[invalid")

		ignoreSet := svc.buildIgnoreSet(&m)
		assert.NotNil(t, ignoreSet)

		// Valid pattern should still work
		assert.True(t, ignoreSet.ShouldIgnore(".cache/file"))
	})
}
