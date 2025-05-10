package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		"Time":  true,
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
	pathsChan := make(chan []string, maxPaths)
	stepsChan := make(chan map[string][]string, maxPaths)
	nodesChan := make(chan int, maxPaths)
	durationChan := make(chan time.Duration, maxPaths)

	var wg sync.WaitGroup
	wg.Add(maxPaths)

	elementsCopy := copyElements(elements)
	mutex := &sync.Mutex{}

	for i := 0; i < maxPaths; i++ {
		go func() {
			defer wg.Done()
			localElements := copyElements(elementsCopy)
			p, s, n, dur := BiSearchBFS(target, localElements, basicElements, tierMap)
			if p != nil {
				pathsChan <- p
				stepsChan <- s
				nodesChan <- n
				durationChan <- dur
				mutex.Lock()
				for k, v := range s {
					removeRecipe(elementsCopy, k, v)
					break
				}
				mutex.Unlock()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(pathsChan)
		close(stepsChan)
		close(nodesChan)
		close(durationChan)
	}()

	var paths [][]string
	var steps []map[string][]string
	total := 0
	var maxDur time.Duration

	for p := range pathsChan {
		paths = append(paths, p)
	}
	for s := range stepsChan {
		steps = append(steps, s)
	}
	for n := range nodesChan {
		total += n
	}
	for d := range durationChan {
		if d > maxDur {
			maxDur = d
		}
	}

	fmt.Printf("\n📦 Total simpul yang dieksplorasi: %d\n", total)
	fmt.Printf("⏱ Waktu eksekusi maksimum: %v\n", maxDur)

	return paths, steps, total
}

func BiSearchMultipleDFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	pathsChan := make(chan []string, maxPaths)
	stepsChan := make(chan map[string][]string, maxPaths)
	nodesChan := make(chan int, maxPaths)
	durationChan := make(chan time.Duration, maxPaths)

	var wg sync.WaitGroup
	wg.Add(maxPaths)

	elementsCopy := copyElements(elements)
	mutex := &sync.Mutex{}

	for i := 0; i < maxPaths; i++ {
		go func() {
			defer wg.Done()
			localElements := copyElements(elementsCopy)
			p, s, n, dur := BiSearchDFS(target, localElements, basicElements, tierMap)
			if p != nil {
				pathsChan <- p
				stepsChan <- s
				nodesChan <- n
				durationChan <- dur
				mutex.Lock()
				for k, v := range s {
					removeRecipe(elementsCopy, k, v)
					break
				}
				mutex.Unlock()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(pathsChan)
		close(stepsChan)
		close(nodesChan)
		close(durationChan)
	}()

	var paths [][]string
	var steps []map[string][]string
	total := 0
	var maxDur time.Duration

	for p := range pathsChan {
		paths = append(paths, p)
	}
	for s := range stepsChan {
		steps = append(steps, s)
	}
	for n := range nodesChan {
		total += n
	}
	for d := range durationChan {
		if d > maxDur {
			maxDur = d
		}
	}

	fmt.Printf("\n📦 Total simpul yang dieksplorasi: %d\n", total)
	fmt.Printf("⏱ Waktu eksekusi maksimum: %v\n", maxDur)

	return paths, steps, total
}

func copyElements(original map[string][][]string) map[string][][]string {
	newMap := make(map[string][][]string)
	for k, recipes := range original {
		var copiedRecipes [][]string
		for _, recipe := range recipes {
			cp := make([]string, len(recipe))
			copy(cp, recipe)
			copiedRecipes = append(copiedRecipes, cp)
		}
		newMap[k] = copiedRecipes
	}
	return newMap
}

func removeRecipe(elements map[string][][]string, elem string, recipe []string) {
	var filtered [][]string
	for _, r := range elements[elem] {
		if len(r) == len(recipe) && r[0] == recipe[0] && r[1] == recipe[1] {
			continue
		}
		filtered = append(filtered, r)
	}
	elements[elem] = filtered
}

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
	var totalVisited int
	var maxDuration time.Duration
	var statsMu sync.Mutex

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
			localVisited := 0

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

				ingredientPaths, duration, visited := findPathDFS(recipes, startingElements, ingredient)
				if len(ingredientPaths) > 0 {
					combinedPaths = append(combinedPaths, ingredientPaths[0])
					localVisited += visited
					statsMu.Lock()
					if duration > maxDuration {
						maxDuration = duration
					}
					statsMu.Unlock()
				} else {
					return
				}
			}

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

			finalStep := Step{Ingredients: combo.ingredients, Result: combo.result}
			k := [3]string{combo.ingredients[0], combo.ingredients[1], combo.result}
			if !stepSet[k] {
				steps = append(steps, finalStep)
			}

			finalPath := Path{
				Steps:     steps,
				FinalItem: targetElement,
			}

			statsMu.Lock()
			totalVisited += localVisited
			statsMu.Unlock()

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
