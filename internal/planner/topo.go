package planner

import "github.com/yaklabco/dot/internal/domain"

// TopologicalSort returns operations in dependency order using depth-first search.
// Operations with no dependencies come first, and each operation appears after
// all its dependencies.
//
// Returns an error if the graph contains cycles, as cyclic dependencies cannot
// be topologically sorted.
//
// Time complexity: O(n + e) where n is the number of operations and e is
// the number of dependency edges.
func (g *DependencyGraph) TopologicalSort() ([]domain.Operation, error) {
	// First check for cycles
	if cycle := g.FindCycle(); cycle != nil {
		return nil, domain.ErrCyclicDependency{Cycle: formatCycle(cycle)}
	}

	visited := make(map[domain.Operation]bool, len(g.ops))
	var result []domain.Operation

	// Visit function performs DFS post-order traversal
	var visit func(domain.Operation) error
	visit = func(op domain.Operation) error {
		if visited[op] {
			return nil
		}

		// Visit all dependencies first (use graph's dependencies, not operation's)
		deps := g.Dependencies(op)
		for _, dep := range deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		// Mark as visited and add to result
		visited[op] = true
		result = append(result, op)
		return nil
	}

	// Visit all operations
	for _, op := range g.ops {
		if !visited[op] {
			if err := visit(op); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// FindCycle detects circular dependencies using depth-first search.
// Returns a slice of operations forming a cycle, or nil if no cycle exists.
//
// The returned cycle starts and ends with the same operation, showing the
// circular path: [A, B, C, A] indicates A depends on B, B depends on C,
// and C depends on A.
//
// Time complexity: O(n + e) where n is the number of operations and e is
// the number of dependency edges.
func (g *DependencyGraph) FindCycle() []domain.Operation {
	visited := make(map[domain.Operation]bool, len(g.ops))
	recStack := make(map[domain.Operation]bool, len(g.ops))
	parent := make(map[domain.Operation]domain.Operation, len(g.ops))

	// DFS function to detect back edges (cycles)
	var dfs func(domain.Operation) []domain.Operation
	dfs = func(op domain.Operation) []domain.Operation {
		visited[op] = true
		recStack[op] = true

		// Use graph's dependencies, not operation's
		deps := g.Dependencies(op)
		for _, dep := range deps {
			if !visited[dep] {
				parent[dep] = op
				if cycle := dfs(dep); cycle != nil {
					return cycle
				}
			} else if recStack[dep] {
				// Back edge found - reconstruct cycle
				return reconstructCycle(op, dep, parent)
			}
		}

		recStack[op] = false
		return nil
	}

	// Check all operations
	for _, op := range g.ops {
		if !visited[op] {
			if cycle := dfs(op); cycle != nil {
				return cycle
			}
		}
	}

	return nil
}

// reconstructCycle builds the cycle path from current node back to the cycle start.
// current is the node where we detected the back edge
// cycleStart is the node we're going back to (the beginning of the cycle)
// parent maps each node to its parent in the DFS tree
func reconstructCycle(current, cycleStart domain.Operation, parent map[domain.Operation]domain.Operation) []domain.Operation {
	// Handle self-loop case
	if current.Equals(cycleStart) {
		return []domain.Operation{cycleStart}
	}

	cycle := []domain.Operation{cycleStart}

	// Walk backwards from current to cycleStart using parent pointers
	node := current
	for !node.Equals(cycleStart) {
		cycle = append(cycle, node)
		nextNode, exists := parent[node]
		if !exists {
			// Should not happen in valid graph, but handle gracefully
			break
		}
		node = nextNode
	}

	// Add cycle start at the end to show complete cycle
	cycle = append(cycle, cycleStart)

	// Reverse to show forward path
	for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
		cycle[i], cycle[j] = cycle[j], cycle[i]
	}

	return cycle
}

// formatCycle converts a cycle of operations into string descriptions.
func formatCycle(cycle []domain.Operation) []string {
	if len(cycle) == 0 {
		return nil
	}

	result := make([]string, len(cycle))
	for i, op := range cycle {
		result[i] = op.String()
	}
	return result
}
