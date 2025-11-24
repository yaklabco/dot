package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaklabco/dot/internal/domain"
)

func TestTopologicalSort_EmptyGraph(t *testing.T) {
	graph := BuildGraph([]domain.Operation{})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	assert.Empty(t, sorted, "empty graph should produce empty result")
}

func TestTopologicalSort_SingleNode(t *testing.T) {
	op := domain.NewLinkCreate("link-auto", mustParsePath("/a"), mustParseTargetPath("/b"))
	graph := BuildGraph([]domain.Operation{op})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 1)
	assert.True(t, op.Equals(sorted[0]))
}

func TestTopologicalSort_IndependentNodes(t *testing.T) {
	op1 := domain.NewLinkCreate("link-auto", mustParsePath("/a"), mustParseTargetPath("/b"))
	op2 := domain.NewLinkCreate("link-auto", mustParsePath("/c"), mustParseTargetPath("/d"))
	graph := BuildGraph([]domain.Operation{op1, op2})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 2)
	// Both operations should be present (order doesn't matter for independent ops)
	assert.Contains(t, sorted, op1)
	assert.Contains(t, sorted, op2)
}

func TestTopologicalSort_LinearChain(t *testing.T) {
	// Create linear dependency: A -> B -> C
	opA := domain.NewDirCreate("dir-auto", mustParsePath("/dir1"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/dir1/dir2")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewLinkCreate("link-auto", mustParsePath("/src"), mustParseTargetPath("/dir1/dir2/file")),
		deps: []domain.Operation{opB},
	}

	graph := BuildGraph([]domain.Operation{opC, opB, opA})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 3)

	// Verify order: A must come before B, B must come before C
	posA := findOperationIndex(sorted, opA)
	posB := findOperationIndex(sorted, opB)
	posC := findOperationIndex(sorted, opC)

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posB, posC, "B should come before C")
}

func TestTopologicalSort_DiamondPattern(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	opA := domain.NewDirCreate("dir-auto", mustParsePath("/root"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/root/dir1")),
		deps: []domain.Operation{opA},
	}
	opC := &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/root/dir2")),
		deps: []domain.Operation{opA},
	}
	opD := &mockOperation{
		op:   domain.NewLinkCreate("link-auto", mustParsePath("/src"), mustParseTargetPath("/root/file")),
		deps: []domain.Operation{opB, opC},
	}

	graph := BuildGraph([]domain.Operation{opD, opC, opB, opA})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 4)

	// Verify dependencies are satisfied
	posA := findOperationIndex(sorted, opA)
	posB := findOperationIndex(sorted, opB)
	posC := findOperationIndex(sorted, opC)
	posD := findOperationIndex(sorted, opD)

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posA, posC, "A should come before C")
	assert.Less(t, posB, posD, "B should come before D")
	assert.Less(t, posC, posD, "C should come before D")
}

func TestTopologicalSort_ComplexGraph(t *testing.T) {
	// More complex graph with multiple dependencies
	ops := make([]domain.Operation, 6)
	ops[0] = domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	ops[1] = &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/b")),
		deps: []domain.Operation{ops[0]},
	}
	ops[2] = &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/c")),
		deps: []domain.Operation{ops[0]},
	}
	ops[3] = &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/d")),
		deps: []domain.Operation{ops[1], ops[2]},
	}
	ops[4] = &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/e")),
		deps: []domain.Operation{ops[2]},
	}
	ops[5] = &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/f")),
		deps: []domain.Operation{ops[3], ops[4]},
	}

	graph := BuildGraph(ops)

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 6)

	// Verify all dependencies satisfied
	for i, op := range sorted {
		deps := op.Dependencies()
		for _, dep := range deps {
			depPos := findOperationIndex(sorted, dep)
			assert.Less(t, depPos, i, "dependency %v should come before %v", dep, op)
		}
	}
}

func TestFindCycle_NoCycle(t *testing.T) {
	opA := domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	opB := &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/b")),
		deps: []domain.Operation{opA},
	}

	graph := BuildGraph([]domain.Operation{opB, opA})

	cycle := graph.FindCycle()

	assert.Nil(t, cycle, "acyclic graph should not have cycles")
}

