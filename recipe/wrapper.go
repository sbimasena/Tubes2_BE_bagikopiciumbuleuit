package recipe

import (
	"fmt"
	"sync"
)

// FindShortestRecipe finds the shortest recipe path to a target element
// Uses non-concurrent DFS as specified
func FindShortestRecipe(target string, elements map[string][][]string, basicElements map[string]bool) []string {
	fmt.Println("Finding shortest recipe for", target)

	visited := map[string]bool{}
	result := [][]string{}

	// Use the optimized DFS algorithm - pass result by reference
	dfsOptimized(target, elements, basicElements, []string{target}, visited, 1, true, &result, true)

	if len(result) == 0 {
		fmt.Println("No recipe found")
		return nil
	}
	// Execute visualization synchronously to ensure it completes
	TraceLive(result[0], elements, basicElements)

	return result[0]
}

// FindMultipleRecipesConcurrent finds multiple recipes for a target element using concurrency
func FindMultipleRecipesConcurrent(target string, elements map[string][][]string, basicElements map[string]bool, maxRecipes int) [][]string {
	fmt.Printf("Finding up to %d recipes for %s\n", maxRecipes, target)

	var wg sync.WaitGroup
	var mu sync.Mutex

	result := [][]string{}

	// Get recipes for the target
	recipes, ok := elements[target]
	if !ok {
		return nil
	}

	// Use a buffered channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, 8)

	for _, recipe := range recipes {
		semaphore <- struct{}{} // Acquire token
		wg.Add(1)

		go func(recipe []string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release token

			tempPath := []string{target}
			success := true

			// Process each ingredient in the recipe
			for _, ing := range recipe {
				subVisited := map[string]bool{}
				subResult := [][]string{}

				// Find a path for this ingredient
				dfsOptimized(ing, elements, basicElements, append(tempPath, ing), subVisited, 1, true, &subResult, true)

				// If no valid paths for this ingredient, recipe fails
				if len(subResult) == 0 {
					success = false
					break
				}

				// Update our path with successful subpath
				tempPath = subResult[0]
			}

			// If we successfully built a path for this recipe
			if success {
				mu.Lock()
				if len(result) < maxRecipes {
					result = append(result, tempPath)
				}
				mu.Unlock()
			}
		}(recipe)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Visualize synchronously if we have results
	if len(result) > 0 {
		TraceLive(result[0], elements, basicElements)
	}

	return result
}
