package planner

import "github.com/yaklabco/dot/internal/domain"

// ParallelizationPlan computes batches of operations that can execute concurrently.
// Returns a slice of batches where operations within each batch have no dependencies
// on each other and can run in parallel. Batches must execute sequentially - batch N
// must complete before batch N+1 begins.
//
// The algorithm uses level-based grouping: operations with no dependencies are at
// level 0, operations depending only on level 0 are at level 1, etc. Operations
// at the same level can execute in parallel.
//
// Time complexity: O(n + e) where n is the number of operations and e is the number
// of dependency edges.
func (g *DependencyGraph) ParallelizationPlan() [][]domain.Operation {
	if len(g.ops) == 0 {
		return nil
	}

	// Compute level for each operation using a slice for contiguous levels
	var levels [][]domain.Operation
	opLevels := make(map[domain.Operation]int, len(g.ops))

	// appendToLevel adds an operation to the specified level, growing the slice if needed
	appendToLevel := func(level int, op domain.Operation) {
		// Grow slice if needed
		for len(levels) <= level {
			levels = append(levels, nil)
		}
		levels[level] = append(levels[level], op)
	}

	// computeLevel recursively computes the level of an operation with memoization
	var computeLevel func(domain.Operation) int
	computeLevel = func(op domain.Operation) int {
		// Return cached result if already computed
		if level, exists := opLevels[op]; exists {
			return level
		}

		// Get dependencies from graph, not operation
		deps := g.Dependencies(op)
		if len(deps) == 0 {
			// No dependencies = level 0
			opLevels[op] = 0
			appendToLevel(0, op)
			return 0
		}

		// Level = max(dependency levels) + 1
		maxDepLevel := -1
		for _, dep := range deps {
			depLevel := computeLevel(dep)
			if depLevel > maxDepLevel {
				maxDepLevel = depLevel
			}
		}

		level := maxDepLevel + 1
		opLevels[op] = level
		appendToLevel(level, op)
		return level
	}

	// Compute levels for all operations
	for _, op := range g.ops {
		computeLevel(op)
	}

	return levels
}
