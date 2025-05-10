package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type ElementData struct {
	Element  string     `json:"element"`
	ImageURL string     `json:"image_url"`
	Recipes  [][]string `json:"recipes"`
	Tier     int        `json:"tier"`
}

type SearchResult struct {
	Paths        [][]string            `json:"paths"`
	Steps        []map[string][]string `json:"steps"`
	NodesVisited int                   `json:"nodes_visited"`
	Duration     string                `json:"duration"`
	Algorithm    string                `json:"algorithm"`
}

func FindSingleRecipe(
	target string,
	elements map[string][][]string,
	basicElements map[string]bool,
	algorithm string,
	tierMap map[string]int,
) ([]string, map[string][]string, int, time.Duration) {
	switch algorithm {
	case "bfs":
		p, s, n, t := BiSearchBFS(target, elements, basicElements, tierMap)
		return p, s, n, t
	case "dfs":
		p, s, n, t := BiSearchDFS(target, elements, basicElements, tierMap)
		return p, s, n, t
	default:
		return nil, nil, 0, 0
	}
}

func FindMultipleRecipes(target string, elements map[string][][]string, basicElements map[string]bool, algorithm string, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	switch algorithm {
	case "bfs":
		return BiSearchMultipleBFS(target, elements, basicElements, maxPaths, tierMap)
	case "dfs":
		return BiSearchMultipleDFS(target, elements, basicElements, maxPaths, tierMap)
	default:
		return nil, nil, 0
	}
}

func LoadElements(filename string) ([]ElementData, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var elements []ElementData
	err = json.Unmarshal(data, &elements)
	if err != nil {
		return nil, err
	}
	return elements, nil
}

func PrepareElementMaps(elements []ElementData) (map[string][][]string, map[string]int, map[string]bool) {
	recipeMap := make(map[string][][]string)
	tierMap := make(map[string]int)
	basicElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}
	for _, elem := range elements {
		recipeMap[elem.Element] = elem.Recipes
		tierMap[elem.Element] = elem.Tier
	}
	for elem := range basicElements {
		tierMap[elem] = 0
	}
	return recipeMap, tierMap, basicElements
}

func BiSearchMultipleBFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	var (
		paths          [][]string
		allSteps       []map[string][]string
		totalNodes     int
		startTime      = time.Now()
		pathSignatures = map[string]bool{}
	)

	fmt.Println("Finding up to", maxPaths, "different paths for", target)

	for attempt := range maxPaths * 5 {
		if len(paths) >= maxPaths {
			break
		}

		// Copy and optionally shuffle elements
		elementsCopy := copyElements(elements)
		shuffleRecipes(elementsCopy, attempt)

		var p []string
		var s map[string][]string
		var n int

		done := make(chan bool, 1)

		go func() {
			p, s, n, _ = BiSearchBFS(target, elementsCopy, basicElements, tierMap)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			fmt.Printf("Search attempt %d timed out\n", attempt)
			continue
		}

		if p == nil {
			continue
		}

		sig := fmt.Sprintf("%v", p)
		if pathSignatures[sig] {
			continue
		}

		pathSignatures[sig] = true
		paths = append(paths, p)
		allSteps = append(allSteps, s)
		totalNodes += n

		fmt.Printf("Path #%d found (length: %d)\n", len(paths), len(p))
	}

	duration := time.Since(startTime)

	fmt.Printf("\n📦 Total nodes explored: %d\n", totalNodes)
	fmt.Printf("⏱ Total execution time: %v\n", duration)

	if len(paths) < maxPaths {
		fmt.Printf("Only found %d different paths (of %d requested)\n", len(paths), maxPaths)
	} else {
		fmt.Printf("Found %d different paths\n", len(paths))
	}

	return paths, allSteps, totalNodes
}

func BiSearchMultipleDFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	// Prepare results
	var paths [][]string
	var allSteps []map[string][]string
	totalNodes := 0
	startTime := time.Now()

	// Track found paths
	pathSignatures := make(map[string]bool)

	// Copy elements once at the beginning
	originalElements := copyElements(elements)

	fmt.Println("Finding up to", maxPaths, "different paths for", target)

	// Run searches sequentially to ensure absolute purity
	for attempt := range maxPaths * 3 {
		if len(paths) >= maxPaths {
			break
		}

		// Use pure bidirectional DFS with timeout
		var p []string
		var s map[string][]string
		var n int

		// Channel for search completion
		done := make(chan bool)

		go func() {
			// Use original elements without any modifications
			p, s, n, _ = BiSearchDFS(target, originalElements, basicElements, tierMap)
			done <- true
		}()

		// Wait for result with timeout
		select {
		case <-done:
			// Continue processing
		case <-time.After(3 * time.Second):
			fmt.Printf("Search attempt %d timed out\n", attempt)
			continue
		}

		if p == nil {
			continue // No path found
		}

		// Generate a signature for this path
		signature := fmt.Sprintf("%v", p) // Simple signature based on path elements

		// Check if this is a unique path
		if !pathSignatures[signature] {
			// Add to results
			pathSignatures[signature] = true
			paths = append(paths, p)
			allSteps = append(allSteps, s)
			totalNodes += n

			fmt.Printf("Path #%d found (length: %d)\n", len(paths), len(p))
		}
	}

	// Calculate total duration
	duration := time.Since(startTime)

	// Print results
	fmt.Printf("\n📦 Total nodes explored: %d\n", totalNodes)
	fmt.Printf("⏱ Total execution time: %v\n", duration)

	if len(paths) < maxPaths {
		fmt.Printf("Only found %d different paths (of %d requested)\n", len(paths), maxPaths)
	} else {
		fmt.Printf("Found %d different paths\n", len(paths))
	}

	return paths, allSteps, totalNodes
}

// Fungsi untuk menyalin struktur data elements
func copyElements(elements map[string][][]string) map[string][][]string {
	result := make(map[string][][]string)

	for elem, recipes := range elements {
		// Salin recipes
		recipesCopy := make([][]string, len(recipes))
		for i, recipe := range recipes {
			recipeCopy := make([]string, len(recipe))
			copy(recipeCopy, recipe)
			recipesCopy[i] = recipeCopy
		}

		result[elem] = recipesCopy
	}

	return result
}

// Generate a unique signature for a path based on all steps
func generateSignature(path Path) string {
	// Create a map to store ingredient combinations used in the path
	stepMap := make(map[string]bool)

	// Add each step to the map
	for _, step := range path.Steps {
		key := fmt.Sprintf("%s+%s=%s", step.Ingredients[0], step.Ingredients[1], step.Result)
		stepMap[key] = true
	}

	// Create a sorted list of steps for consistent signature
	var steps []string
	for step := range stepMap {
		steps = append(steps, step)
	}
	sort.Strings(steps)

	// Join all steps into a single string
	signature := ""
	for _, step := range steps {
		signature += step + "|"
	}

	return signature
}

// Function to find a single recipe
func FindSingleRecipeDFS(recipesFile, targetElement string, startingElements []string) *Path {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	// Cari path + info durasi dan node
	paths, duration, visited := findPathDFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil
	}

	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}

	// Tambahan info
	fmt.Printf("⏱ Time taken to search: %v\n", duration)
	fmt.Printf("📦 Nodes visited: %d\n", visited)

	return &path
}

func FindMultipleRecipesConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding up to %d paths to create: %s\n", maxRecipes, targetElement)

	// Check if target exists
	found := false
	for _, r := range recipes {
		if r.Element == targetElement {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Target element '%s' not found in recipes\n", targetElement)
		return
	}

	// Create recipe variations
	variations := make([][]ElementRecipe, maxRecipes*5)
	variations[0] = make([]ElementRecipe, len(recipes))
	copy(variations[0], recipes)
	for i := 1; i < len(variations); i++ {
		variations[i] = createRecipeVariation(recipes, i)
	}

	var (
		allPaths       []Path
		pathSignatures = map[string]bool{}
		totalVisited   int
		mu             sync.Mutex
		wg             sync.WaitGroup
		startTime      = time.Now()
	)

	resultChan := make(chan struct {
		path    Path
		visited int
		varIdx  int
	}, len(variations))

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)

	// Collector goroutine
	go func() {
		for result := range resultChan {
			mu.Lock()
			sig := generateSignature(result.path)
			if !pathSignatures[sig] && len(allPaths) < maxRecipes {
				pathSignatures[sig] = true
				allPaths = append(allPaths, result.path)
				totalVisited += result.visited
				fmt.Printf("Found path #%d with %d steps (variation %d)\n", len(allPaths), len(result.path.Steps), result.varIdx)
			}
			mu.Unlock()
		}
	}()

	// Launch all variations
	for varIdx, recipeVariation := range variations {
		mu.Lock()
		if len(allPaths) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		sem <- struct{}{}

		go func(recipes []ElementRecipe, idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			type localResult struct {
				paths   []Path
				visited int
			}

			innerChan := make(chan localResult, 1)

			go func() {
				paths, _, visited := findPathDFS(recipes, startingElements, targetElement)
				innerChan <- localResult{paths, visited}
			}()

			select {
			case result := <-innerChan:
				if len(result.paths) > 0 {
					resultChan <- struct {
						path    Path
						visited int
						varIdx  int
					}{result.paths[0], result.visited, idx}
				}
			case <-time.After(5 * time.Second):
				fmt.Printf("Search variation %d timed out\n", idx)
			}
		}(recipeVariation, varIdx)
	}

	// Wait until all search goroutines finish
	wg.Wait()
	close(resultChan)

	// Sort and print results
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	duration := time.Since(startTime)
	if len(allPaths) == 0 {
		fmt.Println("No valid paths found")
		return
	}

	fmt.Printf("\n📦 Total visited nodes: %d\n", totalVisited)
	fmt.Printf("⏱ Time taken: %v\n", duration)

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

