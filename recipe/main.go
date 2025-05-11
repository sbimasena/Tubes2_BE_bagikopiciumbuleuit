package main

import (
	// "bufio"
	"fmt"
	// "os"
	"strconv"
	// "strings"
	"encoding/json"
	"log"
	"net/http"
)

// func main() {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Print("Pilih algoritma utama (bfs/dfs/bidirectional): ")
// 	mainAlg, _ := reader.ReadString('\n')
// 	mainAlg = strings.TrimSpace(strings.ToLower(mainAlg))

// 	var bidiAlg string
// 	if mainAlg == "bidirectional" {
// 		fmt.Print("Pilih metode bidirectional (bfs/dfs): ")
// 		bidiAlgRaw, _ := reader.ReadString('\n')
// 		bidiAlg = strings.TrimSpace(strings.ToLower(bidiAlgRaw))
// 		if bidiAlg != "bfs" && bidiAlg != "dfs" {
// 			fmt.Println("Metode bidirectional tidak valid, gunakan bfs atau dfs.")
// 			return
// 		}
// 	}

// 	fmt.Print("Ingin mencari satu resep atau banyak? (1/multiple): ")
// 	mode, _ := reader.ReadString('\n')
// 	mode = strings.TrimSpace(strings.ToLower(mode))

// 	maxPaths := 1
// 	if mode == "multiple" {
// 		fmt.Print("Berapa jumlah maksimum resep yang ingin dicari? ")
// 		input, _ := reader.ReadString('\n')
// 		input = strings.TrimSpace(input)
// 		val, err := strconv.Atoi(input)
// 		if err == nil && val > 0 {
// 			maxPaths = val
// 		}
// 	}

// 	fmt.Print("Masukkan nama elemen target: ")
// 	target, _ := reader.ReadString('\n')
// 	target = strings.TrimSpace(target)

// 	elements, err := LoadElements("../recipes.json")
// 	if err != nil {
// 		fmt.Println("Gagal membaca file recipes.json:", err)
// 		return
// 	}

// 	recipeMap, tierMap, basicElements := PrepareElementMaps(elements)

// 	var (
// 		paths [][]string
// 		steps []map[string][]string
// 	)

// 	switch mainAlg {
// 	case "dfs":
// 		if mode == "multiple" {
// 			FindMultipleRecipesConcurrent("../recipes.json", target, keys(basicElements), maxPaths)
// 			return
// 		} else {
// 			FindSingleRecipeDFS("../recipes.json", target, keys(basicElements))
// 			return
// 		}
// 	case "bidirectional":
// 		if bidiAlg == "dfs" {
// 			if mode == "multiple" {
// 				paths, steps, _ = FindMultipleRecipes(target, recipeMap, basicElements, "dfs", maxPaths, tierMap)
// 			} else {
// 				path, step, visited, dur := FindSingleRecipe(target, recipeMap, basicElements, "dfs", tierMap)
// 				if path != nil {
// 					paths = append(paths, path)
// 					steps = append(steps, step)
// 					fmt.Println("\nTotal simpul yang dieksplorasi:", visited)
// 					fmt.Println("Waktu eksekusi:", dur)
// 				}
// 			}
// 		} else {
// 			if mode == "multiple" {
// 				paths, steps, _ = FindMultipleRecipes(target, recipeMap, basicElements, "bfs", maxPaths, tierMap)
// 			} else {
// 				path, step, visited, dur := FindSingleRecipe(target, recipeMap, basicElements, "bfs", tierMap)
// 				if path != nil {
// 					paths = append(paths, path)
// 					steps = append(steps, step)
// 					fmt.Println("\nTotal simpul yang dieksplorasi:", visited)
// 					fmt.Println("Waktu eksekusi:", dur)
// 				}
// 			}
// 		}
// 	default: // bfs
// 		if mode == "multiple" {
// 			FindMultipleRecipesBFSConcurrent("../recipes.json", target, keys(basicElements), maxPaths)
// 		} else {
// 			FindSingleRecipeBFS("../recipes.json", target, keys(basicElements))
// 			return
// 		}
// 	}

// 	fmt.Println("\nHasil:")
// 	fmt.Printf("Ditemukan %d jalur resep.\n", len(paths))

// 	for i := range paths {
// 		stepMap := steps[i]
// 		fmt.Printf("\nResep ke-%d:\n", i+1)
// 		counter := 1
// 		printed := make(map[string]bool)

