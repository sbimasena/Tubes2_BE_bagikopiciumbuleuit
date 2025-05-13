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

	"github.com/gorilla/websocket"
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

type StepMessage struct {
	Type     string `json:"type"` // "step" atau "done"
	Step     *Step  `json:"step,omitempty"`
	Visited  int    `json:"visited,omitempty"`
	Duration string `json:"duration,omitempty"`
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
) ([][]string, []map[string][]string, int, time.Duration) {
	if algorithm == "bidirectional" {
		if len(bidiStrategy) == 0 {
			return nil, nil, 0, 0
		}
		strategy := bidiStrategy[0]
		switch strategy {
		case "dfs":
			return BiSearchMultipleDFS(target, elements, basicElements, maxPaths, tierMap)
		case "bfs":
			return BiSearchMultipleBFS(target, elements, basicElements, maxPaths, tierMap)
		default:
			return nil, nil, 0, 0
		}
	} else if algorithm == "dfs" {
		return BiSearchMultipleDFS(target, elements, basicElements, maxPaths, tierMap)
	} else if algorithm == "bfs" {
		return BiSearchMultipleBFS(target, elements, basicElements, maxPaths, tierMap)
	}
	return nil, nil, 0, 0
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

func BiSearchMultipleBFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int, time.Duration) {
	var (
		paths          [][]string
		allSteps       []map[string][]string
		totalNodes     int
		startTime      = time.Now()
		pathSignatures = make(map[string]bool, maxPaths)
	)
	fmt.Println("Finding up to", maxPaths, "different paths for", target)

	var wg sync.WaitGroup
	resultChan := make(chan struct {
		path  []string
		steps map[string][]string
		nodes int
	}, maxPaths)

	attemptsToRun := maxPaths * 3

	for attempt := 0; attempt < attemptsToRun; attempt++ {
		wg.Add(1)
		go func(attemptNum int) {
			defer wg.Done()

			if len(pathSignatures) >= maxPaths {
				return
			}

			elementsCopy := copyElements(elements)
			shuffleRecipes(elementsCopy, attemptNum)

			path, steps, nodes, _ := BiSearchBFS(target, elementsCopy, basicElements, tierMap)
			if path == nil {
				return
			}

			resultChan <- struct {
				path  []string
				steps map[string][]string
				nodes int
			}{path, steps, nodes}
		}(attempt)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		// Early exit if we have enough paths
		if len(paths) >= maxPaths {
			break
		}

		signature := hashPath(result.path)
		if pathSignatures[signature] {
			continue
		}

		pathSignatures[signature] = true
		paths = append(paths, result.path)
		allSteps = append(allSteps, result.steps)
		totalNodes += result.nodes

		if len(paths) >= maxPaths {
			break
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("Total nodes explored: %d\n", totalNodes)
	fmt.Printf("Total execution time: %v\n", duration)

	if len(paths) < maxPaths {
		fmt.Printf("Only found %d different paths (of %d requested)\n", len(paths), maxPaths)
	} else {
		fmt.Printf("Found %d different paths\n", len(paths))
	}

	return paths, allSteps, totalNodes, duration
}

func hashPath(path []string) string {
	var b strings.Builder
	b.Grow(len(path) * 16)

	for i, elem := range path {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(elem)
	}
	return b.String()
}

func BiSearchMultipleDFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int, time.Duration) {
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

	return paths, allSteps, totalNodes, duration
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
	resultMap := make(map[string]bool)

	for _, step := range path.Steps {
		if resultMap[step.Result] {
			return ""
		}
		resultMap[step.Result] = true
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

	for _, elem := range startingElements {
		if elem == targetElement {
			fmt.Printf("Target element '%s' is already in starting elements\n", targetElement)
			return nil, 0, 0
		}
	}

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

	for _, elem := range startingElements {
		if elem == targetElement {
			fmt.Printf("Target element '%s' is already in starting elements\n", targetElement)
			return nil, 0, 0
		}
	}

	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

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
		return nil, 0, 0
	}

	// create recipe variations
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
		mu             sync.Mutex     // mutex utk antar goroutine
		wg             sync.WaitGroup // utk nungu semua goroutine selesai
		startTime      = time.Now()
	)

	resultChan := make(chan struct { // utk kirim hasil antar goroutine
		path    Path
		visited int
		varIdx  int
	}, len(variations))

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent) // manual semaphore

	go func() {
		for result := range resultChan {
			mu.Lock()
			sig := generateSignature(result.path)
			if !pathSignatures[sig] && len(allPaths) < maxRecipes {
				pathSignatures[sig] = true
				allPaths = append(allPaths, result.path)
				totalVisited += result.visited
			}
			mu.Unlock()
		}
	}()

	// loop all variation parallelly
	for varIdx, recipeVariation := range variations {
		mu.Lock()
		if len(allPaths) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		sem <- struct{}{} // masuk ke semaphore
		go func(recipes []ElementRecipe, idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			type localResult struct {
				paths   []Path
				visited int
			}

			innerChan := make(chan localResult, 1) // jalankin dfs dlam goroutine
			go func() {
				paths, _, visited := findPathDFS(recipes, startingElements, targetElement)
				innerChan <- localResult{paths, visited}
			}()

			result := <-innerChan // masukkin ke result channel utama
			if len(result.paths) > 0 {
				resultChan <- struct {
					path    Path
					visited int
					varIdx  int
				}{result.paths[0], result.visited, idx}
			}
		}(recipeVariation, varIdx)
	}

	wg.Wait()
	close(resultChan)

	// urutin dari paling pendek
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	duration := time.Since(startTime)

	return allPaths, totalVisited, duration
}

