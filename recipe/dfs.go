package main

import (
	"encoding/json"
	"os"
	"time"
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

func findPathDFS(recipes []ElementRecipe, startElements []string, target string) ([]Path, time.Duration, int) {
	startTime := time.Now()

	// Build lookup maps
	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	recipeMap := make(map[string][][2]string)

	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
		for _, combo := range recipe.Recipes {
			recipeMap[recipe.Element] = append(recipeMap[recipe.Element], combo)
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

	// Modifikasi: memo hanya menyimpan apakah elemen sudah pernah berhasil dibuat atau tidak
	// Tidak menyimpan jalur terbaik untuk elemen tersebut
	memo := make(map[string]bool)
	visitedCounter := make(map[string]bool)

	var dfs func(string, map[string]bool) *Path
	dfs = func(current string, visited map[string]bool) *Path {
		if basics[current] {
			return &Path{Steps: []Step{}, FinalItem: current}
		}

		if visited[current] {
			return nil
		}

		// Cek memo: jika kita tahu elemen ini tidak bisa dibuat, langsung return nil
		if success, ok := memo[current]; ok && !success {
			return nil
		}

		visited[current] = true
		defer delete(visited, current)
		visitedCounter[current] = true

		combos, ok := recipeMap[current]
		if !ok {
			memo[current] = false // Catat bahwa elemen ini tidak bisa dibuat
			return nil
		}

		// Ubah dari mencari "best" (jalur terpendek) menjadi mengambil jalur pertama yang valid
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

			// Gabungkan langkah-langkah dari kedua jalur, menghilangkan duplikat
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

			// Tambahkan langkah untuk membuat elemen saat ini
			currentStep := Step{Ingredients: [2]string{a, b}, Result: current}
			k := [3]string{a, b, current}
			if !stepSet[k] {
				steps = append(steps, currentStep)
			}

			// Buat jalur lengkap
			p := &Path{Steps: steps, FinalItem: current}

			// Ubah dari memilih jalur terpendek menjadi mengambil jalur pertama yang valid
			memo[current] = true // Catat bahwa elemen ini bisa dibuat
			return p             // Return jalur pertama yang valid
		}

		// Jika tidak ada jalur yang valid, catat dalam memo
		memo[current] = false
		return nil
	}

	visited := make(map[string]bool)
	path := dfs(target, visited)
	duration := time.Since(startTime)

	if path != nil {
		return []Path{*path}, duration, len(visitedCounter)
	}

	return nil, duration, len(visitedCounter)
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