// 		var printSteps func(res string)
// 		printSteps = func(res string) {
// 			if printed[res] {
// 				return
// 			}
// 			ing, ok := stepMap[res]
// 			if !ok {
// 				return
// 			}
// 			printSteps(ing[0])
// 			printSteps(ing[1])
// 			fmt.Printf("%d. %s + %s = %s\n", counter, ing[0], ing[1], res)
// 			counter++
// 			printed[res] = true
// 		}
// 		printSteps(target)
// 	}
// }

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Missing target", http.StatusBadRequest)
			return
		}

		algorithm := r.URL.Query().Get("algorithm")
		if algorithm == "" {
			algorithm = "dfs"
		}

		maxPaths := 1
		if mp := r.URL.Query().Get("maxPaths"); mp != "" {
			if val, err := strconv.Atoi(mp); err == nil && val > 0 {
				maxPaths = val
			} else {
				http.Error(w, "Invalid maxPaths", http.StatusBadRequest)
				return
			}
		}

		startingElements := []string{"Air", "Earth", "Fire", "Water"}

		var result SearchResult

		if algorithm == "dfs" {
			if maxPaths > 1 {
				// MULTI DFS
				paths := FindMultipleRecipesDFSConcurrent("../recipes.json", target, startingElements, maxPaths)
				if len(paths) == 0 {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				var converted [][]string
				var stepsList []map[string][]string
				for _, p := range paths {
					var pathSteps []string
					stepMap := make(map[string][]string)
					for _, s := range p.Steps {
						pathSteps = append(pathSteps, s.Result)
						stepMap[s.Result] = []string{s.Ingredients[0], s.Ingredients[1]}
					}
					converted = append(converted, pathSteps)
					stepsList = append(stepsList, stepMap)
				}

				result = SearchResult{
					Paths:        converted,
					Steps:        stepsList,
					NodesVisited: -1, // You can add tracking if needed
					Algorithm:    "dfs",
				}
			} else {
				path, visited, duration := FindSingleRecipeDFS("../recipes.json", target, startingElements)
				if path == nil {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				stepMap := make(map[string][]string)
				var pathList []string
				for _, s := range path.Steps {
					stepMap[s.Result] = []string{s.Ingredients[0], s.Ingredients[1]}
					pathList = append(pathList, s.Result)
				}

				result = SearchResult{
					Paths:        [][]string{pathList},
					Steps:        []map[string][]string{stepMap},
					NodesVisited: visited,           // ← tambahkan field ini di struct Path
					Duration:     duration.String(), // ← tambahkan juga
					Algorithm:    "bfs",
				}
			}
		} else if algorithm == "bfs" {
			if maxPaths > 1 {
				// MULTI BFS
				paths := FindMultipleRecipesBFSConcurrent("../recipes.json", target, startingElements, maxPaths)
				if len(paths) == 0 {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				var converted [][]string
				var stepsList []map[string][]string
				for _, p := range paths {
					var pathSteps []string
					stepMap := make(map[string][]string)
					for _, s := range p.Steps {
						pathSteps = append(pathSteps, s.Result)
						stepMap[s.Result] = []string{s.Ingredients[0], s.Ingredients[1]}
					}
					converted = append(converted, pathSteps)
					stepsList = append(stepsList, stepMap)
				}

				result = SearchResult{
					Paths:        converted,
					Steps:        stepsList,
					NodesVisited: -1,
					Algorithm:    "bfs",
				}
			} else {
				path, visited, duration := FindSingleRecipeBFS("../recipes.json", target, startingElements)
				if path == nil {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				stepMap := make(map[string][]string)
				var pathList []string
				for _, s := range path.Steps {
					stepMap[s.Result] = []string{s.Ingredients[0], s.Ingredients[1]}
					pathList = append(pathList, s.Result)
				}

				result = SearchResult{
					Paths:        [][]string{pathList},
					Steps:        []map[string][]string{stepMap},
					NodesVisited: visited,           // ← tambahkan field ini di struct Path
					Duration:     duration.String(), // ← tambahkan juga
					Algorithm:    "bfs",
				}
			}
		} else {
			http.Error(w, "Unknown algorithm", http.StatusBadRequest)
			return
		}

		writeJSON(w, result)
	})

	fmt.Println("🌐 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
