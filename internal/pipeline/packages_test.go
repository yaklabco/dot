package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/adapters"
	"github.com/yaklabco/dot/internal/domain"
	"github.com/yaklabco/dot/internal/ignore"
	"github.com/yaklabco/dot/internal/planner"
)

func TestNewManagePipeline(t *testing.T) {
	fs := adapters.NewOSFilesystem()
	ignoreSet := ignore.NewIgnoreSet()
	policies := planner.DefaultPolicies()

	pipeline := NewManagePipeline(ManagePipelineOpts{
		FS:                 fs,
		IgnoreSet:          ignoreSet,
		Policies:           policies,
		PackageNameMapping: false,
	})

	require.NotNil(t, pipeline)
	assert.Equal(t, fs, pipeline.opts.FS)
	assert.Equal(t, ignoreSet, pipeline.opts.IgnoreSet)
	assert.Equal(t, policies, pipeline.opts.Policies)
}

func TestManagePipeline_Execute(t *testing.T) {
	t.Run("empty package list", func(t *testing.T) {
		fs := adapters.NewOSFilesystem()
		ignoreSet := ignore.NewIgnoreSet()

		pipeline := NewManagePipeline(ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignoreSet,
			Policies:           planner.DefaultPolicies(),
			PackageNameMapping: false,
		})

		packagePathResult := domain.NewPackagePath("/packages")
		require.True(t, packagePathResult.IsOk(), "failed to create package path")
		packagePath := packagePathResult.Unwrap()

		targetPathResult := domain.NewTargetPath("/target")
		require.True(t, targetPathResult.IsOk(), "failed to create target path")
		targetPath := targetPathResult.Unwrap()

		result := pipeline.Execute(context.Background(), ManageInput{
			PackageDir: packagePath,
			TargetDir:  targetPath,
			Packages:   []string{},
		})

		require.True(t, result.IsOk())
		plan := result.Unwrap()
		assert.Empty(t, plan.Operations)

		// Verify metadata for empty package list
		assert.Equal(t, 0, plan.Metadata.PackageCount)
		assert.Equal(t, 0, plan.Metadata.OperationCount)
		assert.Empty(t, plan.Metadata.Conflicts)
		assert.Empty(t, plan.Metadata.Warnings)
	})

	t.Run("package not found", func(t *testing.T) {
		fs := adapters.NewOSFilesystem()
		ignoreSet := ignore.NewIgnoreSet()

		pipeline := NewManagePipeline(ManagePipelineOpts{
			FS:                 fs,
			IgnoreSet:          ignoreSet,
			Policies:           planner.DefaultPolicies(),
			PackageNameMapping: false,
		})

		packagePathResult := domain.NewPackagePath("/packages")
		require.True(t, packagePathResult.IsOk(), "failed to create package path")
		packagePath := packagePathResult.Unwrap()

		targetPathResult := domain.NewTargetPath("/target")
		require.True(t, targetPathResult.IsOk(), "failed to create target path")
		targetPath := targetPathResult.Unwrap()

		result := pipeline.Execute(context.Background(), ManageInput{
			PackageDir: packagePath,
			TargetDir:  targetPath,
			Packages:   []string{"nonexistent"},
		})

		require.False(t, result.IsOk())
		err := result.UnwrapErr()

		var pkgErr domain.ErrPackageNotFound
		assert.ErrorAs(t, err, &pkgErr)
	})
}
