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

	// Compute level for each operation
	levels := make(map[int][]domain.Operation)
	opLevels := make(map[domain.Operation]int, len(g.ops))

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
			levels[0] = append(levels[0], op)
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
		levels[level] = append(levels[level], op)
		return level
	}

	// Compute levels for all operations
	for _, op := range g.ops {
		computeLevel(op)
	}

	// Convert level map to ordered slice of batches
	if len(levels) == 0 {
		return nil
	}

	// Find max level
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Build batches in order
	batches := make([][]domain.Operation, maxLevel+1)
	for level := 0; level <= maxLevel; level++ {
		if ops, exists := levels[level]; exists {
			batches[level] = ops
		}
	}

	return batches
}
