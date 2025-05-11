package main

import (
	"alchemy/recipe"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func isElementInRecipes(target string, elements []recipe.ElementData) bool {
	for _, element := range elements {
		if element.Element == target {
			return true
		}
	}

	return false
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Scraping atau tidak? (y/n): ")
	scrape, _ := reader.ReadString('\n')
	scrape = strings.TrimSpace(strings.ToLower(scrape))
	if scrape == "y" {
		fmt.Println("Melakukan scraping...")
		mainScrap()
	}
	fmt.Print("Pilih algoritma utama (bfs/dfs/bidirectional): ")
	mainAlg, _ := reader.ReadString('\n')
	mainAlg = strings.TrimSpace(strings.ToLower(mainAlg))

	var bidiAlg string
	if mainAlg == "bidirectional" {
		fmt.Print("Pilih metode bidirectional (bfs/dfs): ")
		bidiAlgRaw, _ := reader.ReadString('\n')
		bidiAlg = strings.TrimSpace(strings.ToLower(bidiAlgRaw))
		if bidiAlg != "bfs" && bidiAlg != "dfs" {
			fmt.Println("Metode bidirectional tidak valid, gunakan bfs atau dfs.")
			return
		}
	}

	fmt.Print("Ingin mencari satu resep atau banyak? (1/multiple): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(strings.ToLower(mode))

	maxPaths := 1
	if mode == "multiple" {
		fmt.Print("Berapa jumlah maksimum resep yang ingin dicari? ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		val, err := strconv.Atoi(input)
		if err == nil && val > 0 {
			maxPaths = val
		}
	}

	elements, err := recipe.LoadElements("recipes.json")
	if err != nil {
		fmt.Println("Gagal membaca file recipes.json:", err)
		return
	}

	fmt.Print("Masukkan nama elemen target: ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	if target == "" || target == " " || target == "\n" {
		fmt.Println("Nama elemen target tidak boleh kosong.")
		return
	}

	if !isElementInRecipes(target, elements) {
		fmt.Println("Elemen target tidak ditemukan dalam database.")
		return
	}

	recipeMap, tierMap, basicElements := recipe.PrepareElementMaps(elements)

	var (
		paths [][]string
		steps []map[string][]string
	)

	switch mainAlg {
	case "dfs":
		if mode == "multiple" {
			recipe.FindMultipleRecipesConcurrent("recipes.json", target, keys(basicElements), maxPaths)
			return
		} else {
			recipe.FindSingleRecipeDFS("recipes.json", target, keys(basicElements))
			return
		}
	case "bidirectional":
		if bidiAlg == "dfs" {
			if mode == "multiple" {
				paths, steps, _ = recipe.FindMultipleRecipes(target, recipeMap, basicElements, "dfs", maxPaths, tierMap)
			} else {
				path, step, visited, dur := recipe.FindSingleRecipe(target, recipeMap, basicElements, "dfs", tierMap)
				if path != nil {
					paths = append(paths, path)
					steps = append(steps, step)
					fmt.Println("\nTotal simpul yang dieksplorasi:", visited)
					fmt.Println("Waktu eksekusi:", dur)
				}
			}
		} else {
			if mode == "multiple" {
				paths, steps, _ = recipe.FindMultipleRecipes(target, recipeMap, basicElements, "bfs", maxPaths, tierMap)
			} else {
				path, step, visited, dur := recipe.FindSingleRecipe(target, recipeMap, basicElements, "bfs", tierMap)
				if path != nil {
					paths = append(paths, path)
					steps = append(steps, step)
					fmt.Println("\nTotal simpul yang dieksplorasi:", visited)
					fmt.Println("Waktu eksekusi:", dur)
				}
			}
		}
	default: // bfs
		if mode == "multiple" {
			recipe.FindMultipleRecipesBFSConcurrent("recipes.json", target, keys(basicElements), maxPaths)
		} else {
			recipe.FindSingleRecipeBFS("recipes.json", target, keys(basicElements))
			return
		}
	}

	fmt.Println("\nHasil:")
	fmt.Printf("Ditemukan %d jalur resep.\n", len(paths))

	for i := range paths {
		stepMap := steps[i]
		fmt.Printf("\nResep ke-%d:\n", i+1)
		counter := 1
		printed := make(map[string]bool)

		var printSteps func(res string)
		printSteps = func(res string) {
			if printed[res] {
				return
			}
			ing, ok := stepMap[res]
			if !ok {
				return
			}
			printSteps(ing[0])
			printSteps(ing[1])
			fmt.Printf("%d. %s + %s = %s\n", counter, ing[0], ing[1], res)
			counter++
			printed[res] = true
		}
		printSteps(target)
	}
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
