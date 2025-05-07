package recipe

import (
	"fmt"
	"strings"
	"sync"
)

// FindShortestRecipe finds the shortest recipe path to a target element
func FindShortestRecipe(target string, elements map[string][][]string, basicElements map[string]bool) []string {
	fmt.Println("Finding shortest recipe for", target)

	// Jika target sudah elemen dasar, kembalikan langsung
	if basicElements[target] {
		fmt.Printf("Found recipe with 1 step\n")
		recipe := []string{target}
		// Tambahkan output yang diinginkan
		fmt.Printf("%s (basic element)\n", target)
		TraceLive(recipe, elements, basicElements)
		return recipe
	}

	// Memoization untuk menyimpan resep terpendek yang sudah ditemukan
	memo := make(map[string][]string)
	visited := make(map[string]bool)

	// Juga simpan informasi tentang kombinasi yang digunakan
	combinations := make(map[string][]string)

	// Melakukan DFS
	recipe := dfsWithMemo(target, elements, basicElements, memo, visited, combinations)

	if len(recipe) == 0 {
		fmt.Println("No recipe found")
		return nil
	}

	fmt.Printf("Found recipe with %d steps\n", len(recipe))

	// Tampilkan resep dalam format yang diinginkan
	printFormattedRecipe(recipe, combinations)

	TraceLive(recipe, elements, basicElements)
	return recipe
}

func FindMultipleRecipesConcurrent(target string, elements map[string][][]string, basicElements map[string]bool, maxRecipes int) [][]string {
	fmt.Printf("Finding up to %d recipes for %s\n", maxRecipes, target)

	var result [][]string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxRecipes)

	recipes := elements[target]
	if len(recipes) == 0 {
		return nil
	}

	for _, recipe := range recipes {
		sem <- struct{}{}
		wg.Add(1)

		go func(recipe []string) {
			defer wg.Done()
			defer func() { <-sem }()

			memo := make(map[string][]string)
			combinations := make(map[string][]string)
			visited := make(map[string]bool)

			fullPath := []string{}
			success := true

			for _, ing := range recipe {
				path := dfsWithMemo(ing, elements, basicElements, memo, visited, combinations)
				if path == nil {
					success = false
					break
				}
				fullPath = append(fullPath, path...)
			}

			if success {
				fullPath = append(fullPath, target)

				mu.Lock()
				if len(result) < maxRecipes {
					result = append(result, fullPath)
					fmt.Printf("\nRecipe %d (%d steps):\n", len(result), len(fullPath))
					printFormattedRecipe(fullPath, combinations)
				}
				mu.Unlock()
			}
		}(recipe)

		mu.Lock()
		if len(result) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}

	wg.Wait()
	return result
}

func printFormattedRecipe(recipe []string, combinations map[string][]string) {
	fmt.Println("Recipe steps:")
	used := make(map[string]bool)

	for _, elem := range recipe {
		if combo, found := combinations[elem]; found {
			// hanya print kalau semua bahan combo-nya udah ada di used
			ready := true
			for _, part := range combo {
				if !used[part] {
					ready = false
					break
				}
			}
			if ready {
				fmt.Printf("* %s -> %s\n", strings.Join(combo, " + "), elem)
				used[elem] = true
			}
		} else {
			fmt.Printf("* %s (basic element)\n", elem)
			used[elem] = true
		}
	}

	fmt.Printf("Found recipe with %d steps: %v\n", len(recipe), recipe)
}
