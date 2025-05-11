package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
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

func FindMultipleRecipesBi(target string, elements map[string][][]string, basicElements map[string]bool, algorithm string, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
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
	// Persiapkan hasil
	var paths [][]string
	var allSteps []map[string][]string
	totalNodes := 0
	var maxDuration time.Duration

	// Lacak path yang sudah ditemukan
	pathSignatures := make(map[string]bool)

	// Salin elements sekali di awal
	originalElements := copyElements(elements)

	fmt.Println("Mencari maksimal", maxPaths, "jalur berbeda untuk", target)

	// Persiapkan channel untuk membatasi jumlah pencarian yang berjalan bersamaan
	limiter := make(chan struct{}, 2) // Hanya 2 pencarian sekaligus untuk mengurangi penggunaan memori

	// Cari jalur satu per satu hingga maxPaths
	found := 0
	for attempt := 0; attempt < maxPaths*5 && found < maxPaths; attempt++ {
		limiter <- struct{}{} // Acquire semaphore

		// Salin elements dan acak urutan resep-resepnya
		elementsCopy := copyElements(originalElements)
		elementsCopy = shuffleElements(elementsCopy, attempt)

		// Modifikasi elements berdasarkan jalur yang sudah ditemukan
		if attempt > 0 && len(paths) > 0 {
			for i, path := range paths {
				if i >= 3 { // Tidak perlu memproses semua jalur untuk efisiensi
					break
				}
				tweakElementsBasedOnPath(elementsCopy, path, allSteps[i], attempt+i)
			}
		}

		// Jalankan pencarian
		p, s, n, dur := BiSearchBFS(target, elementsCopy, basicElements, tierMap)
		<-limiter // Release semaphore

		if p == nil {
			continue
		}

		// Cek apakah jalur ini unik
		signature := generateLightPathSignature(p, s)
		if pathSignatures[signature] {
			continue
		}

		// Jalur baru ditemukan
		pathSignatures[signature] = true
		paths = append(paths, p)
		allSteps = append(allSteps, s)
		totalNodes += n
		if dur > maxDuration {
			maxDuration = dur
		}
		found++

		fmt.Printf("Jalur #%d ditemukan (panjang: %d)\n", found, len(p))
	}

	fmt.Printf("\n📦 Total simpul yang dieksplorasi: %d\n", totalNodes)
	fmt.Printf("⏱ Waktu eksekusi maksimum: %v\n", maxDuration)

	if found < maxPaths {
		fmt.Printf("Hanya ditemukan %d jalur berbeda (dari %d yang diminta)\n", found, maxPaths)
	} else {
		fmt.Printf("Ditemukan %d jalur berbeda\n", found)
	}

	return paths, allSteps, totalNodes
}