func createRecipeVariation(recipes []ElementRecipe, seed int) []ElementRecipe {
	variation := make([]ElementRecipe, len(recipes))

	r := rand.New(rand.NewSource(int64(seed)))

	for i, recipe := range recipes {
		recipeCopy := ElementRecipe{
			Element:  recipe.Element,
			ImageURL: recipe.ImageURL,
			Tier:     recipe.Tier,
		}
		if len(recipe.Recipes) > 1 {
			// deep copy recipes
			recipesCopy := make([][2]string, len(recipe.Recipes))
			copy(recipesCopy, recipe.Recipes)

			// shuffle the recipes
			r.Shuffle(len(recipesCopy), func(i, j int) {
				recipesCopy[i], recipesCopy[j] = recipesCopy[j], recipesCopy[i]
			})

			recipeCopy.Recipes = recipesCopy
		} else {
			// just copy without shuffling
			recipeCopy.Recipes = make([][2]string, len(recipe.Recipes))
			copy(recipeCopy.Recipes, recipe.Recipes)
		}

		variation[i] = recipeCopy
	}

	return variation
}

func isBasicElement(element string, basicElements []string) bool {
	return slices.Contains(basicElements, element)
}

func FindSingleRecipeBFS(recipesFile, targetElement string, startingElements []string) (*Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)

	for _, elem := range startingElements {
		if elem == targetElement {
			fmt.Printf("Target element '%s' is already in starting elements\n", targetElement)
			return nil, 0, 0
		}
	}

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
	startTime := time.Now()
	recipes, err := LoadRecipes(recipesFile)

	for _, elem := range startingElements {
		if elem == targetElement {
			fmt.Printf("Target element '%s' is already in starting elements\n", targetElement)
			return nil, 0, 0
		}
	}

	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	// Check if target exists and get its recipes
	targetRecipes := make([][2]string, 0)
	for _, r := range recipes {
		if r.Element == targetElement {
			targetRecipes = r.Recipes
			break
		}
	}
	if len(targetRecipes) == 0 {
		fmt.Printf("Target element '%s' not found in recipes\n", targetElement)
		return nil, 0, 0
	}

	// Build recipe maps and hierarchies for faster lookups
	elemToRecipes := make(map[string][][2]string)
	tierMap := make(map[string]int)

	for _, recipe := range recipes {
		elemToRecipes[recipe.Element] = recipe.Recipes
		tierMap[recipe.Element] = recipe.Tier
	}

	// Create recipe variations to increase diversity of results
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
	)

	// Channel for collecting results from goroutines
	resultChan := make(chan struct {
		path    Path
		visited int
		varIdx  int
	}, len(variations)*2)

	const maxConcurrent = 12
	sem := make(chan struct{}, maxConcurrent)

	// Collector goroutine
	go func() {
		for result := range resultChan {
			mu.Lock()
			sig := generateSignature(result.path)
			if !pathSignatures[sig] && len(allPaths) < maxRecipes && sig != "" {
				pathSignatures[sig] = true
				allPaths = append(allPaths, result.path)
				totalVisited += result.visited
			}
			mu.Unlock()
		}
	}()

	// Create context for cancellation, but with a longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for comboIdx, combo := range targetRecipes {
		if comboIdx >= maxRecipes*2 {
			break
		}

		a, b := combo[0], combo[1]
		aTier, aExists := tierMap[a]
		bTier, bExists := tierMap[b]
		targetTier, targetExists := tierMap[targetElement]

		// Skip this combo if tier constraint would be violated
		if !aExists || !bExists || !targetExists ||
			aTier >= targetTier || bTier >= targetTier {
			continue // Skip if any ingredient has equal or higher tier
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(ingredients [2]string) {
			defer wg.Done()
			defer func() { <-sem }()

			// For each combination, try to find paths for both ingredients
			var allIngPaths [][]Path
			var totalIteration int

			// Try to find paths for both ingredients
			for _, ing := range ingredients {
				// Skip if it's a basic element
				if isBasicElement(ing, startingElements) {
					continue
				}

				// Find path for this ingredient
				var ingPathsForThisIng []Path
				for _, recipeVariation := range variations {
					ingPaths, _, iterations := findPathBFS(recipeVariation, startingElements, ing)
					totalIteration += iterations
					if len(ingPaths) > 0 {
						ingPathsForThisIng = append(ingPathsForThisIng, ingPaths...)
					}
				}

				if len(ingPathsForThisIng) > 0 {
					allIngPaths = append(allIngPaths, ingPathsForThisIng)
				}
			}

			// If we found paths for all needed ingredients, generate combinations
			if len(allIngPaths) > 0 {
				// Function to recursively generate all possible path combinations
				var generatePathCombinations func(currentIndex int, currentCombination []Path)
				var pathCombinations [][]Path

				generatePathCombinations = func(currentIndex int, currentCombination []Path) {
					// If we've processed all ingredients, add this combination
					if currentIndex >= len(allIngPaths) {
						// Create a deep copy of the combination
						combinationCopy := make([]Path, len(currentCombination))
						copy(combinationCopy, currentCombination)
						pathCombinations = append(pathCombinations, combinationCopy)
						return
					}

					// Try each path for the current ingredient
					for _, path := range allIngPaths[currentIndex] {
						currentCombination = append(currentCombination, path)
						generatePathCombinations(currentIndex+1, currentCombination)
						currentCombination = currentCombination[:len(currentCombination)-1] // Backtrack
					}
				}

				// Start generating combinations
				generatePathCombinations(0, []Path{})

				// Create different path variations from the combinations
				for i, combo := range pathCombinations {
					// Create a combined path for this variation
					var combinedPath Path
					combinedPath.FinalItem = targetElement
					combinedPath.Steps = make([]Step, 0)

					// Track used steps to avoid duplicates
					stepMap := make(map[[3]string]bool)

					// Add steps from all paths in this combination
					for _, p := range combo {
						for _, s := range p.Steps {
							key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
							if !stepMap[key] {
								stepMap[key] = true
								combinedPath.Steps = append(combinedPath.Steps, s)
							}
						}
					}

					// Add the final step
					finalStep := Step{
						Ingredients: ingredients,
						Result:      targetElement,
					}
					combinedPath.Steps = append(combinedPath.Steps, finalStep)

					// Submit this path variation
					resultChan <- struct {
						path    Path
						visited int
						varIdx  int
					}{combinedPath, totalIteration, i}
				}
			}
		}(combo)
	}

	// Launch goroutines for each recipe variation
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

			// Use channel with timeout to prevent hanging
			innerChan := make(chan struct {
				paths   []Path
				visited int
			}, 1)

			go func() {
				select {
				case <-ctx.Done():
					return
				default:
					paths, _, visited := findPathBFS(recipes, startingElements, targetElement)
					if len(paths) > 0 {
						innerChan <- struct {
							paths   []Path
							visited int
						}{paths, visited}
					}
				}
			}()

			// Wait for result with timeout (longer timeout)
			select {
			case result := <-innerChan:
				for _, p := range result.paths {
					mu.Lock()
					if len(allPaths) >= maxRecipes {
						mu.Unlock()
						break
					}
					mu.Unlock()

					resultChan <- struct {
						path    Path
						visited int
						varIdx  int
					}{p, result.visited, idx}
				}

				// Cancel if we have enough results
				mu.Lock()
				if len(allPaths) >= maxRecipes {
					cancel()
				}
				mu.Unlock()
			case <-time.After(8 * time.Second): // Longer timeout
				fmt.Print("Timeout for variation", idx, "after 8 seconds\n")
				// Timeout, continue with other variations
				return
			case <-ctx.Done():
				return
			}
		}(recipeVariation, varIdx)
	}

	wg.Wait()
	close(resultChan)

	// Sort paths by number of steps (shortest first)
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	duration := time.Since(startTime)
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

