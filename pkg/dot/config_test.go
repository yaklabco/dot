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

func TestConfigBuilder_DefaultsApplied(t *testing.T) {
	// When no optional bools are set, Build() should apply defaults
	cfg := dot.NewConfigBuilder().
		WithPackageDir("/packages").
		WithTargetDir("/target").
		Build()

	// These should default to true
	assert.True(t, cfg.PackageNameMapping, "PackageNameMapping should default to true")
	assert.True(t, cfg.UseDefaultIgnorePatterns, "UseDefaultIgnorePatterns should default to true")
	assert.True(t, cfg.PerPackageIgnore, "PerPackageIgnore should default to true")
	assert.True(t, cfg.InteractiveLargeFiles, "InteractiveLargeFiles should default to true")

	// These should default to false (Go zero value)
	assert.False(t, cfg.DryRun, "DryRun should default to false")
	assert.False(t, cfg.Folding, "Folding should default to false")
	assert.False(t, cfg.Backup, "Backup should default to false")
	assert.False(t, cfg.Overwrite, "Overwrite should default to false")
}

func TestConfigBuilder_ExplicitFalse(t *testing.T) {
	// When optional bools are explicitly set to false, Build() should preserve that
	cfg := dot.NewConfigBuilder().
		WithPackageDir("/packages").
		WithTargetDir("/target").
		WithPackageNameMapping(false).
		WithUseDefaultIgnorePatterns(false).
		WithPerPackageIgnore(false).
		WithInteractiveLargeFiles(false).
		Build()

	// These should be false because we explicitly set them
	assert.False(t, cfg.PackageNameMapping, "PackageNameMapping should be false when explicitly set")
	assert.False(t, cfg.UseDefaultIgnorePatterns, "UseDefaultIgnorePatterns should be false when explicitly set")
	assert.False(t, cfg.PerPackageIgnore, "PerPackageIgnore should be false when explicitly set")
	assert.False(t, cfg.InteractiveLargeFiles, "InteractiveLargeFiles should be false when explicitly set")
}

func TestConfigBuilder_ExplicitTrue(t *testing.T) {
	// When optional bools are explicitly set to true, Build() should preserve that
	cfg := dot.NewConfigBuilder().
		WithPackageDir("/packages").
		WithTargetDir("/target").
		WithDryRun(true).
		WithFolding(true).
		WithBackup(true).
		WithOverwrite(true).
		Build()

	// These should be true because we explicitly set them
	assert.True(t, cfg.DryRun, "DryRun should be true when explicitly set")
	assert.True(t, cfg.Folding, "Folding should be true when explicitly set")
	assert.True(t, cfg.Backup, "Backup should be true when explicitly set")
	assert.True(t, cfg.Overwrite, "Overwrite should be true when explicitly set")
}

func TestConfigBuilder_IsSetTracking(t *testing.T) {
	builder := dot.NewConfigBuilder()

	// Initially nothing is set
	assert.False(t, builder.IsFoldingSet())
	assert.False(t, builder.IsDryRunSet())
	assert.False(t, builder.IsBackupSet())
	assert.False(t, builder.IsOverwriteSet())
	assert.False(t, builder.IsPackageNameMappingSet())
	assert.False(t, builder.IsUseDefaultIgnorePatternsSet())
	assert.False(t, builder.IsPerPackageIgnoreSet())
	assert.False(t, builder.IsInteractiveLargeFilesSet())

	// Set each one
	builder.WithFolding(false)
	builder.WithDryRun(false)
	builder.WithBackup(false)
	builder.WithOverwrite(false)
	builder.WithPackageNameMapping(false)
	builder.WithUseDefaultIgnorePatterns(false)
	builder.WithPerPackageIgnore(false)
	builder.WithInteractiveLargeFiles(false)

	// Now all should be set
	assert.True(t, builder.IsFoldingSet())
	assert.True(t, builder.IsDryRunSet())
	assert.True(t, builder.IsBackupSet())
	assert.True(t, builder.IsOverwriteSet())
	assert.True(t, builder.IsPackageNameMappingSet())
	assert.True(t, builder.IsUseDefaultIgnorePatternsSet())
	assert.True(t, builder.IsPerPackageIgnoreSet())
	assert.True(t, builder.IsInteractiveLargeFilesSet())
}

func TestConfigBuilder_BuildRaw(t *testing.T) {
	// BuildRaw should not apply defaults
	cfg := dot.NewConfigBuilder().
		WithPackageDir("/packages").
		WithTargetDir("/target").
		BuildRaw()

	// These should all be false (Go zero values) since BuildRaw doesn't apply defaults
	assert.False(t, cfg.PackageNameMapping, "BuildRaw should not apply defaults")
	assert.False(t, cfg.UseDefaultIgnorePatterns, "BuildRaw should not apply defaults")
	assert.False(t, cfg.PerPackageIgnore, "BuildRaw should not apply defaults")
	assert.False(t, cfg.InteractiveLargeFiles, "BuildRaw should not apply defaults")
}

func TestConfigBuilder_AllFields(t *testing.T) {
	stdin := bytes.NewBufferString("test")
	stdout := &bytes.Buffer{}

	cfg := dot.NewConfigBuilder().
		WithPackageDir("/packages").
		WithTargetDir("/target").
		WithLinkMode(dot.LinkAbsolute).
		WithFolding(true).
		WithDryRun(true).
		WithVerbosity(2).
		WithBackupDir("/backup").
		WithBackup(true).
		WithOverwrite(true).
		WithManifestDir("/manifest").
		WithConcurrency(4).
		WithPackageNameMapping(true).
		WithIgnorePatterns([]string{"*.tmp", "*.log"}).
		WithUseDefaultIgnorePatterns(true).
		WithPerPackageIgnore(true).
		WithMaxFileSize(1024).
		WithInteractiveLargeFiles(true).
		WithStdin(stdin).
		WithStdout(stdout).
		Build()

	assert.Equal(t, "/packages", cfg.PackageDir)
	assert.Equal(t, "/target", cfg.TargetDir)
	assert.Equal(t, dot.LinkAbsolute, cfg.LinkMode)
	assert.True(t, cfg.Folding)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, 2, cfg.Verbosity)
	assert.Equal(t, "/backup", cfg.BackupDir)
	assert.True(t, cfg.Backup)
	assert.True(t, cfg.Overwrite)
	assert.Equal(t, "/manifest", cfg.ManifestDir)
	assert.Equal(t, 4, cfg.Concurrency)
	assert.True(t, cfg.PackageNameMapping)
	assert.Equal(t, []string{"*.tmp", "*.log"}, cfg.IgnorePatterns)
	assert.True(t, cfg.UseDefaultIgnorePatterns)
	assert.True(t, cfg.PerPackageIgnore)
	assert.Equal(t, int64(1024), cfg.MaxFileSize)
	assert.True(t, cfg.InteractiveLargeFiles)
	assert.Equal(t, stdin, cfg.Stdin)
	assert.Equal(t, stdout, cfg.Stdout)
}

func TestConfigBuilder_FluentChaining(t *testing.T) {
	// Verify fluent chaining returns the same builder
	builder := dot.NewConfigBuilder()

	result := builder.
		WithPackageDir("/packages").
		WithTargetDir("/target").
		WithDryRun(true)

	assert.Same(t, builder, result, "Fluent methods should return the same builder instance")
}