func BiSearchMultipleDFS(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int, tierMap map[string]int) ([][]string, []map[string][]string, int) {
	// Persiapkan hasil
	var paths [][]string
	var allSteps []map[string][]string
	totalNodes := 0
	var maxDuration time.Duration

	// Lacak path yang sudah ditemukan
	pathSignatures := make(map[string]bool)

	// Salin elements sekali di awal
	originalElements := copyElements(elements)

	fmt.Println("Mencari maksimal", maxPaths, "jalur berbeda untuk", target)

	// Persiapkan channel untuk membatasi jumlah pencarian yang berjalan bersamaan
	limiter := make(chan struct{}, 2) // Hanya 2 pencarian sekaligus untuk mengurangi penggunaan memori

	// Cari jalur satu per satu hingga maxPaths
	found := 0
	for attempt := 0; attempt < maxPaths*5 && found < maxPaths; attempt++ {
		limiter <- struct{}{} // Acquire semaphore

		// Salin elements dan acak urutan resep-resepnya
		elementsCopy := copyElements(originalElements)
		elementsCopy = shuffleElements(elementsCopy, attempt)

		// Modifikasi elements berdasarkan jalur yang sudah ditemukan
		if attempt > 0 && len(paths) > 0 {
			for i, path := range paths {
				if i >= 3 { // Tidak perlu memproses semua jalur untuk efisiensi
					break
				}
				tweakElementsBasedOnPath(elementsCopy, path, allSteps[i], attempt+i)
			}
		}

		// Jalankan pencarian
		p, s, n, dur := BiSearchDFS(target, elementsCopy, basicElements, tierMap)
		<-limiter // Release semaphore

		if p == nil {
			continue
		}

		// Cek apakah jalur ini unik
		signature := generateLightPathSignature(p, s)
		if pathSignatures[signature] {
			continue
		}

		// Jalur baru ditemukan
		pathSignatures[signature] = true
		paths = append(paths, p)
		allSteps = append(allSteps, s)
		totalNodes += n
		if dur > maxDuration {
			maxDuration = dur
		}
		found++

		fmt.Printf("Jalur #%d ditemukan (panjang: %d)\n", found, len(p))
	}

	fmt.Printf("\n📦 Total simpul yang dieksplorasi: %d\n", totalNodes)
	fmt.Printf("⏱ Waktu eksekusi maksimum: %v\n", maxDuration)

	if found < maxPaths {
		fmt.Printf("Hanya ditemukan %d jalur berbeda (dari %d yang diminta)\n", found, maxPaths)
	} else {
		fmt.Printf("Ditemukan %d jalur berbeda\n", found)
	}

	return paths, allSteps, totalNodes
}

// Versi yang lebih ringan untuk generate signature jalur
func generateLightPathSignature(path []string, steps map[string][]string) string {
	// Hash-based signature
	h := fnv.New64a()

	// Tambahkan panjang jalur
	h.Write([]byte(fmt.Sprintf("%d:", len(path))))

	// Tambahkan setiap langkah yang bukan elemen dasar
	for _, elem := range path {
		if recipe, exists := steps[elem]; exists {
			// Urutkan ingredients untuk konsistensi
			ingredients := make([]string, len(recipe))
			copy(ingredients, recipe)
			sort.Strings(ingredients)

			// Tambahkan ke hash
			h.Write([]byte(elem + "=" + strings.Join(ingredients, "+") + ";"))
		}
	}

	return fmt.Sprintf("%x", h.Sum64())
}

// Modifikasi elements berdasarkan jalur
func tweakElementsBasedOnPath(elements map[string][][]string, path []string, steps map[string][]string, seed int) {
	// Menentukan berapa banyak elemen yang akan dimodifikasi (20-30%)
	modifyCount := 1 + (len(path)*(2+seed%3))/10
	if modifyCount > len(path)/2 {
		modifyCount = len(path) / 2
	}

	// Pilih elemen-elemen acak dari jalur (hindari elemen dasar di awal)
	startIdx := 1
	if len(path) > 5 {
		startIdx = 2
	}

	// Track elemen yang akan dimodifikasi
	elemsToModify := make(map[string]bool)

	for i := 0; i < modifyCount && startIdx < len(path); i++ {
		// Pilih posisi yang berbeda setiap kali berdasarkan seed
		idx := (startIdx + (seed+i)*7) % len(path)
		if idx < len(path) {
			elem := path[idx]
			if _, exists := steps[elem]; exists {
				elemsToModify[elem] = true
			}
		}
		startIdx += 2
	}

	// Untuk setiap elemen yang dipilih, modifikasi resepnya
	for elem := range elemsToModify {
		if recipes, exists := elements[elem]; exists && len(recipes) > 1 {
			// Pilih strategi berdasarkan seed:
			// 1. Acak urutan resep, atau
			// 2. Blokir resep yang digunakan dalam jalur

			if seed%2 == 0 {
				// Strategi 1: Acak urutan resep
				shuffledRecipes := make([][]string, len(recipes))
				copy(shuffledRecipes, recipes)

				// Fisher-Yates shuffle dengan seed
				for i := len(shuffledRecipes) - 1; i > 0; i-- {
					j := (seed * (i + 1)) % (i + 1)
					shuffledRecipes[i], shuffledRecipes[j] = shuffledRecipes[j], shuffledRecipes[i]
				}

				elements[elem] = shuffledRecipes
			} else {
				// Strategi 2: Blokir resep yang digunakan
				if recipe, exists := steps[elem]; exists {
					// Kita perlu membandingkan tanpa memperhatikan urutan
					recipeSet := make(map[string]bool)
					for _, ing := range recipe {
						recipeSet[ing] = true
					}

					// Cari resep yang cocok dan hapus jika ketemu
					for i, r := range recipes {
						// Verifikasi apakah resep sama
						if len(r) == len(recipe) {
							match := true
							for _, ing := range r {
								if !recipeSet[ing] {
									match = false
									break
								}
							}

							if match {
								// Hapus resep ini dari daftar
								elements[elem] = append(recipes[:i], recipes[i+1:]...)
								break
							}
						}
					}
				}
			}
		}
	}
}