// Single DFS live
func FindSingleRecipeDFSLive(recipes []ElementRecipe, target string, starting []string, conn *websocket.Conn) error {
	path, visited, duration := DFSWithCallback(recipes, starting, target, func(step Step) {
		msg := StepMessage{
			Type: "step",
			Step: &step,
		}
		conn.WriteJSON(msg)
	})

	if path == nil {
		return fmt.Errorf("no path found")
	}

	conn.WriteJSON(StepMessage{
		Type:     "done",
		Visited:  visited,
		Duration: duration.String(),
	})
	return nil
}

// Single BFS live
func FindSingleRecipeBFSLive(recipes []ElementRecipe, target string, starting []string, conn *websocket.Conn) error {
	path, visited, duration := BFSWithCallback(recipes, starting, target, func(step Step) {
		msg := StepMessage{
			Type: "step",
			Step: &step,
		}
		conn.WriteJSON(msg)
	})

	if path == nil {
		return fmt.Errorf("no path found")
	}

	conn.WriteJSON(StepMessage{
		Type:     "done",
		Visited:  visited,
		Duration: duration.String(),
	})
	return nil
}

func FindMultipleRecipesDFSLive(recipes []ElementRecipe, target string, starting []string, maxPaths int, conn *websocket.Conn) error {
	sendStep := func(step Step) {
		msg := StepMessage{
			Type: "step",
			Step: &step,
		}
		fmt.Println("🧪 DFS STEP:", step)
		time.Sleep(100 * time.Millisecond)
		conn.WriteJSON(msg)
	}

	_, visited, duration := DFSAllWithCallback(recipes, starting, target, maxPaths, sendStep)

	conn.WriteJSON(StepMessage{
		Type:     "done",
		Visited:  visited,
		Duration: duration.String(),
	})

	return nil
}

