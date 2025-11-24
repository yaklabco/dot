package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestScanMode_String(t *testing.T) {
	tests := []struct {
		mode dot.ScanMode
		want string
	}{
		{dot.ScanOff, "off"},
		{dot.ScanScoped, "scoped"},
		{dot.ScanDeep, "deep"},
		{dot.ScanMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultScanConfig(t *testing.T) {
	cfg := dot.DefaultScanConfig()

	// Default is now scoped scanning (not off)
	assert.Equal(t, dot.ScanScoped, cfg.Mode)
	assert.Equal(t, 3, cfg.MaxDepth) // Reduced from 10 to 3 for better performance
	assert.Nil(t, cfg.ScopeToDirs)
	assert.NotEmpty(t, cfg.SkipPatterns)
	assert.Contains(t, cfg.SkipPatterns, ".git")
	assert.Contains(t, cfg.SkipPatterns, "node_modules")
	// New performance features
	assert.Equal(t, 0, cfg.MaxWorkers) // 0 means use NumCPU
	assert.Equal(t, 0, cfg.MaxIssues)  // 0 means unlimited
	// Check expanded skip patterns
	assert.Contains(t, cfg.SkipPatterns, "Library")
	assert.Contains(t, cfg.SkipPatterns, ".docker")
	assert.Contains(t, cfg.SkipPatterns, ".pyenv")
}

func TestScopedScanConfig(t *testing.T) {
	cfg := dot.ScopedScanConfig()

	assert.Equal(t, dot.ScanScoped, cfg.Mode)
	assert.Equal(t, 3, cfg.MaxDepth) // Reduced for performance
	assert.Nil(t, cfg.ScopeToDirs)
	assert.NotEmpty(t, cfg.SkipPatterns)
	assert.Equal(t, 0, cfg.MaxWorkers)
	assert.Equal(t, 0, cfg.MaxIssues)
}

func TestDeepScanConfig(t *testing.T) {
	t.Run("with positive depth", func(t *testing.T) {
		cfg := dot.DeepScanConfig(15)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 15, cfg.MaxDepth)
		assert.Nil(t, cfg.ScopeToDirs)
		assert.NotEmpty(t, cfg.SkipPatterns)
	})

	t.Run("with zero depth defaults to 10", func(t *testing.T) {
		cfg := dot.DeepScanConfig(0)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 10, cfg.MaxDepth)
	})

	t.Run("with negative depth defaults to 10", func(t *testing.T) {
		cfg := dot.DeepScanConfig(-5)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 10, cfg.MaxDepth)
	})
}

func TestScanConfig_CustomConfiguration(t *testing.T) {
	cfg := dot.ScanConfig{
		Mode:         dot.ScanScoped,
		MaxDepth:     5,
		ScopeToDirs:  []string{"/home/user/.config", "/home/user/.local"},
		SkipPatterns: []string{".git", "target"},
		MaxWorkers:   4,
		MaxIssues:    100,
	}

	assert.Equal(t, dot.ScanScoped, cfg.Mode)
	assert.Equal(t, 5, cfg.MaxDepth)
	assert.Len(t, cfg.ScopeToDirs, 2)
	assert.Len(t, cfg.SkipPatterns, 2)
	assert.Equal(t, 4, cfg.MaxWorkers)
	assert.Equal(t, 100, cfg.MaxIssues)
}

func TestScanConfig_PerformanceFeatures(t *testing.T) {
	t.Run("MaxWorkers for parallel scanning", func(t *testing.T) {
		cfg := dot.ScanConfig{
			Mode:       dot.ScanScoped,
			MaxDepth:   5,
			MaxWorkers: 8, // Explicit worker count
		}

		assert.Equal(t, 8, cfg.MaxWorkers)
	})

	t.Run("MaxIssues for early termination", func(t *testing.T) {
		cfg := dot.ScanConfig{
			Mode:      dot.ScanScoped,
			MaxDepth:  5,
			MaxIssues: 50, // Stop after 50 issues
		}

		assert.Equal(t, 50, cfg.MaxIssues)
	})

	t.Run("sequential scan with MaxWorkers=1", func(t *testing.T) {
		cfg := dot.ScanConfig{
			Mode:       dot.ScanScoped,
			MaxDepth:   5,
			MaxWorkers: 1, // Force sequential
		}

		assert.Equal(t, 1, cfg.MaxWorkers)
	})
}
