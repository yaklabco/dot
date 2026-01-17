package dot_test

import (
	"bytes"
	"os"
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

func TestConfig_StdinStdout_Defaults(t *testing.T) {
	cfg := &dot.Config{}

	// Should return os.Stdin/os.Stdout when not set
	assert.Equal(t, os.Stdin, cfg.GetStdin())
	assert.Equal(t, os.Stdout, cfg.GetStdout())
}

func TestConfig_StdinStdout_Custom(t *testing.T) {
	stdin := bytes.NewBufferString("test input\n")
	stdout := &bytes.Buffer{}

	cfg := &dot.Config{
		Stdin:  stdin,
		Stdout: stdout,
	}

	assert.Equal(t, stdin, cfg.GetStdin())
	assert.Equal(t, stdout, cfg.GetStdout())
}
