package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/pkg/dot"
)

func TestPlan_PackageOperations(t *testing.T) {
	// Create test operations
	sourceResult := dot.NewFilePath("/packages/vim/dot-vimrc")
	require.True(t, sourceResult.IsOk())
	source := sourceResult.Unwrap()

	targetResult := dot.NewTargetPath("/home/user/.vimrc")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op1 := dot.NewLinkCreate(dot.OperationID("vim-link-1"), source, target)
	op2 := dot.NewLinkCreate(dot.OperationID("vim-link-2"), source, target)
	op3 := dot.NewLinkCreate(dot.OperationID("zsh-link-1"), source, target)

	// Create plan with package mappings
	plan := dot.Plan{
		Operations: []dot.Operation{op1, op2, op3},
		Metadata: dot.PlanMetadata{
			PackageCount:   2,
			OperationCount: 3,
		},
		PackageOperations: map[string][]dot.OperationID{
			"vim": {dot.OperationID("vim-link-1"), dot.OperationID("vim-link-2")},
			"zsh": {dot.OperationID("zsh-link-1")},
		},
	}

	// Test vim operations
	vimOps := plan.OperationsForPackage("vim")
	assert.Len(t, vimOps, 2)
	assert.Equal(t, dot.OperationID("vim-link-1"), vimOps[0].ID())
	assert.Equal(t, dot.OperationID("vim-link-2"), vimOps[1].ID())

	// Test zsh operations
	zshOps := plan.OperationsForPackage("zsh")
	assert.Len(t, zshOps, 1)
	assert.Equal(t, dot.OperationID("zsh-link-1"), zshOps[0].ID())

	// Test non-existent package
	gitOps := plan.OperationsForPackage("git")
	assert.Len(t, gitOps, 0)
}

func TestPlan_OperationsForPackage_EmptyPlan(t *testing.T) {
	plan := dot.Plan{
		Operations:        []dot.Operation{},
		PackageOperations: map[string][]dot.OperationID{},
	}

	ops := plan.OperationsForPackage("vim")
	assert.Len(t, ops, 0)
}

func TestPlan_OperationsForPackage_NoMapping(t *testing.T) {
	sourceResult := dot.NewFilePath("/packages/vim/dot-vimrc")
	require.True(t, sourceResult.IsOk())
	source := sourceResult.Unwrap()

	targetResult := dot.NewTargetPath("/home/user/.vimrc")
	require.True(t, targetResult.IsOk())
	target := targetResult.Unwrap()

	op := dot.NewLinkCreate(dot.OperationID("link-1"), source, target)

	// Plan without PackageOperations (backward compatibility)
	plan := dot.Plan{
		Operations: []dot.Operation{op},
	}

	ops := plan.OperationsForPackage("vim")
	assert.Len(t, ops, 0, "should return empty when no mapping exists")
}

func TestPlan_PackageNames(t *testing.T) {
	plan := dot.Plan{
		PackageOperations: map[string][]dot.OperationID{
			"vim":  {dot.OperationID("vim-1")},
			"zsh":  {dot.OperationID("zsh-1")},
			"git":  {dot.OperationID("git-1")},
			"tmux": {dot.OperationID("tmux-1")},
		},
	}

	names := plan.PackageNames()
	assert.Len(t, names, 4)
	assert.Contains(t, names, "vim")
	assert.Contains(t, names, "zsh")
	assert.Contains(t, names, "git")
	assert.Contains(t, names, "tmux")
}

func TestPlan_PackageNames_Empty(t *testing.T) {
	plan := dot.Plan{
		PackageOperations: map[string][]dot.OperationID{},
	}

	names := plan.PackageNames()
	assert.Len(t, names, 0)
}

func TestPlan_PackageNames_Nil(t *testing.T) {
	plan := dot.Plan{
		PackageOperations: nil,
	}

	names := plan.PackageNames()
	assert.Len(t, names, 0)
}

func TestPlan_HasPackage(t *testing.T) {
	plan := dot.Plan{
		PackageOperations: map[string][]dot.OperationID{
			"vim": {dot.OperationID("vim-1")},
			"zsh": {dot.OperationID("zsh-1")},
		},
	}

	assert.True(t, plan.HasPackage("vim"))
	assert.True(t, plan.HasPackage("zsh"))
	assert.False(t, plan.HasPackage("git"))
}

func TestPlan_OperationCount_ByPackage(t *testing.T) {
	plan := dot.Plan{
		PackageOperations: map[string][]dot.OperationID{
			"vim": {
				dot.OperationID("vim-1"),
				dot.OperationID("vim-2"),
				dot.OperationID("vim-3"),
			},
			"zsh": {
				dot.OperationID("zsh-1"),
				dot.OperationID("zsh-2"),
			},
		},
	}

	assert.Equal(t, 3, plan.OperationCountForPackage("vim"))
	assert.Equal(t, 2, plan.OperationCountForPackage("zsh"))
	assert.Equal(t, 0, plan.OperationCountForPackage("git"))
}
