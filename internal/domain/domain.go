package domain

// Package represents a collection of configuration files to be managed.
type Package struct {
	Name string
	Path PackagePath
	Tree *Node // Optional: file tree for the package
}

// NodeType identifies the type of filesystem node.
type NodeType int

const (
	// NodeFile represents a regular file.
	NodeFile NodeType = iota

	// NodeDir represents a directory.
	NodeDir

	// NodeSymlink represents a symbolic link.
	NodeSymlink
)

// String returns the string representation of a NodeType.
func (t NodeType) String() string {
	switch t {
	case NodeFile:
		return "File"
	case NodeDir:
		return "Dir"
	case NodeSymlink:
		return "Symlink"
	default:
		return "Unknown"
	}
}

// Node represents a node in a filesystem tree.
type Node struct {
	Path     FilePath
	Type     NodeType
	Children []Node
}

// IsFile returns true if the node is a file.
func (n Node) IsFile() bool {
	return n.Type == NodeFile
}

// IsDir returns true if the node is a directory.
func (n Node) IsDir() bool {
	return n.Type == NodeDir
}

// IsSymlink returns true if the node is a symbolic link.
func (n Node) IsSymlink() bool {
	return n.Type == NodeSymlink
}

// Plan represents a set of operations to execute.
type Plan struct {
	Operations []Operation
	Metadata   PlanMetadata
	Batches    [][]Operation // Parallel execution batches (if computed)

	// PackageOperations maps package names to operation IDs that belong to that package.
	// This enables tracking which operations were generated for which packages,
	// allowing accurate manifest updates and selective operations.
	// Optional field for backward compatibility.
	PackageOperations map[string][]OperationID `json:"package_operations,omitempty"`

	// PackageSkippedLinks maps package names to absolute target paths of links
	// that already exist on disk pointing at the correct source. They generate
	// no operations but are part of the managed state and must be recorded in
	// the manifest alongside newly created links.
	PackageSkippedLinks map[string][]string `json:"package_skipped_links,omitempty"`
}

// SkippedLinksForPackage returns the already-correct link target paths for the
// specified package. Returns nil when none were recorded.
func (p Plan) SkippedLinksForPackage(pkg string) []string {
	if p.PackageSkippedLinks == nil {
		return nil
	}
	return p.PackageSkippedLinks[pkg]
}

// Validate checks if the plan is valid.
func (p Plan) Validate() error {
	// Validate each operation
	for _, op := range p.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CanParallelize returns true if the plan has computed parallel batches.
func (p Plan) CanParallelize() bool {
	return len(p.Batches) > 0
}

// ParallelBatches returns the parallel execution batches.
// Returns nil if parallelization has not been computed.
func (p Plan) ParallelBatches() [][]Operation {
	return p.Batches
}

// OperationsForPackage returns all operations that belong to the specified package.
// Returns an empty slice if the package is not in the plan or if PackageOperations is not set.
func (p Plan) OperationsForPackage(pkg string) []Operation {
	if p.PackageOperations == nil {
		return []Operation{}
	}

	ids := p.PackageOperations[pkg]
	if len(ids) == 0 {
		return []Operation{}
	}

	result := make([]Operation, 0, len(ids))
	for _, op := range p.Operations {
		for _, id := range ids {
			if op.ID() == id {
				result = append(result, op)
				break
			}
		}
	}

	return result
}

// PackageNames returns a list of all package names in the plan.
// Returns an empty slice if PackageOperations is not set.
func (p Plan) PackageNames() []string {
	if p.PackageOperations == nil {
		return []string{}
	}

	names := make([]string, 0, len(p.PackageOperations))
	for name := range p.PackageOperations {
		names = append(names, name)
	}

	return names
}

// HasPackage returns true if the plan contains operations for the specified package.
func (p Plan) HasPackage(pkg string) bool {
	if p.PackageOperations == nil {
		return false
	}

	_, exists := p.PackageOperations[pkg]
	return exists
}

// OperationCountForPackage returns the number of operations for the specified package.
// Returns 0 if the package is not in the plan or if PackageOperations is not set.
func (p Plan) OperationCountForPackage(pkg string) int {
	if p.PackageOperations == nil {
		return 0
	}

	return len(p.PackageOperations[pkg])
}

// PlanMetadata contains statistics and diagnostic information about a plan.
type PlanMetadata struct {
	PackageCount   int            `json:"package_count"`
	OperationCount int            `json:"operation_count"`
	LinkCount      int            `json:"link_count"`
	DirCount       int            `json:"dir_count"`
	Conflicts      []ConflictInfo `json:"conflicts,omitempty"`
	Warnings       []WarningInfo  `json:"warnings,omitempty"`
}
