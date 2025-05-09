package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// New function to find multiple recipes concurrently
func FindMultipleRecipesConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding up to %d paths to create: %s\n", maxRecipes, targetElement)

	elementMap := make(map[string]ElementRecipe)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
	}

	type RecipeCombo struct {
		ingredients [2]string
		result      string
	}

	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		fmt.Printf("Target element '%s' not found in recipes\n", targetElement)
		return
	}

	var possibleCombinations []RecipeCombo
	for _, combo := range targetRecipe.Recipes {
		possibleCombinations = append(possibleCombinations, RecipeCombo{
			ingredients: combo,
			result:      targetElement,
		})
	}

	if len(possibleCombinations) == 0 {
		fmt.Printf("No recipes found for element '%s'\n", targetElement)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 10)
	var allPaths []Path

	for _, combo := range possibleCombinations {
		mu.Lock()
		if len(allPaths) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		sem <- struct{}{}
		wg.Add(1)

		go func(combo RecipeCombo) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			var combinedPaths []Path
			for _, ingredient := range combo.ingredients {
				isBasic := false
				for _, basic := range startingElements {
					if basic == ingredient {
						isBasic = true
						break
					}
				}
				if isBasic {
					continue
				}

				ingredientPaths := findPathDFS(recipes, startingElements, ingredient)
				if len(ingredientPaths) > 0 {
					combinedPaths = append(combinedPaths, ingredientPaths[0])
				} else {
					return
				}
			}

			// Tier validation for final combination
			a, b := combo.ingredients[0], combo.ingredients[1]
			aRecipe, aOk := elementMap[a]
			bRecipe, bOk := elementMap[b]
			if !aOk || !bOk {
				return
			}
			maxTier := aRecipe.Tier
			if bRecipe.Tier > maxTier {
				maxTier = bRecipe.Tier
			}
			if targetRecipe.Tier <= maxTier {
				return
			}

			// Deduplicate steps
			stepSet := make(map[[3]string]bool)
			var steps []Step

			for _, path := range combinedPaths {
				for _, s := range path.Steps {
					k := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
					if !stepSet[k] {
						stepSet[k] = true
						steps = append(steps, s)
					}
				}
			}

			// Add the final combination step
			finalStep := Step{Ingredients: combo.ingredients, Result: combo.result}
			k := [3]string{combo.ingredients[0], combo.ingredients[1], combo.result}
			if !stepSet[k] {
				steps = append(steps, finalStep)
			}

			finalPath := Path{
				Steps:     steps,
				FinalItem: targetElement,
			}

			mu.Lock()
			if len(allPaths) < maxRecipes {
				allPaths = append(allPaths, finalPath)
				if len(allPaths) >= maxRecipes {
					cancel()
				}
			}
			mu.Unlock()
		}(combo)
	}

	wg.Wait()

	if len(allPaths) == 0 {
		fmt.Println("No valid paths found")
		return
	}

	if len(allPaths) < maxRecipes {
		fmt.Printf("Only found %d valid path(s) (requested %d):\n", len(allPaths), maxRecipes)
	} else {
		fmt.Printf("Found %d different paths to create %s:\n", len(allPaths), targetElement)
	}

	for i, path := range allPaths {
		fmt.Printf("\nRecipe %d with %d steps:\n", i+1, len(path.Steps))
		for j, step := range path.Steps {
			fmt.Printf("%d. %s + %s = %s\n", j+1, step.Ingredients[0], step.Ingredients[1], step.Result)
		}
	}
}

// Function to find a single recipe
func FindSingleRecipe(recipesFile, targetElement string, startingElements []string) {
	FindPathToElement(recipesFile, targetElement, startingElements)
}
