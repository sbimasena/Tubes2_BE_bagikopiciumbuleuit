package main

import (
	"alchemy/recipe"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

// func isElementInRecipes(target string, elements []recipe.ElementData) bool {
// 	for _, element := range elements {
// 		if element.Element == target {
// 			return true
// 		}
// 	}

// 	return false
// }

// func main() {
// 	reader := bufio.NewReader(os.Stdin)
// 	log.Print("Scraping atau tidak? (y/n): ")
// 	scrape, _ := reader.ReadString('\n')
// 	scrape = strings.TrimSpace(strings.ToLower(scrape))
// 	if scrape == "y" {
// 		log.Println("Melakukan scraping...")
// 		mainScrap()
// 	}
// 	log.Print("Pilih algoritma utama (bfs/dfs/bidirectional): ")
// 	mainAlg, _ := reader.ReadString('\n')
// 	mainAlg = strings.TrimSpace(strings.ToLower(mainAlg))

// 	var bidiAlg string
// 	if mainAlg == "bidirectional" {
// 		log.Print("Pilih metode bidirectional (bfs/dfs): ")
// 		bidiAlgRaw, _ := reader.ReadString('\n')
// 		bidiAlg = strings.TrimSpace(strings.ToLower(bidiAlgRaw))
// 		if bidiAlg != "bfs" && bidiAlg != "dfs" {
// 			log.Println("Metode bidirectional tidak valid, gunakan bfs atau dfs.")
// 			return
// 		}
// 	}

// 	log.Print("Ingin mencari satu resep atau banyak? (1/multiple): ")
// 	mode, _ := reader.ReadString('\n')
// 	mode = strings.TrimSpace(strings.ToLower(mode))

// 	maxPaths := 1
// 	if mode == "multiple" {
// 		log.Print("Berapa jumlah maksimum resep yang ingin dicari? ")
// 		input, _ := reader.ReadString('\n')
// 		input = strings.TrimSpace(input)
// 		val, err := strconv.Atoi(input)
// 		if err == nil && val > 0 {
// 			maxPaths = val
// 		}
// 	}

// 	elements, err := recipe.LoadElements("recipes.json")
// 	if err != nil {
// 		log.Println("Gagal membaca file recipes.json:", err)
// 		return
// 	}

// 	log.Print("Masukkan nama elemen target: ")
// 	target, _ := reader.ReadString('\n')
// 	target = strings.TrimSpace(target)

// 	if target == "" || target == " " || target == "\n" {
// 		log.Println("Nama elemen target tidak boleh kosong.")
// 		return
// 	}

// 	if !isElementInRecipes(target, elements) {
// 		log.Println("Elemen target tidak ditemukan dalam database.")
// 		return
// 	}

// 	recipeMap, tierMap, basicElements := recipe.PrepareElementMaps(elements)

// 	var (
// 		paths [][]string
// 		steps []map[string][]string
// 	)

// 	switch mainAlg {
// 	case "dfs":
// 		if mode == "multiple" {
// 			recipe.FindMultipleRecipesConcurrent("recipes.json", target, keys(basicElements), maxPaths)
// 			return
// 		} else {
// 			recipe.FindSingleRecipeDFS("recipes.json", target, keys(basicElements))
// 			return
// 		}
// 	case "bidirectional":
// 		if bidiAlg == "dfs" {
// 			if mode == "multiple" {
// 				paths, steps, _ = recipe.FindMultipleRecipes(target, recipeMap, basicElements, "dfs", maxPaths, tierMap)
// 			} else {
// 				path, step, visited, dur := recipe.FindSingleRecipe(target, recipeMap, basicElements, "dfs", tierMap)
// 				if path != nil {
// 					paths = append(paths, path)
// 					steps = append(steps, step)
// 					log.Println("\nTotal simpul yang dieksplorasi:", visited)
// 					log.Println("Waktu eksekusi:", dur)
// 				}
// 			}
// 		} else {
// 			if mode == "multiple" {
// 				paths, steps, _ = recipe.FindMultipleRecipes(target, recipeMap, basicElements, "bfs", maxPaths, tierMap)
// 			} else {
// 				path, step, visited, dur := recipe.FindSingleRecipe(target, recipeMap, basicElements, "bfs", tierMap)
// 				if path != nil {
// 					paths = append(paths, path)
// 					steps = append(steps, step)
// 					log.Println("\nTotal simpul yang dieksplorasi:", visited)
// 					log.Println("Waktu eksekusi:", dur)
// 				}
// 			}
// 		}
// 	default: // bfs
// 		if mode == "multiple" {
// 			recipe.FindMultipleRecipesBFSConcurrent("recipes.json", target, keys(basicElements), maxPaths)
// 		} else {
// 			recipe.FindSingleRecipeBFS("recipes.json", target, keys(basicElements))
// 			return
// 		}
// 	}

// 	log.Println("\nHasil:")
// 	log.Printf("Ditemukan %d jalur resep.\n", len(paths))

// 	for i := range paths {
// 		stepMap := steps[i]
// 		log.Printf("\nResep ke-%d:\n", i+1)
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
// 			log.Printf("%d. %s + %s = %s\n", counter, ing[0], ing[1], res)
// 			counter++
// 			printed[res] = true
// 		}
// 		printSteps(target)
// 	}
// }

// func keys(m map[string]bool) []string {
// 	var out []string
// 	for k := range m {
// 		out = append(out, k)
// 	}
// 	return out
// }

func main() {
	mux := http.NewServeMux()

	// 🔍 SEARCH HANDLER
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

		var result recipe.SearchResult

		bidi := r.URL.Query().Get("bidi")
		elements, err := recipe.LoadElements("recipes.json")
		if err != nil {
			http.Error(w, "Failed to load recipe elements", http.StatusInternalServerError)
			return
		}
		recipeMap, tierMap, basicElements := recipe.PrepareElementMaps(elements)

		switch algorithm {
		case "dfs":
			if maxPaths > 1 {
				paths, visited, duration := recipe.FindMultipleRecipesDFSConcurrent("recipes.json", target, startingElements, maxPaths)
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

				result = recipe.SearchResult{
					Paths:        converted,
					Steps:        stepsList,
					NodesVisited: visited,
					Duration:     duration.String(),
					Algorithm:    "dfs",
				}
			} else {
				path, visited, duration := recipe.FindSingleRecipeDFS("recipes.json", target, startingElements)
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
				result = recipe.SearchResult{
					Paths:        [][]string{pathList},
					Steps:        []map[string][]string{stepMap},
					NodesVisited: visited,
					Duration:     duration.String(),
					Algorithm:    "dfs",
				}
			}
		case "bfs":
			if maxPaths > 1 {
				paths, visited, duration := recipe.FindMultipleRecipesBFSConcurrent("recipes.json", target, startingElements, maxPaths)
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

				result = recipe.SearchResult{
					Paths:        converted,
					Steps:        stepsList,
					NodesVisited: visited,
					Duration:     duration.String(),
					Algorithm:    "bfs",
				}
			} else {
				path, visited, duration := recipe.FindSingleRecipeBFS("recipes.json", target, startingElements)
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
				result = recipe.SearchResult{
					Paths:        [][]string{pathList},
					Steps:        []map[string][]string{stepMap},
					NodesVisited: visited,
					Duration:     duration.String(),
					Algorithm:    "bfs",
				}
			}
		case "bidirectional":
			if bidi != "bfs" && bidi != "dfs" {
				http.Error(w, "Invalid bidi parameter (must be bfs or dfs)", http.StatusBadRequest)
				return
			}

			if maxPaths > 1 {
				paths, steps, totalNodes, duration := recipe.FindMultipleRecipesBi(target, recipeMap, basicElements, bidi, maxPaths, tierMap)
				if len(paths) == 0 {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				result = recipe.SearchResult{
					Paths:        paths,
					Steps:        steps,
					NodesVisited: totalNodes,
					Duration:     duration.String(),
					Algorithm:    "bidirectional-" + bidi,
				}
			} else {
				path, step, visited, duration := recipe.FindSingleRecipeBi(target, recipeMap, basicElements, bidi, tierMap)
				if path == nil {
					http.Error(w, "No path found", http.StatusNotFound)
					return
				}

				result = recipe.SearchResult{
					Paths:        [][]string{path},
					Steps:        []map[string][]string{step},
					NodesVisited: visited,
					Duration:     duration.String(),
					Algorithm:    "bidirectional-" + bidi,
				}
			}
		default:
			http.Error(w, "Unknown algorithm", http.StatusBadRequest)
			return
		}

		writeJSON(w, result)
	})

	// 🧲 SCRAPING HANDLER
	mux.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Scraping triggered via API...")
		if err := mainScrap(); err != nil {
			http.Error(w, "Scraping failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Scraping completed successfully"))
	})

	mux.HandleFunc("/api/elements", func(w http.ResponseWriter, r *http.Request) {
		elements, err := recipe.LoadElements("recipes.json")
		if err != nil {
			http.Error(w, "Failed to load elements", http.StatusInternalServerError)
			return
		}

		type ElementImage struct {
			Element  string `json:"element"`
			ImageURL string `json:"image_url"`
		}

		var result []ElementImage
		for _, e := range elements {
			result = append(result, ElementImage{
				Element:  e.Element,
				ImageURL: e.ImageURL,
			})
		}

		writeJSON(w, result)
	})

	mux.HandleFunc("/api/image", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "Missing image URL", http.StatusBadRequest)
			return
		}

		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "Failed to fetch image", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		io.Copy(w, resp.Body)
	})

	log.Println("🌐 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
