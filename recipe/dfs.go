package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ElementRecipe struct from your existing code
type ElementRecipe struct {
	Element  string      `json:"element"`
	ImageURL string      `json:"image_url"`
	Recipes  [][2]string `json:"recipes"`
	Tier     int         `json:"tier"`
}

// Path represents a sequence of combinations to create an element
type Path struct {
	Steps     []Step `json:"steps"`
	FinalItem string `json:"final_item"`
}

// Step represents a single combination in a path
type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

func findPathDFS(recipes []ElementRecipe, startElements []string, target string) []Path {
	// Build lookup maps
	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	recipeMap := make(map[string][][2]string)
	resultMap := make(map[[2]string]string)

	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier

		for _, combo := range recipe.Recipes {
			recipeMap[recipe.Element] = append(recipeMap[recipe.Element], combo)
			key1 := [2]string{combo[0], combo[1]}
			key2 := [2]string{combo[1], combo[0]}
			resultMap[key1] = recipe.Element
			resultMap[key2] = recipe.Element
		}
	}

	for _, elem := range startElements {
		if _, exists := tierMap[elem]; !exists {
			tierMap[elem] = 1
		}
	}

	basics := make(map[string]bool)
	for _, e := range startElements {
		basics[e] = true
	}

	memo := make(map[string]*Path)

	var dfs func(string, map[string]bool) *Path
	dfs = func(current string, visited map[string]bool) *Path {
		if basics[current] {
			return &Path{Steps: []Step{}, FinalItem: current}
		}

		if visited[current] {
			return nil
		}

		if p, ok := memo[current]; ok {
			return p
		}

		visited[current] = true
		defer delete(visited, current)

		combos, ok := recipeMap[current]
		if !ok {
			return nil
		}

		var best *Path
		for _, combo := range combos {
			a, b := combo[0], combo[1]

			aTier, aOk := tierMap[a]
			bTier, bOk := tierMap[b]
			resultTier := tierMap[current]

			if !aOk || !bOk {
				continue
			}

			maxTier := aTier
			if bTier > maxTier {
				maxTier = bTier
			}
			if resultTier <= maxTier {
				continue
			}

			pathA := dfs(a, visited)
			if pathA == nil {
				continue
			}
			pathB := dfs(b, visited)
			if pathB == nil {
				continue
			}

			stepSet := make(map[[3]string]bool)
			var steps []Step

			for _, s := range pathA.Steps {
				k := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
				if !stepSet[k] {
					stepSet[k] = true
					steps = append(steps, s)
				}
			}
			for _, s := range pathB.Steps {
				k := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
				if !stepSet[k] {
					stepSet[k] = true
					steps = append(steps, s)
				}
			}

			k := [3]string{a, b, current}
			if !stepSet[k] {
				steps = append(steps, Step{Ingredients: [2]string{a, b}, Result: current})
			}

			p := &Path{Steps: steps, FinalItem: current}
			if best == nil || len(p.Steps) < len(best.Steps) {
				best = p
			}
		}

		memo[current] = best
		return best
	}

	visited := make(map[string]bool)
	path := dfs(target, visited)

	if path != nil {
		return []Path{*path}
	}
	return nil
}

// LoadRecipes loads element recipes from a JSON file
func LoadRecipes(filename string) ([]ElementRecipe, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var recipes []ElementRecipe
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}

// FindPathToElement finds the first path to create a specific element
func FindPathToElement(recipesFile, targetElement string, startingElements []string) *Path {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	// Cari path dengan DFS
	paths := findPathDFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil
	}

	// Ambil path pertama (DFS mode = first-found)
	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}
	return &path
}