func TestFindCycle_SelfLoop(t *testing.T) {
	// Operation depends on itself
	baseOp := domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	opA := &mockOperation{
		op:   baseOp,
		deps: nil, // Will set after creation
	}
	// Create self-reference
	opA.deps = []domain.Operation{opA}

	graph := BuildGraph([]domain.Operation{opA})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "self-loop should be detected as cycle")
	require.Len(t, cycle, 1)
	assert.True(t, opA.Equals(cycle[0]))
}

func TestFindCycle_SimpleCycle(t *testing.T) {
	// Create cycle: A -> B -> A
	var opA, opB domain.Operation

	baseA := domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	baseB := domain.NewDirCreate("dir-auto", mustParsePath("/b"))

	opA = &mockOperation{
		op:   baseA,
		deps: []domain.Operation{nil}, // Will set after creating opB
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []domain.Operation{opA},
	}
	// Complete the cycle
	opA.(*mockOperation).deps = []domain.Operation{opB}

	graph := BuildGraph([]domain.Operation{opA, opB})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "cycle A->B->A should be detected")
	assert.GreaterOrEqual(t, len(cycle), 2, "cycle should contain at least 2 operations")
}

func TestFindCycle_LongerCycle(t *testing.T) {
	// Create cycle: A -> B -> C -> A
	var opA, opB, opC domain.Operation

	baseA := domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	baseB := domain.NewDirCreate("dir-auto", mustParsePath("/b"))
	baseC := domain.NewDirCreate("dir-auto", mustParsePath("/c"))

	opA = &mockOperation{
		op:   baseA,
		deps: []domain.Operation{nil}, // Will set after creating opC
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []domain.Operation{opA},
	}
	opC = &mockOperation{
		op:   baseC,
		deps: []domain.Operation{opB},
	}
	// Complete the cycle
	opA.(*mockOperation).deps = []domain.Operation{opC}

	graph := BuildGraph([]domain.Operation{opA, opB, opC})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "cycle A->B->C->A should be detected")
	assert.GreaterOrEqual(t, len(cycle), 3, "cycle should contain at least 3 operations")
}

func TestTopologicalSort_WithCycle(t *testing.T) {
	// Create cycle: A -> B -> A
	var opA, opB domain.Operation

	baseA := domain.NewDirCreate("dir-auto", mustParsePath("/a"))
	baseB := domain.NewDirCreate("dir-auto", mustParsePath("/b"))

	opA = &mockOperation{
		op:   baseA,
		deps: []domain.Operation{nil},
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []domain.Operation{opA},
	}
	opA.(*mockOperation).deps = []domain.Operation{opB}

	graph := BuildGraph([]domain.Operation{opA, opB})

	sorted, err := graph.TopologicalSort()

	assert.Error(t, err, "cyclic graph should return error")
	assert.Nil(t, sorted, "cyclic graph should return nil operations")

	// Verify error type
	var cyclicErr domain.ErrCyclicDependency
	assert.ErrorAs(t, err, &cyclicErr, "error should be ErrCyclicDependency")
}

func TestTopologicalSort_CycleInLargerGraph(t *testing.T) {
	// Graph with acyclic part and cycle: A, B -> C -> D -> C
	opA := domain.NewDirCreate("dir-auto", mustParsePath("/a"))

	var opC, opD domain.Operation
	baseC := domain.NewDirCreate("dir-auto", mustParsePath("/c"))
	baseD := domain.NewDirCreate("dir-auto", mustParsePath("/d"))

	opC = &mockOperation{
		op:   baseC,
		deps: []domain.Operation{nil}, // Will set after creating opD
	}
	opD = &mockOperation{
		op:   baseD,
		deps: []domain.Operation{opC},
	}
	opC.(*mockOperation).deps = []domain.Operation{opD}

	opB := &mockOperation{
		op:   domain.NewDirCreate("dir-auto", mustParsePath("/b")),
		deps: []domain.Operation{opC},
	}

	graph := BuildGraph([]domain.Operation{opA, opB, opC, opD})

	sorted, err := graph.TopologicalSort()

	assert.Error(t, err, "graph with cycle should return error")
	assert.Nil(t, sorted)
}

// findOperationIndex returns the index of an operation in a slice.
// Returns -1 if not found.
func findOperationIndex(ops []domain.Operation, target domain.Operation) int {
	for i, op := range ops {
		if target.Equals(op) {
			return i
		}
	}
	return -1
}