// Create a variation of recipes by shuffling recipe orders
func createRecipeVariation(recipes []ElementRecipe, seed int) []ElementRecipe {
	// Create a deep copy of recipes
	variation := make([]ElementRecipe, len(recipes))

	// Seed the random number generator
	r := rand.New(rand.NewSource(int64(seed)))

	// Copy and possibly shuffle each recipe's combinations
	for i, recipe := range recipes {
		// Make a deep copy of the recipe
		recipeCopy := ElementRecipe{
			Element:  recipe.Element,
			ImageURL: recipe.ImageURL,
			Tier:     recipe.Tier,
		}

		// Only shuffle if there are multiple recipes
		if len(recipe.Recipes) > 1 {
			// Deep copy recipes
			recipesCopy := make([][2]string, len(recipe.Recipes))
			copy(recipesCopy, recipe.Recipes)

			// Shuffle the recipes
			r.Shuffle(len(recipesCopy), func(i, j int) {
				recipesCopy[i], recipesCopy[j] = recipesCopy[j], recipesCopy[i]
			})

			recipeCopy.Recipes = recipesCopy
		} else {
			// Just copy without shuffling
			recipeCopy.Recipes = make([][2]string, len(recipe.Recipes))
			copy(recipeCopy.Recipes, recipe.Recipes)
		}

		variation[i] = recipeCopy
	}

	return variation
}

// Helper to check if an element is in the list of basic elements
func isBasicElement(element string, basicElements []string) bool {
	for _, basic := range basicElements {
		if basic == element {
			return true
		}
	}
	return false
}

// Generate a simple signature for a path
func generateSimpleSignature(path Path) string {
	// Sort steps for consistent signature
	steps := make([]Step, len(path.Steps))
	copy(steps, path.Steps)

	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Result < steps[j].Result
	})

	var builder strings.Builder
	for _, step := range steps {
		fmt.Fprintf(&builder, "%s=%s+%s;", step.Result, step.Ingredients[0], step.Ingredients[1])
	}

	return builder.String()
}

// Helper to get tier of element, defaulting to 0 for basics
func getTier(element string, tierMap map[string]int) int {
	if tier, exists := tierMap[element]; exists {
		return tier
	}
	return 0 // Default tier for unknown elements (usually basics)
}

// Verify that a path follows tier constraints
// For each step, result tier must be greater than ingredients tier
func isValidPath(path Path, tierMap map[string]int) bool {
	for _, step := range path.Steps {
		resultTier := getTier(step.Result, tierMap)
		ing1Tier := getTier(step.Ingredients[0], tierMap)
		ing2Tier := getTier(step.Ingredients[1], tierMap)

		// Strict check: BOTH ingredients must have tier less than result
		if ing1Tier >= resultTier || ing2Tier >= resultTier {
			return false
		}
	}
	return true
}

func FindSingleRecipeBFS(recipesFile, targetElement string, startingElements []string) *Path {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	// Cari path + info durasi dan node
	paths, duration, visited := findPathBFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil
	}

	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}

	// Tambahan info
	fmt.Printf("⏱ Time taken to search: %v\n", duration)
	fmt.Printf("📦 Nodes visited: %d\n", visited)

	return &path
}

func FindMultipleRecipesBFSConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding up to %d paths to create: %s\n", maxRecipes, targetElement)

	// Build recipe maps for analysis
	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
	}

	// Get the target recipe
	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		fmt.Printf("Target element '%s' not found in recipes\n", targetElement)
		return
	}
	targetTier := targetRecipe.Tier

	// Find valid combinations that can create the target
	// Check tier constraints upfront
	var validCombinations [][2]string
	for _, combo := range targetRecipe.Recipes {
		a, b := combo[0], combo[1]
		aTier := getTier(a, tierMap)
		bTier := getTier(b, tierMap)
		// Strict check: Both ingredients must have tier less than target
		if aTier < targetTier && bTier < targetTier {
			validCombinations = append(validCombinations, combo)
		}
	}

	if len(validCombinations) == 0 {
		fmt.Printf("No valid recipes found for element '%s'\n", targetElement)
		return
	}

	// Find all alternative paths to the ingredients of the target
	var alternativeIngredientPaths = make(map[string][][2]string)
	// For each valid combination to make target
	for _, combo := range validCombinations {
		ingredient1, ingredient2 := combo[0], combo[1]
		// For each non-basic ingredient, find its recipes
		for _, ingredient := range []string{ingredient1, ingredient2} {
			// Skip if it's a basic element or we already found alternatives
			if isBasicElement(ingredient, startingElements) || len(alternativeIngredientPaths[ingredient]) > 0 {
				continue
			}
			// Find the recipes for this ingredient
			if ingRecipe, exists := elementMap[ingredient]; exists {
				// Add only valid recipes (tier check)
				var validRecipes [][2]string
				ingTier := ingRecipe.Tier
				for _, ingCombo := range ingRecipe.Recipes {
					a, b := ingCombo[0], ingCombo[1]
					aTier := getTier(a, tierMap)
					bTier := getTier(b, tierMap)
					// Strict check: Both ingredients must have tier less than this ingredient
					if aTier < ingTier && bTier < ingTier {
						validRecipes = append(validRecipes, ingCombo)
					}
				}
				alternativeIngredientPaths[ingredient] = validRecipes
			}
		}
	}

	// Synchronization primitives
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 10)

	// Results
	var allPaths []Path
	var pathSignatures = make(map[string]bool)
	var totalVisited int
	var maxDuration time.Duration

	// First approach: Use the standard direct approach for first set of paths
	for _, combo := range validCombinations {
		mu.Lock()
		if len(allPaths) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		sem <- struct{}{}
		wg.Add(1)
		go func(combo [2]string) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			var combinedPaths []Path
			localVisited := 0

			// For each ingredient
			for _, ingredient := range combo {
				if isBasicElement(ingredient, startingElements) {
					continue
				}

				// Find path to this ingredient using BFS
				ingredientPaths, duration, visited := findPathBFS(recipes, startingElements, ingredient)
				if len(ingredientPaths) > 0 {
					// Verify the path follows tier constraints
					if isValidPath(ingredientPaths[0], tierMap) {
						combinedPaths = append(combinedPaths, ingredientPaths[0])
						localVisited += visited
						mu.Lock()
						if duration > maxDuration {
							maxDuration = duration
						}
						mu.Unlock()
					} else {
						return
					}
				} else {
					return
				}
			}

			// Combine steps without duplicates
			stepSet := make(map[[3]string]bool)
			var steps []Step
			for _, path := range combinedPaths {
				for _, s := range path.Steps {
					key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
					if !stepSet[key] {
						stepSet[key] = true
						steps = append(steps, s)
					}
				}
			}

			// Add final step
			finalStep := Step{Ingredients: combo, Result: targetElement}
			key := [3]string{combo[0], combo[1], targetElement}
			if !stepSet[key] {
				stepSet[key] = true
				steps = append(steps, finalStep)
			}

			finalPath := Path{
				Steps:     steps,
				FinalItem: targetElement,
			}

			// Final check: ensure the whole path is valid
			if isValidPath(finalPath, tierMap) {
				signature := generateSimpleSignature(finalPath)
				mu.Lock()
				totalVisited += localVisited
				if !pathSignatures[signature] && len(allPaths) < maxRecipes {
					pathSignatures[signature] = true
					allPaths = append(allPaths, finalPath)
					if len(allPaths) >= maxRecipes {
						cancel()
					}
				}
				mu.Unlock()
			}
		}(combo)
	}

	// Second approach: Use alternative paths for ingredients
	for _, combo := range validCombinations {
		mu.Lock()
		if len(allPaths) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		// For each ingredient in the combo
		for ingIndex, ingredient := range combo {
			// Skip basic elements
			if isBasicElement(ingredient, startingElements) {
				continue
			}
			// Get alternative ways to create this ingredient
			alternatives := alternativeIngredientPaths[ingredient]
			if len(alternatives) == 0 {
				continue
			}
			// Try each alternative recipe for this ingredient
			for _, altCombo := range alternatives {
				mu.Lock()
				if len(allPaths) >= maxRecipes {
					mu.Unlock()
					break
				}
				mu.Unlock()

				sem <- struct{}{}
				wg.Add(1)
				go func(targetCombo [2]string, ingIndex int, ingredient string, altCombo [2]string) {
					defer wg.Done()
					defer func() { <-sem }()

					select {
					case <-ctx.Done():
						return
					default:
					}

					// Step 1: Find paths for both ingredients of the alternative combo
					var altIngredientPaths []Path
					localVisited := 0
					for _, altIng := range altCombo {
						if isBasicElement(altIng, startingElements) {
							continue
						}
						paths, duration, visited := findPathBFS(recipes, startingElements, altIng)
						if len(paths) > 0 {
							// Verify path follows tier constraints
							if isValidPath(paths[0], tierMap) {
								altIngredientPaths = append(altIngredientPaths, paths[0])
								localVisited += visited
								mu.Lock()
								if duration > maxDuration {
									maxDuration = duration
								}
								mu.Unlock()
							} else {
								return
							}
						} else {
							return
						}
					}

					// Step 2: Find path for the other ingredient in the target combo
					var otherIngredientPath []Path
					otherIngredient := targetCombo[1-ingIndex]
					if !isBasicElement(otherIngredient, startingElements) {
						paths, duration, visited := findPathBFS(recipes, startingElements, otherIngredient)
						if len(paths) > 0 {
							// Verify path follows tier constraints
							if isValidPath(paths[0], tierMap) {
								otherIngredientPath = append(otherIngredientPath, paths[0])
								localVisited += visited
								mu.Lock()
								if duration > maxDuration {
									maxDuration = duration
								}
								mu.Unlock()
							} else {
								return
							}
						} else {
							return
						}
					}

					// Step 3: Combine all steps
					stepSet := make(map[[3]string]bool)
					var steps []Step
					// Add steps for alternative ingredients
					for _, path := range altIngredientPaths {
						for _, s := range path.Steps {
							key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
							if !stepSet[key] {
								stepSet[key] = true
								steps = append(steps, s)
							}
						}
					}

					// Add step to create the ingredient using alternative
					// Verify tier constraints for this step
					ingTier := getTier(ingredient, tierMap)
					alt1Tier := getTier(altCombo[0], tierMap)
					alt2Tier := getTier(altCombo[1], tierMap)
					if alt1Tier < ingTier && alt2Tier < ingTier {
						ingredientStep := Step{Ingredients: altCombo, Result: ingredient}
						ingredientKey := [3]string{altCombo[0], altCombo[1], ingredient}
						if !stepSet[ingredientKey] {
							stepSet[ingredientKey] = true
							steps = append(steps, ingredientStep)
						}
					} else {
						return
					}

					// Add steps for other ingredient
					for _, path := range otherIngredientPath {
						for _, s := range path.Steps {
							key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
							if !stepSet[key] {
								stepSet[key] = true
								steps = append(steps, s)
							}
						}
					}

					// Add final step to create target
					// Tier check for final step was already done when building validCombinations
					finalStep := Step{Ingredients: targetCombo, Result: targetElement}
					finalKey := [3]string{targetCombo[0], targetCombo[1], targetElement}
					if !stepSet[finalKey] {
						stepSet[finalKey] = true
						steps = append(steps, finalStep)
					}

					// Create final path
					finalPath := Path{
						Steps:     steps,
						FinalItem: targetElement,
					}

					// Final check: ensure the whole path is valid
					if isValidPath(finalPath, tierMap) {
						signature := generateSimpleSignature(finalPath)
						mu.Lock()
						totalVisited += localVisited
						if !pathSignatures[signature] && len(allPaths) < maxRecipes {
							pathSignatures[signature] = true
							allPaths = append(allPaths, finalPath)
							if len(allPaths) >= maxRecipes {
								cancel()
							}
						}
						mu.Unlock()
					}
				}(combo, ingIndex, ingredient, altCombo)
			}
		}
	}

	wg.Wait()

	// Sort paths by number of steps (shortest first)
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	// Print results
	if len(allPaths) == 0 {
		fmt.Println("No valid paths found")
		return
	}

	fmt.Printf("\n📦 Total visited nodes: %d\n", totalVisited)
	fmt.Printf("⏱ Time taken: %v\n", maxDuration)

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

func shuffleRecipes(elements map[string][][]string, seed int) {
	rng := rand.New(rand.NewSource(int64(seed)))
	for elem, recipes := range elements {
		for i := len(recipes) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			recipes[i], recipes[j] = recipes[j], recipes[i]
		}
		elements[elem] = recipes
	}
}