// Fungsi untuk mengacak urutan resep dalam struktur elements
func shuffleElements(elements map[string][][]string, seed int) map[string][][]string {
	result := make(map[string][][]string)

	for elem, recipes := range elements {
		if len(recipes) <= 1 {
			// Untuk elemen dengan 0-1 resep, salin langsung
			result[elem] = recipes
			continue
		}

		// Salin recipes
		recipesCopy := make([][]string, len(recipes))
		for i, recipe := range recipes {
			recipeCopy := make([]string, len(recipe))
			copy(recipeCopy, recipe)
			recipesCopy[i] = recipeCopy
		}

		// Acak urutan resep jika ada lebih dari 1
		// Gunakan Fisher-Yates shuffle dengan seed
		for i := len(recipesCopy) - 1; i > 0; i-- {
			j := (seed * (i + 1)) % (i + 1)
			recipesCopy[i], recipesCopy[j] = recipesCopy[j], recipesCopy[i]
		}

		result[elem] = recipesCopy
	}

	return result
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

// Function to find a single recipe
func FindSingleRecipeDFS(recipesFile, targetElement string, startingElements []string) (*Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	// Cari path + info durasi dan node
	paths, duration, visited := findPathDFS(recipes, startingElements, targetElement)

	if len(paths) == 0 {
		fmt.Printf("No path found to create '%s'\n", targetElement)
		return nil, 0, 0
	}

	path := paths[0]
	fmt.Printf("Found path to create %s with %d steps:\n", targetElement, len(path.Steps))
	for i, step := range path.Steps {
		fmt.Printf("%d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}

	// Tambahan info
	fmt.Printf("⏱ Time taken to search: %v\n", duration)
	fmt.Printf("📦 Nodes visited: %d\n", visited)

	return &path, visited, duration
}

func FindMultipleRecipesDFSConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) []Path {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Printf("Error loading recipes: %v", err)
		return nil
	}

	// Build recipe maps
	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
	}

	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		log.Printf("Target element '%s' not found in recipes", targetElement)
		return nil
	}
	targetTier := targetRecipe.Tier

	// Cari kombinasi valid
	var validCombinations [][2]string
	for _, combo := range targetRecipe.Recipes {
		a, b := combo[0], combo[1]
		aTier := getTier(a, tierMap)
		bTier := getTier(b, tierMap)
		if aTier < targetTier && bTier < targetTier {
			validCombinations = append(validCombinations, combo)
		}
	}

	if len(validCombinations) == 0 {
		log.Printf("No valid recipes found for element '%s'", targetElement)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 10)

	var allPaths []Path
	pathSignatures := make(map[string]bool)

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

			// Tambahkan pengecekan context
			select {
			case <-ctx.Done():
				return
			default:
			}

			var combinedSteps []Step
			visitedSet := make(map[[3]string]bool)

			for _, ingredient := range combo {
				if isBasicElement(ingredient, startingElements) {
					continue
				}
				paths, _, _ := findPathDFS(recipes, startingElements, ingredient)
				if len(paths) == 0 {
					return
				}
				for _, s := range paths[0].Steps {
					key := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
					if !visitedSet[key] {
						visitedSet[key] = true
						combinedSteps = append(combinedSteps, s)
					}
				}
			}

			// Tambahkan step akhir
			finalStep := Step{Ingredients: combo, Result: targetElement}
			key := [3]string{combo[0], combo[1], targetElement}
			if !visitedSet[key] {
				combinedSteps = append(combinedSteps, finalStep)
			}

			path := Path{Steps: combinedSteps, FinalItem: targetElement}
			signature := generateSimpleSignature(path)

			mu.Lock()
			defer mu.Unlock()
			if !pathSignatures[signature] && len(allPaths) < maxRecipes {
				pathSignatures[signature] = true
				allPaths = append(allPaths, path)
				if len(allPaths) >= maxRecipes {
					cancel()
				}
			}
		}(combo)
	}

	wg.Wait()

	// Sort hasilnya berdasarkan jumlah langkah
	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	return allPaths
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

