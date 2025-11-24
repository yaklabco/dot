package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestConfig_WithDefaults(t *testing.T) {
	cfg := dot.Config{
		PackageDir: "/packages",
		TargetDir:  "/target",
	}

	cfg = cfg.WithDefaults()

	// WithDefaults sets Tracer and Metrics, BackupDir, Concurrency
	assert.NotNil(t, cfg.Tracer)
	assert.NotNil(t, cfg.Metrics)
	assert.NotEmpty(t, cfg.BackupDir)
	assert.Greater(t, cfg.Concurrency, 0)
}