func FindMultipleRecipesBFSLive(recipes []ElementRecipe, target string, starting []string, maxPaths int, conn *websocket.Conn) error {
	sendStep := func(step Step) {
		msg := StepMessage{
			Type: "step",
			Step: &step,
		}
		conn.WriteJSON(msg)
	}

	_, visited, duration := BFSAllWithCallback(recipes, starting, target, maxPaths, sendStep)

	conn.WriteJSON(StepMessage{
		Type:     "done",
		Visited:  visited,
		Duration: duration.String(),
	})
	return nil
}

func DFSWithCallback(
	recipes []ElementRecipe,
	startingElements []string,
	target string,
	onStep func(Step),
) ([]Step, int, time.Duration) {
	visited := make(map[string]bool)
	start := time.Now()
	var nodesVisited int
	var path []Step

	var dfs func(curr string) bool
	dfs = func(curr string) bool {
		nodesVisited++
		if contains(startingElements, curr) {
			return true
		}
		if visited[curr] {
			return false
		}
		visited[curr] = true

		for _, recipe := range recipes {
			if recipe.Element != curr {
				continue
			}
			for _, ing := range recipe.Recipes {
				if dfs(ing[0]) && dfs(ing[1]) {
					path = append(path, Step{
						Ingredients: ing,
						Result:      curr,
					})
					return true
				}
			}
		}
		return false
	}

	ok := dfs(target)
	if !ok {
		return nil, nodesVisited, time.Since(start)
	}

	// Kirim step secara terbalik, agar dari akar ke target
	for i := len(path) - 1; i >= 0; i-- {
		onStep(path[i])
	}

	return path, nodesVisited, time.Since(start)
}