func FindSingleRecipeBFS(recipesFile, targetElement string, startingElements []string) (*Path, int, time.Duration) {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
		return nil, 0, 0
	}

	fmt.Printf("Loaded %d recipes\n", len(recipes))
	fmt.Printf("Finding path to create: %s\n", targetElement)

	// Cari path + info durasi dan node
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

	// Tambahan info
	fmt.Printf("⏱ Time taken to search: %v\n", duration)
	fmt.Printf("📦 Nodes visited: %d\n", visited)

	return &path, visited, duration
}

func FindMultipleRecipesBFSConcurrent(recipesFile, targetElement string, startingElements []string, maxRecipes int) []Path {
	recipes, err := LoadRecipes(recipesFile)
	if err != nil {
		log.Printf("Error loading recipes: %v", err)
		return nil
	}

	// Build maps
	elementMap := make(map[string]ElementRecipe)
	tierMap := make(map[string]int)
	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
	}

	targetRecipe, exists := elementMap[targetElement]
	if !exists {
		log.Printf("Target element '%s' not found in recipes", targetElement)
		return nil
	}
	targetTier := targetRecipe.Tier

	var validCombinations [][2]string
	for _, combo := range targetRecipe.Recipes {
		a, b := combo[0], combo[1]
		aTier := getTier(a, tierMap)
		bTier := getTier(b, tierMap)
		if aTier < targetTier && bTier < targetTier {
			validCombinations = append(validCombinations, combo)
		}
	}
	if len(validCombinations) == 0 {
		log.Printf("No valid recipes found for element '%s'", targetElement)
		return nil
	}

	// Prepare concurrency
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 10)

	var allPaths []Path
	pathSignatures := make(map[string]bool)

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
			for _, ingredient := range combo {
				if isBasicElement(ingredient, startingElements) {
					continue
				}
				paths, _, _ := findPathBFS(recipes, startingElements, ingredient)
				if len(paths) == 0 {
					return
				}
				if !isValidPath(paths[0], tierMap) {
					return
				}
				combinedPaths = append(combinedPaths, paths[0])
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

			finalPath := Path{
				Steps:     steps,
				FinalItem: targetElement,
			}
			if !isValidPath(finalPath, tierMap) {
				return
			}

			signature := generateSimpleSignature(finalPath)

			mu.Lock()
			if !pathSignatures[signature] && len(allPaths) < maxRecipes {
				pathSignatures[signature] = true
				allPaths = append(allPaths, finalPath)
				if len(allPaths) >= maxRecipes {
					cancel()
				}
			}
			mu.Unlock()
		}(combo)
	}

	wg.Wait()

	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i].Steps) < len(allPaths[j].Steps)
	})

	return allPaths
}
