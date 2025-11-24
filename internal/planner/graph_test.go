package planner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/dot/internal/domain"
)

func TestBuildGraph_Empty(t *testing.T) {
	ops := []domain.Operation{}

	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 0, graph.Size(), "empty operation list should produce empty graph")
}

func TestBuildGraph_SingleOperation(t *testing.T) {
	source := mustParsePath("/packages/package/file")
	target := mustParseTargetPath("/home/user/.config/file")
	op := domain.NewLinkCreate("link1", source, target)

	ops := []domain.Operation{op}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 1, graph.Size(), "single operation should produce graph with one node")
	assert.True(t, graph.HasOperation(op), "graph should contain the operation")
}

func TestBuildGraph_IndependentOperations(t *testing.T) {
	op1 := domain.NewLinkCreate(
		"link1",
		mustParsePath("/packages/pkg/file1"),
		mustParseTargetPath("/home/user/file1"),
	)
	op2 := domain.NewLinkCreate(
		"link2",
		mustParsePath("/packages/pkg/file2"),
		mustParseTargetPath("/home/user/file2"),
	)

	ops := []domain.Operation{op1, op2}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 2, graph.Size())
	assert.True(t, graph.HasOperation(op1))
	assert.True(t, graph.HasOperation(op2))

	// No dependencies between operations
	deps1 := graph.Dependencies(op1)
	deps2 := graph.Dependencies(op2)
	assert.Empty(t, deps1, "independent operations should have no dependencies")
	assert.Empty(t, deps2, "independent operations should have no dependencies")
}

func TestBuildGraph_LinearDependencies(t *testing.T) {
	// Create a linear dependency chain: dirCreate -> linkCreate
	dirPath := mustParsePath("/home/user/.config")
	dirOp := domain.NewDirCreate("dir1", dirPath)

	linkOp := domain.NewLinkCreate(
		"link1",
		mustParsePath("/packages/pkg/config"),
		mustParseTargetPath("/home/user/.config/app.conf"),
	)

	// Mock linkOp to depend on dirOp
	linkOpWithDep := &mockOperation{
		op:   linkOp,
		deps: []domain.Operation{dirOp},
	}

	ops := []domain.Operation{linkOpWithDep, dirOp}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 2, graph.Size())

	// linkOp should depend on dirOp
	deps := graph.Dependencies(linkOpWithDep)
	require.Len(t, deps, 1)
	assert.True(t, dirOp.Equals(deps[0]), "link should depend on directory creation")

	// dirOp should have no dependencies
	dirDeps := graph.Dependencies(dirOp)
	assert.Empty(t, dirDeps)
}

func TestBuildGraph_DiamondPattern(t *testing.T) {
	// Diamond dependency: A -> B, A -> C, B -> D, C -> D
	opA := domain.NewDirCreate("dir1", mustParsePath("/home/user/.config"))
	opB := domain.NewDirCreate("dir2", mustParsePath("/home/user/.config/app1"))
	opC := domain.NewDirCreate("dir3", mustParsePath("/home/user/.config/app2"))
	opD := domain.NewLinkCreate(
		"link1",
		mustParsePath("/packages/pkg/file"),
		mustParseTargetPath("/home/user/.config/file"),
	)

	// Create mock operations with dependencies
	opBWithDep := &mockOperation{op: opB, deps: []domain.Operation{opA}}
	opCWithDep := &mockOperation{op: opC, deps: []domain.Operation{opA}}
	opDWithDep := &mockOperation{op: opD, deps: []domain.Operation{opBWithDep, opCWithDep}}

	ops := []domain.Operation{opDWithDep, opCWithDep, opBWithDep, opA}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 4, graph.Size())

	// Verify dependencies
	assert.Empty(t, graph.Dependencies(opA))
	assert.Len(t, graph.Dependencies(opBWithDep), 1)
	assert.Len(t, graph.Dependencies(opCWithDep), 1)
	assert.Len(t, graph.Dependencies(opDWithDep), 2)
}

func TestGraph_Size(t *testing.T) {
	tests := []struct {
		name     string
		ops      []domain.Operation
		expected int
	}{
		{
			name:     "empty graph",
			ops:      []domain.Operation{},
			expected: 0,
		},
		{
			name: "single operation",
			ops: []domain.Operation{
				domain.NewLinkCreate("link1", mustParsePath("/a"), mustParseTargetPath("/b")),
			},
			expected: 1,
		},
		{
			name: "multiple operations",
			ops: []domain.Operation{
				domain.NewLinkCreate("link1", mustParsePath("/a"), mustParseTargetPath("/b")),
				domain.NewLinkCreate("link2", mustParsePath("/c"), mustParseTargetPath("/d")),
				domain.NewDirCreate("dir1", mustParsePath("/e")),
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := BuildGraph(tt.ops)
			assert.Equal(t, tt.expected, graph.Size())
		})
	}
}

func TestGraph_HasOperation(t *testing.T) {
	op1 := domain.NewLinkCreate("link1", mustParsePath("/a"), mustParseTargetPath("/b"))
	op2 := domain.NewLinkCreate("link2", mustParsePath("/c"), mustParseTargetPath("/d"))
	op3 := domain.NewDirCreate("dir1", mustParsePath("/e"))

	graph := BuildGraph([]domain.Operation{op1, op2})

	assert.True(t, graph.HasOperation(op1))
	assert.True(t, graph.HasOperation(op2))
	assert.False(t, graph.HasOperation(op3), "operation not in graph should return false")
}

// mockOperation wraps an operation with custom dependencies for testing
type mockOperation struct {
	op   domain.Operation
	deps []domain.Operation
}

func (m *mockOperation) ID() domain.OperationID {
	return m.op.ID()
}

func (m *mockOperation) Kind() domain.OperationKind {
	return m.op.Kind()
}

func (m *mockOperation) Validate() error {
	return m.op.Validate()
}

func (m *mockOperation) Dependencies() []domain.Operation {
	return m.deps
}

func (m *mockOperation) Execute(ctx context.Context, fs domain.FS) error {
	return m.op.Execute(ctx, fs)
}

func (m *mockOperation) Rollback(ctx context.Context, fs domain.FS) error {
	return m.op.Rollback(ctx, fs)
}

func (m *mockOperation) String() string {
	return m.op.String()
}

func (m *mockOperation) Equals(other domain.Operation) bool {
	if otherMock, ok := other.(*mockOperation); ok {
		return m.op.Equals(otherMock.op)
	}
	return m.op.Equals(other)
}

// mustParsePath creates a FilePath or panics (for test convenience)
func mustParsePath(s string) domain.FilePath {
	result := domain.NewFilePath(s)
	if !result.IsOk() {
		panic(result.UnwrapErr())
	}
	return result.Unwrap()
}

func mustParseTargetPath(s string) domain.TargetPath {
	result := domain.NewTargetPath(s)
	if !result.IsOk() {
		panic(result.UnwrapErr())
	}
	return result.Unwrap()
}
