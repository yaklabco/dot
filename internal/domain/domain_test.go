package domain_test

import (
	"testing"

	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNodeType_String(t *testing.T) {
	tests := []struct {
		name     string
		nodeType domain.NodeType
		expected string
	}{
		{
			name:     "File",
			nodeType: domain.NodeFile,
			expected: "File",
		},
		{
			name:     "Dir",
			nodeType: domain.NodeDir,
			expected: "Dir",
		},
		{
			name:     "Symlink",
			nodeType: domain.NodeSymlink,
			expected: "Symlink",
		},
		{
			name:     "Unknown",
			nodeType: domain.NodeType(99),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.nodeType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNode_IsFile(t *testing.T) {
	tests := []struct {
		name     string
		node     domain.Node
		expected bool
	}{
		{
			name: "File returns true",
			node: domain.Node{
				Path: domain.NewFilePath("/test/file.txt").Unwrap(),
				Type: domain.NodeFile,
			},
			expected: true,
		},
		{
			name: "Dir returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/dir").Unwrap(),
				Type: domain.NodeDir,
			},
			expected: false,
		},
		{
			name: "Symlink returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/link").Unwrap(),
				Type: domain.NodeSymlink,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.IsFile()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNode_IsDir(t *testing.T) {
	tests := []struct {
		name     string
		node     domain.Node
		expected bool
	}{
		{
			name: "Dir returns true",
			node: domain.Node{
				Path: domain.NewFilePath("/test/dir").Unwrap(),
				Type: domain.NodeDir,
			},
			expected: true,
		},
		{
			name: "File returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/file.txt").Unwrap(),
				Type: domain.NodeFile,
			},
			expected: false,
		},
		{
			name: "Symlink returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/link").Unwrap(),
				Type: domain.NodeSymlink,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.IsDir()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNode_IsSymlink(t *testing.T) {
	tests := []struct {
		name     string
		node     domain.Node
		expected bool
	}{
		{
			name: "Symlink returns true",
			node: domain.Node{
				Path: domain.NewFilePath("/test/link").Unwrap(),
				Type: domain.NodeSymlink,
			},
			expected: true,
		},
		{
			name: "File returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/file.txt").Unwrap(),
				Type: domain.NodeFile,
			},
			expected: false,
		},
		{
			name: "Dir returns false",
			node: domain.Node{
				Path: domain.NewFilePath("/test/dir").Unwrap(),
				Type: domain.NodeDir,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.IsSymlink()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustParsePath(t *testing.T) {
	t.Run("Valid file path", func(t *testing.T) {
		path := domain.MustParsePath("/home/user/.vimrc")
		assert.NotNil(t, path)
		assert.Equal(t, "/home/user/.vimrc", path.String())
	})

	t.Run("Panics on invalid path", func(t *testing.T) {
		assert.Panics(t, func() {
			domain.MustParsePath("")
		})
	})
}

func TestMustParseTargetPath(t *testing.T) {
	t.Run("Valid target path", func(t *testing.T) {
		path := domain.MustParseTargetPath("/home/user/.vimrc")
		assert.NotNil(t, path)
		assert.Equal(t, "/home/user/.vimrc", path.String())
	})

	t.Run("Panics on invalid path", func(t *testing.T) {
		assert.Panics(t, func() {
			domain.MustParseTargetPath("")
		})
	})
}

func TestPlan_Validate(t *testing.T) {
	t.Run("Valid plan", func(t *testing.T) {
		plan := domain.Plan{
			Operations: []domain.Operation{},
		}
		err := plan.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid operation with empty ID", func(t *testing.T) {
		// Create invalid operation
		op := domain.LinkCreate{
			OpID: "", // Invalid: empty ID
		}
		plan := domain.Plan{
			Operations: []domain.Operation{op},
		}
		err := plan.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation ID cannot be empty")
	})
}

func TestPlan_CanParallelize(t *testing.T) {
	tests := []struct {
		name     string
		plan     domain.Plan
		expected bool
	}{
		{
			name:     "Plan with batches can parallelize",
			plan:     domain.Plan{Batches: [][]domain.Operation{{}}},
			expected: true,
		},
		{
			name:     "Plan without batches cannot parallelize",
			plan:     domain.Plan{Batches: nil},
			expected: false,
		},
		{
			name:     "Plan with empty batches cannot parallelize",
			plan:     domain.Plan{Batches: [][]domain.Operation{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.CanParallelize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlan_ParallelBatches(t *testing.T) {
	batches := [][]domain.Operation{{}}
	plan := domain.Plan{Batches: batches}

	result := plan.ParallelBatches()
	assert.Equal(t, batches, result)
}

func TestPlan_PackageNames(t *testing.T) {
	tests := []struct {
		name     string
		plan     domain.Plan
		expected []string
	}{
		{
			name: "Plan with packages",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"pkg1": {"op1"},
					"pkg2": {"op2"},
				},
			},
			expected: []string{"pkg1", "pkg2"},
		},
		{
			name:     "Plan without PackageOperations",
			plan:     domain.Plan{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.PackageNames()
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestPlan_HasPackage(t *testing.T) {
	tests := []struct {
		name     string
		plan     domain.Plan
		pkg      string
		expected bool
	}{
		{
			name: "Package exists",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"vim": {"op1"},
				},
			},
			pkg:      "vim",
			expected: true,
		},
		{
			name: "Package does not exist",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"vim": {"op1"},
				},
			},
			pkg:      "emacs",
			expected: false,
		},
		{
			name:     "Nil PackageOperations",
			plan:     domain.Plan{},
			pkg:      "vim",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.HasPackage(tt.pkg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlan_OperationCountForPackage(t *testing.T) {
	tests := []struct {
		name     string
		plan     domain.Plan
		pkg      string
		expected int
	}{
		{
			name: "Package with operations",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"vim": {"op1", "op2", "op3"},
				},
			},
			pkg:      "vim",
			expected: 3,
		},
		{
			name: "Package with no operations",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"vim": {},
				},
			},
			pkg:      "vim",
			expected: 0,
		},
		{
			name: "Package not in plan",
			plan: domain.Plan{
				PackageOperations: map[string][]domain.OperationID{
					"vim": {"op1"},
				},
			},
			pkg:      "emacs",
			expected: 0,
		},
		{
			name:     "Nil PackageOperations",
			plan:     domain.Plan{},
			pkg:      "vim",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.OperationCountForPackage(tt.pkg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
