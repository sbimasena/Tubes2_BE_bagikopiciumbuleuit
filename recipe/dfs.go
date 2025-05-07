package recipe

import (
	"sync"
)

// Edge represents a directed edge in the visualization graph
type Edge struct {
	From string
	To   string
}

// Global visualization state
var (
	// Controls whether visualization is enabled
	VisualEnabled = false

	// Frame counter and mutex
	frameCount int = 0
	frameMutex sync.Mutex

	// Current edges in the graph
	currentEdges []Edge

	// Current visited nodes
	currentVisited map[string]bool = make(map[string]bool)

	// Mutex to protect visual state
	visualMutex sync.Mutex
)

// Optimized DFS with visual tracing
func dfsOptimized(
	target string,
	elements map[string][][]string,
	basicElements map[string]bool,
	path []string,
	visited map[string]bool,
	maxRecipes int,
	stopEarly bool,
	result *[][]string,
	enableVisual bool, // Flag to enable/disable visualization
) {
	// Visited check to prevent cycles
	if visited[target] {
		return
	}

	// Early stop conditions
	if stopEarly && len(*result) > 0 {
		return
	}
	if !stopEarly && len(*result) >= maxRecipes {
		return
	}

	// Mark as visited
	visited[target] = true

	// Update visualization if enabled
	if enableVisual {
		// Add edge from parent to this node
		if len(path) > 1 {
			parent := path[len(path)-2]
			// We want parent -> child direction for visualization
			AddEdge(parent, target)
		}

		// Mark this node as currently being visited
		MarkVisited(target, true)

		// Capture the current state
		CaptureFrame(elements, basicElements)
	}

	// Check if we've reached a basic element
	if basicElements[target] {
		complete := append([]string{}, path...)
		*result = append(*result, complete)

		// Clear visited mark before returning
		if enableVisual {
			MarkVisited(target, false)
			CaptureFrame(elements, basicElements)
		}

		visited[target] = false
		return
	}

	// Get recipes for this target
	recipes, ok := elements[target]
	if !ok {
		// Clear visited mark before returning
		if enableVisual {
			MarkVisited(target, false)
			CaptureFrame(elements, basicElements)
		}

		visited[target] = false
		return
	}

	// Try each recipe
	for _, recipe := range recipes {
		success := true
		tempPath := append([]string{}, path...)

		// Try each ingredient in the recipe
		for _, ing := range recipe {
			before := len(*result)

			// Recursively find path for this ingredient
			dfsOptimized(ing, elements, basicElements, append(tempPath, ing), visited, 1, true, result, enableVisual)

			// Check if we found a path
			if len(*result) == before {
				success = false
				break
			}

			// Update our path with the successful subpath
			tempPath = (*result)[len(*result)-1]
		}

		// If we successfully built a path for this recipe
		if success {
			if !stopEarly {
				*result = append(*result, tempPath)

				// Capture state after finding a complete recipe
				if enableVisual {
					CaptureFrame(elements, basicElements)
				}
			}

			// Stop if we have enough recipes or stopEarly is set
			if stopEarly || len(*result) >= maxRecipes {
				// Clear visited mark before returning
				if enableVisual {
					MarkVisited(target, false)
					CaptureFrame(elements, basicElements)
				}

				visited[target] = false
				return
			}
		}
	}

	// Clear visited mark before backtracking
	if enableVisual {
		MarkVisited(target, false)
		CaptureFrame(elements, basicElements)
	}

	visited[target] = false
}
