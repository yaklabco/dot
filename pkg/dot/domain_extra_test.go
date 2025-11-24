package dot_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaklabco/dot/pkg/dot"
)

func TestPlan_CanParallelize(t *testing.T) {
	plan := dot.Plan{}
	canParallelize := plan.CanParallelize()
	// Returns false for empty plan
	assert.False(t, canParallelize)
}

func TestPlan_ParallelBatches(t *testing.T) {
	plan := dot.Plan{}
	batches := plan.ParallelBatches()
	assert.Empty(t, batches)
}
