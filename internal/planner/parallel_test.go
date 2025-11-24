package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/domain"
)

func TestParallelizationPlan_EmptyGraph(t *testing.T) {
	graph := BuildGraph([]domain.Operation{})

	batches := graph.ParallelizationPlan()

	assert.Empty(t, batches, "empty graph should produce empty parallelization plan")
}

func TestParallelizationPlan_SingleOperation(t *testing.T) {
	op := domain.NewLinkCreate("link1", mustParsePath("/a"), mustParseTargetPath("/b"))
	graph := BuildGraph([]domain.Operation{op})

	batches := graph.ParallelizationPlan()

	require.Len(t, batches, 1, "single operation should produce single batch")
	require.Len(t, batches[0], 1)
	assert.True(t, op.Equals(batches[0][0]))
}

func TestParallelizationPlan_IndependentOperations(t *testing.T) {
	op1 := domain.NewLinkCreate("link1", mustParsePath("/a"), mustParseTargetPath("/b"))
	op2 := domain.NewLinkCreate("link2", mustParsePath("/c"), mustParseTargetPath("/d"))
	op3 := domain.NewDirCreate("dir1", mustParsePath("/e"))
	graph := BuildGraph([]domain.Operation{op1, op2, op3})

	batches := graph.ParallelizationPlan()

	require.Len(t, batches, 1, "independent operations should be in single batch")
	require.Len(t, batches[0], 3, "all three operations should be in same batch")

	// Verify all operations present
	assert.Contains(t, batches[0], op1)
	assert.Contains(t, batches[0], op2)
	assert.Contains(t, batches[0], op3)
}

func TestParallelizationPlan_LinearChain(t *testing.T) {
	// Create linear dependency: A -> B -> C
	opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewLinkCreate("link1", mustParsePath("/src"), mustParseTargetPath("/c")),
		deps: []domain.Operation{opB},
	}

	graph := BuildGraph([]domain.Operation{opC, opB, opA})

	batches := graph.ParallelizationPlan()

	require.Len(t, batches, 3, "linear chain should produce 3 batches")
	assert.Len(t, batches[0], 1, "batch 0 should have 1 operation")
	assert.Len(t, batches[1], 1, "batch 1 should have 1 operation")
	assert.Len(t, batches[2], 1, "batch 2 should have 1 operation")

	// Verify order: A in batch 0, B in batch 1, C in batch 2
	assert.True(t, opA.Equals(batches[0][0]))
	assert.True(t, opB.Equals(batches[1][0]))
	assert.True(t, opC.Equals(batches[2][0]))
}

func TestParallelizationPlan_DiamondPattern(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewDirCreate("dir3", mustParsePath("/c")),
		deps: []domain.Operation{opA},
	}
	opD := &mockOperation{
		op:   domain.NewLinkCreate("link1", mustParsePath("/src"), mustParseTargetPath("/d")),
		deps: []domain.Operation{opB, opC},
	}

	graph := BuildGraph([]domain.Operation{opD, opC, opB, opA})

	batches := graph.ParallelizationPlan()

	require.Len(t, batches, 3, "diamond should produce 3 levels")

	// Batch 0: A (no dependencies)
	require.Len(t, batches[0], 1)
	assert.True(t, opA.Equals(batches[0][0]))

	// Batch 1: B and C (both depend only on A, can run in parallel)
	require.Len(t, batches[1], 2)
	assert.Contains(t, batches[1], opB)
	assert.Contains(t, batches[1], opC)

	// Batch 2: D (depends on both B and C)
	require.Len(t, batches[2], 1)
	assert.True(t, opD.Equals(batches[2][0]))
}

func TestParallelizationPlan_ComplexGraph(t *testing.T) {
	// More complex graph with multiple parallelization opportunities
	//     A
	//    / \
	//   B   C
	//   |\ /|
	//   | X |
	//   |/ \|
	//   D   E
	//    \ /
	//     F

	opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewDirCreate("dir3", mustParsePath("/c")),
		deps: []domain.Operation{opA},
	}
	opD := &mockOperation{
		op:   domain.NewDirCreate("dir4", mustParsePath("/d")),
		deps: []domain.Operation{opB, opC},
	}
	opE := &mockOperation{
		op:   domain.NewDirCreate("dir5", mustParsePath("/e")),
		deps: []domain.Operation{opB, opC},
	}
	opF := &mockOperation{
		op:   domain.NewLinkCreate("link1", mustParsePath("/src"), mustParseTargetPath("/f")),
		deps: []domain.Operation{opD, opE},
	}

	graph := BuildGraph([]domain.Operation{opF, opE, opD, opC, opB, opA})

	batches := graph.ParallelizationPlan()

	require.Len(t, batches, 4, "complex graph should produce 4 levels")

	// Level 0: A
	assert.Len(t, batches[0], 1)
	assert.True(t, opA.Equals(batches[0][0]))

	// Level 1: B and C
	assert.Len(t, batches[1], 2)
	assert.Contains(t, batches[1], opB)
	assert.Contains(t, batches[1], opC)

	// Level 2: D and E
	assert.Len(t, batches[2], 2)
	assert.Contains(t, batches[2], opD)
	assert.Contains(t, batches[2], opE)

	// Level 3: F
	assert.Len(t, batches[3], 1)
	assert.True(t, opF.Equals(batches[3][0]))
}

