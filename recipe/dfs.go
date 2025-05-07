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

func dfsWithMemo(target string, elements map[string][][]string, basicElements map[string]bool,
	memo map[string][]string, visited map[string]bool,
	combinations map[string][]string) []string {

	// Jika sudah dihitung sebelumnya, kembalikan hasil dari memo
	if recipe, found := memo[target]; found {
		return recipe
	}

	// Jika target adalah elemen dasar
	if basicElements[target] {
		memo[target] = []string{target}
		return memo[target]
	}

	// Tandai sebagai dikunjungi untuk menghindari siklus
	if visited[target] {
		return nil
	}
	visited[target] = true
	defer func() { visited[target] = false }()

	// Periksa kombinasi yang dapat membuat target
	targetCombos, exists := elements[target]
	if !exists {
		memo[target] = nil
		return nil
	}

	var shortestRecipe []string
	var bestCombo []string

	// Coba setiap kombinasi
	for _, combo := range targetCombos {
		var recipesFromCombo [][]string

		// Dapatkan resep untuk setiap elemen dalam kombinasi
		allValid := true
		for _, elem := range combo {
			elemRecipe := dfsWithMemo(elem, elements, basicElements, memo, visited, combinations)
			if len(elemRecipe) == 0 {
				allValid = false
				break
			}
			recipesFromCombo = append(recipesFromCombo, elemRecipe)
		}

		if !allValid {
			continue
		}

		// Gabungkan semua resep
		combinedRecipe := []string{}
		for _, recipe := range recipesFromCombo {
			// Tambahkan elemen yang belum ada di combinedRecipe
			for _, elem := range recipe {
				// Cek apakah elemen sudah ada di combinedRecipe
				found := false
				for _, existing := range combinedRecipe {
					if existing == elem {
						found = true
						break
					}
				}
				if !found {
					combinedRecipe = append(combinedRecipe, elem)
				}
			}
		}

		// Tambahkan target ke resep
		combinedRecipe = append(combinedRecipe, target)

		// Update shortestRecipe jika ini adalah resep pertama atau lebih pendek
		if len(shortestRecipe) == 0 || len(combinedRecipe) < len(shortestRecipe) {
			shortestRecipe = combinedRecipe
			bestCombo = combo
		}
	}

	// Simpan kombinasi terbaik yang digunakan untuk membuat target
	if len(bestCombo) > 0 {
		combinations[target] = bestCombo
	}

	// Simpan hasil ke memo
	memo[target] = shortestRecipe
	return shortestRecipe
}