// BFS SINGLE PATH WITH CALLBACK
func BFSWithCallback(recipes []ElementRecipe, startingElements []string, target string, onStep func(Step)) ([]Step, int, time.Duration) {
	start := time.Now()
	type Node struct {
		Name  string
		Path  []Step
		Known map[string]bool
	}

	queue := []Node{{Name: "", Known: map[string]bool{}, Path: nil}}
	elements := make(map[string]bool)
	for _, e := range startingElements {
		elements[e] = true
	}

	for {
		if len(queue) == 0 {
			return nil, len(elements), time.Since(start)
		}

		n := queue[0]
		queue = queue[1:]

		for _, recipe := range recipes {
			for _, ing := range recipe.Recipes {
				if elements[ing[0]] && elements[ing[1]] && !elements[recipe.Element] {
					newStep := Step{Ingredients: ing, Result: recipe.Element}
					newPath := append([]Step{}, n.Path...)
					newPath = append(newPath, newStep)
					if recipe.Element == target {
						// hanya kirim step kalau path valid ke target
						for _, step := range newPath {
							onStep(step)
						}
						return newPath, len(elements), time.Since(start)
					}
					elements[recipe.Element] = true
					queue = append(queue, Node{Name: recipe.Element, Path: newPath})
				}
			}
		}
	}
}

// DFS MULTIPLE PATHS WITH CALLBACK (for concurrent mode)
func DFSAllWithCallback(recipes []ElementRecipe, startingElements []string, target string, maxPaths int, onStep func(Step)) ([][]Step, int, time.Duration) {
	var results [][]Step
	start := time.Now()
	var nodesVisited int
	var path []Step

	var dfs func(curr string) bool
	dfs = func(curr string) bool {
		if len(results) >= maxPaths {
			return true
		}
		nodesVisited++
		if contains(startingElements, curr) {
			return true
		}
		for _, recipe := range recipes {
			if recipe.Element != curr {
				continue
			}
			for _, ing := range recipe.Recipes {
				step := Step{Ingredients: ing, Result: curr}
				if dfs(ing[0]) && dfs(ing[1]) {
					path = append(path, step)
					onStep(step)
					if curr == target {
						copyPath := make([]Step, len(path))
						copy(copyPath, path)
						results = append(results, copyPath)
					}
					path = path[:len(path)-1]
				}
			}
		}
		return false
	}

	dfs(target)
	return results, nodesVisited, time.Since(start)
}

func BFSAllWithCallback(recipes []ElementRecipe, startingElements []string, target string, maxPaths int, onStep func(Step)) ([][]Step, int, time.Duration) {
	start := time.Now()
	type Node struct {
		Name string
		Path []Step
	}
	queue := []Node{}
	foundPaths := [][]Step{}
	visited := make(map[string]bool)
	known := make(map[string]bool)
	for _, e := range startingElements {
		known[e] = true
	}

	for {
		if len(queue) == 0 && len(foundPaths) == 0 {
			// Initial expansion
			for _, r := range recipes {
				for _, ing := range r.Recipes {
					if known[ing[0]] && known[ing[1]] && !known[r.Element] {
						step := Step{Ingredients: ing, Result: r.Element}
						onStep(step)
						queue = append(queue, Node{
							Name: r.Element,
							Path: []Step{step},
						})
						known[r.Element] = true
					}
				}
			}
		}

		if len(queue) == 0 || len(foundPaths) >= maxPaths {
			break
		}

		n := queue[0]
		queue = queue[1:]
		if n.Name == target {
			foundPaths = append(foundPaths, n.Path)
			continue
		}

		for _, r := range recipes {
			if r.Element == n.Name {
				continue
			}
			for _, ing := range r.Recipes {
				if !visited[r.Element] && known[ing[0]] && known[ing[1]] {
					step := Step{Ingredients: ing, Result: r.Element}
					onStep(step)
					newPath := append([]Step{}, n.Path...)
					newPath = append(newPath, step)
					queue = append(queue, Node{Name: r.Element, Path: newPath})
					visited[r.Element] = true
				}
			}
		}
	}

	return foundPaths, len(known), time.Since(start)
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