func TestParallelizationSafety(t *testing.T) {
	// Create various graph structures and verify safety properties
	testCases := []struct {
		name string
		ops  []domain.Operation
	}{
		{
			name: "simple chain",
			ops: func() []domain.Operation {
				opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
				opB := &mockOperation{
					op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
					deps: []domain.Operation{opA},
				}
				return []domain.Operation{opB, opA}
			}(),
		},
		{
			name: "diamond",
			ops: func() []domain.Operation {
				opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
				opB := &mockOperation{
					op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
					deps: []domain.Operation{opA},
				}
				opC := &mockOperation{
					op:   domain.NewDirCreate("dir3", mustParsePath("/c")),
					deps: []domain.Operation{opA},
				}
				opD := &mockOperation{
					op:   domain.NewLinkCreate("link1", mustParsePath("/src"), mustParseTargetPath("/d")),
					deps: []domain.Operation{opB, opC},
				}
				return []domain.Operation{opD, opC, opB, opA}
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			graph := BuildGraph(tc.ops)
			batches := graph.ParallelizationPlan()

			// Verify safety: no operation in a batch depends on another in same batch
			for batchIdx, batch := range batches {
				for _, op := range batch {
					deps := op.Dependencies()
					for _, dep := range deps {
						// Verify dependency is not in same batch
						depInSameBatch := false
						for _, batchOp := range batch {
							if dep.Equals(batchOp) {
								depInSameBatch = true
								break
							}
						}
						assert.False(t, depInSameBatch,
							"operation %v in batch %d depends on %v which is also in same batch",
							op, batchIdx, dep)
					}
				}
			}
		})
	}
}

func TestParallelizationDependenciesSatisfied(t *testing.T) {
	// Create a complex graph
	opA := domain.NewDirCreate("dir1", mustParsePath("/a"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir2", mustParsePath("/b")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewDirCreate("dir3", mustParsePath("/c")),
		deps: []domain.Operation{opA},
	}
	opD := &mockOperation{
		op:   domain.NewDirCreate("dir4", mustParsePath("/d")),
		deps: []domain.Operation{opB, opC},
	}

	graph := BuildGraph([]domain.Operation{opD, opC, opB, opA})
	batches := graph.ParallelizationPlan()

	// Build a map of operation to batch index
	opToBatch := make(map[domain.Operation]int)
	for batchIdx, batch := range batches {
		for _, op := range batch {
			opToBatch[op] = batchIdx
		}
	}

	// Verify all dependencies of each operation are in earlier batches
	for batchIdx, batch := range batches {
		for _, op := range batch {
			deps := op.Dependencies()
			for _, dep := range deps {
				depBatch, exists := opToBatch[dep]
				require.True(t, exists, "dependency %v not found in any batch", dep)
				assert.Less(t, depBatch, batchIdx,
					"dependency %v (batch %d) should be in earlier batch than %v (batch %d)",
					dep, depBatch, op, batchIdx)
			}
		}
	}
}

func TestParallelizationPlan_LargeGraph(t *testing.T) {
	// Build a larger graph to test performance
	numLevels := 10
	opsPerLevel := 10
	var ops []domain.Operation
	levels := make([][]domain.Operation, numLevels)

	// Level 0: independent operations
	idCounter := 0
	for i := 0; i < opsPerLevel; i++ {
		idCounter++
		opID := domain.OperationID(formatPath("dir-%d", idCounter))
		op := domain.NewDirCreate(opID, mustParsePath(formatPath("/level0/op", i)))
		levels[0] = append(levels[0], op)
		ops = append(ops, op)
	}

	// Each subsequent level depends on all operations from previous level
	for level := 1; level < numLevels; level++ {
		for i := 0; i < opsPerLevel; i++ {
			idCounter++
			opID := domain.OperationID(formatPath("dir-%d", idCounter))
			op := &mockOperation{
				op:   domain.NewDirCreate(opID, mustParsePath(formatPath("/level%d/op", level, i))),
				deps: levels[level-1],
			}
			levels[level] = append(levels[level], op)
			ops = append(ops, op)
		}
	}

	graph := BuildGraph(ops)
	batches := graph.ParallelizationPlan()

	require.Len(t, batches, numLevels, "should have one batch per level")
	for i, batch := range batches {
		assert.Len(t, batch, opsPerLevel, "level %d should have %d operations", i, opsPerLevel)
	}
}

// formatPath is a helper for creating paths in tests
func formatPath(template string, args ...interface{}) string {
	// Simple string formatting for test paths
	result := template
	for i, arg := range args {
		// Simple replacement - not using fmt to avoid import
		_ = i
		_ = arg
	}
	// For simplicity in tests, just append numbers
	if len(args) > 0 {
		result = template + string(rune('0'+args[0].(int)))
	}
	if len(args) > 1 {
		result += string(rune('0' + args[1].(int)))
	}
	return result
}
