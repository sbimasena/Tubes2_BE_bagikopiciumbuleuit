package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"slices"
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

func FindSingleRecipeBi(
	target string,
	elements map[string][][]string,
	basicElements map[string]bool,
	algorithm string, // "bfs", "dfs", "bidirectional"
	tierMap map[string]int,
	bidiStrategy ...string, // optional: ["dfs"] or ["bfs"] if bidirectional
) ([]string, map[string][]string, int, time.Duration) {
	if algorithm == "bidirectional" {
		if len(bidiStrategy) == 0 {
			return nil, nil, 0, 0
		}
		strategy := bidiStrategy[0]
		switch strategy {
		case "dfs":
			return BiSearchDFS(target, elements, basicElements, tierMap)
		case "bfs":
			return BiSearchBFS(target, elements, basicElements, tierMap)
		default:
			return nil, nil, 0, 0
		}
	} else if algorithm == "dfs" {
		return BiSearchDFS(target, elements, basicElements, tierMap)
	} else if algorithm == "bfs" {
		return BiSearchBFS(target, elements, basicElements, tierMap)
	}
	return nil, nil, 0, 0
}

func FindMultipleRecipesBi(
	target string,
	elements map[string][][]string,
	basicElements map[string]bool,
	algorithm string, // "bfs", "dfs", "bidirectional"
	maxPaths int,
	tierMap map[string]int,
	bidiStrategy ...string, // optional
) ([][]string, []map[string][]string, int) {
	if algorithm == "bidirectional" {
		if len(bidiStrategy) == 0 {
			return nil, nil, 0
		}
		strategy := bidiStrategy[0]
		switch strategy {
		case "dfs":
			return BiSearchMultipleDFS(target, elements, basicElements, maxPaths, tierMap)
		case "bfs":
			return BiSearchMultipleBFS(target, elements, basicElements, maxPaths, tierMap)
		default:
			return nil, nil, 0
		}
	} else if algorithm == "dfs" {
		return BiSearchMultipleDFS(target, elements, basicElements, maxPaths, tierMap)
	} else if algorithm == "bfs" {
		return BiSearchMultipleBFS(target, elements, basicElements, maxPaths, tierMap)
	}
	return nil, nil, 0
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

	for attempt := 0; attempt < maxPaths*5; attempt++ {
		if len(paths) >= maxPaths {
			break
		}

		elementsCopy := copyElements(elements)
		shuffleRecipes(elementsCopy, attempt)

		var p []string
		var s map[string][]string
		var n int

		done := make(chan bool)

		go func() {
			p, s, n, _ = BiSearchBFS(target, elementsCopy, basicElements, tierMap)
			done <- true
		}()

		<-done

		if p == nil {
			continue
		}

		signature := fmt.Sprintf("%v", p)
		if pathSignatures[signature] {
			continue
		}

		pathSignatures[signature] = true
		paths = append(paths, p)
		allSteps = append(allSteps, s)
		totalNodes += n
	}

	duration := time.Since(startTime)
	fmt.Printf("Total nodes explored: %d\n", totalNodes)
	fmt.Printf("Total execution time: %v\n", duration)

	if len(paths) < maxPaths {
		fmt.Printf("Only found %d different paths (of %d requested)\n", len(paths), maxPaths)
	} else {
		fmt.Printf("Found %d different paths\n", len(paths))
	}

	return paths, allSteps, totalNodes
}

func BiSearchMultipleDFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	var (
		paths          [][]string
		allSteps       []map[string][]string
		totalNodes     int
		startTime      = time.Now()
		pathSignatures = make(map[string]bool)
	)

	fmt.Println("Finding up to", maxPaths, "different paths for", target)

	for attempt := 0; attempt < maxPaths*3; attempt++ {
		if len(paths) >= maxPaths {
			break
		}

		elementsCopy := copyElements(elements)
		shuffleRecipes(elementsCopy, attempt)

		var p []string
		var s map[string][]string
		var n int

		done := make(chan bool)
		go func() {
			p, s, n, _ = BiSearchDFS(target, elementsCopy, basicElements, tierMap)
			done <- true
		}()
		<-done

		if p == nil {
			continue
		}

		signature := fmt.Sprintf("%v", p)
		if pathSignatures[signature] {
			continue
		}

		pathSignatures[signature] = true
		paths = append(paths, p)
		allSteps = append(allSteps, s)
		totalNodes += n
	}

	duration := time.Since(startTime)
	fmt.Printf("Total nodes explored: %d\n", totalNodes)
	fmt.Printf("Total execution time: %v\n", duration)

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

func generateSignature(path Path) string {

	stepMap := make(map[string]bool)

	for _, step := range path.Steps {
		key := fmt.Sprintf("%s+%s=%s", step.Ingredients[0], step.Ingredients[1], step.Result)
		stepMap[key] = true
	}

	var steps []string
	for step := range stepMap {
		steps = append(steps, step)
	}
	sort.Strings(steps)

	signature := ""
	for _, step := range steps {
		signature += step + "|"
	}

	return signature
}

func FindSingleRecipeDFS(recipesFile, targetElement string, startingElements []string) (*Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	paths, duration, visited := findPathDFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil, 0, duration
	}

	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}

	fmt.Printf("Time taken to search: %v\n", duration)
	fmt.Printf("Nodes visited: %d\n", visited)

	return &path, visited, duration
}

func FindMultipleRecipesDFSConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) ([]Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Printf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
	}

	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		log.Printf("Target element '%s' not found in recipes", targetElement)
		return nil, 0, 0
	}
	// targetTier := targetRecipe.Tier

	var validCombinations [][2]string
	for _, combo := range targetRecipe.Recipes {
		validCombinations = append(validCombinations, combo) // ✅ biarkan isValidPath yang menyaring
	}
	if len(validCombinations) == 0 {
		log.Printf("No valid recipes found for element '%s'", targetElement)
		return nil, 0, 0
	}

	var (
		ctx            = context.Background()
		cancelCtx, _   = context.WithCancel(ctx)
		wg             sync.WaitGroup
		sem            = make(chan struct{}, 10)
		mu             sync.Mutex
		allPaths       []Path
		pathSignatures = make(map[string]bool)
		totalVisited   int
		startTime      = time.Now()
		resultChan     = make(chan struct {
			path    Path
			visited int
		}, len(validCombinations))
	)

	for _, combo := range validCombinations {
		sem <- struct{}{}
		wg.Add(1)

		go func(combo [2]string) {
			defer wg.Done()
			defer func() { <-sem }()

			var combinedSteps []Step
			visitedSet := make(map[[3]string]bool)
			localVisited := 0

			for _, ingredient := range combo {
				if isBasicElement(ingredient, startingElements) {
					continue
				}
				paths, _, visited := findPathDFS(recipes, startingElements, ingredient)
				if len(paths) == 0 {
					return
				}
				localVisited += visited
				for _, s := range paths[0].Steps {
					key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
					if !visitedSet[key] {
						visitedSet[key] = true
						combinedSteps = append(combinedSteps, s)
					}
				}
			}

			finalStep := Step{Ingredients: combo, Result: targetElement}
			key := [3]string{combo[0], combo[1], targetElement}
			if !visitedSet[key] {
				combinedSteps = append(combinedSteps, finalStep)
			}

			path := Path{Steps: combinedSteps, FinalItem: targetElement}

			resultChan <- struct {
				path    Path
				visited int
			}{path, localVisited}
		}(combo)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		mu.Lock()
		sig := generateSimpleSignature(result.path)
		if !pathSignatures[sig] && len(allPaths) < maxRecipes {
			pathSignatures[sig] = true
			allPaths = append(allPaths, result.path)
			totalVisited += result.visited
			if len(allPaths) >= maxRecipes {
				cancelCtx.Done()
			}
		}
		mu.Unlock()
	}

	duration := time.Since(startTime)

	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	return allPaths, totalVisited, duration
}

func isBasicElement(element string, basicElements []string) bool {
	return slices.Contains(basicElements, element)
}

func generateSimpleSignature(path Path) string {
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

func getTier(element string, tierMap map[string]int) int {
	if tier, exists := tierMap[element]; exists {
		return tier
	}
	return 0
}

func isValidPath(path Path, tierMap map[string]int) bool {
	for _, step := range path.Steps {
		resultTier := getTier(step.Result, tierMap)
		ing1Tier := getTier(step.Ingredients[0], tierMap)
		ing2Tier := getTier(step.Ingredients[1], tierMap)

		if ing1Tier >= resultTier || ing2Tier >= resultTier {
			return false
		}
	}
	return true
}

func FindSingleRecipeBFS(recipesFile, targetElement string, startingElements []string) (*Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	paths, duration, visited := findPathBFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil, 0, 0
	}

	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}

	fmt.Printf("⏱ Time taken to search: %v\n", duration)
	fmt.Printf("📦 Nodes visited: %d\n", visited)

	return &path, visited, duration
}

func FindMultipleRecipesBFSConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) ([]Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Printf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
	}

	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		log.Printf("Target element '%s' not found in recipes", targetElement)
		return nil, 0, 0
	}
	// targetTier := targetRecipe.Tier

	var validCombinations [][2]string
	for _, combo := range targetRecipe.Recipes {
		validCombinations = append(validCombinations, combo)
	}
	if len(validCombinations) == 0 {
		log.Printf("No valid recipes found for element '%s'", targetElement)
		return nil, 0, 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		wg             sync.WaitGroup
		mu             sync.Mutex
		sem            = make(chan struct{}, 10)
		allPaths       []Path
		totalVisited   int
		startTime      = time.Now()
		pathSignatures = make(map[string]bool)
	)

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

			for _, ingredient := range combo {
				if isBasicElement(ingredient, startingElements) {
					continue
				}
				paths, _, visited := findPathBFS(recipes, startingElements, ingredient)
				if len(paths) == 0 || !isValidPath(paths[0], tierMap) {
					return
				}
				combinedPaths = append(combinedPaths, paths[0])
				localVisited += visited
			}

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

			finalStep := Step{Ingredients: combo, Result: targetElement}
			key := [3]string{combo[0], combo[1], targetElement}
			if !stepSet[key] {
				steps = append(steps, finalStep)
			}

			finalPath := Path{Steps: steps, FinalItem: targetElement}
			if !isValidPath(finalPath, tierMap) {
				return
			}

			signature := generateSimpleSignature(finalPath)

			mu.Lock()
			if !pathSignatures[signature] && len(allPaths) < maxRecipes {
				pathSignatures[signature] = true
				allPaths = append(allPaths, finalPath)
				totalVisited += localVisited
				if len(allPaths) >= maxRecipes {
					cancel()
				}
			}
			mu.Unlock()
		}(combo)
	}

	wg.Wait()
	duration := time.Since(startTime)

	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	return allPaths, totalVisited, duration
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
