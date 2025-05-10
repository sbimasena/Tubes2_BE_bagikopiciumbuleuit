package recipe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	// --- Setup HTTP Server ---
	mux := http.NewServeMux()

	// --- API Search Handler ---
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target element is required", http.StatusBadRequest)
			return
		}

		algorithm := r.URL.Query().Get("algorithm")
		if algorithm == "" {
			algorithm = "dfs"
		}
		if algorithm != "dfs" {
			http.Error(w, "Only DFS is supported in this version", http.StatusBadRequest)
			return
		}

		maxPaths := 1
		if maxPathsStr := r.URL.Query().Get("maxPaths"); maxPathsStr != "" {
			var err error
			maxPaths, err = strconv.Atoi(maxPathsStr)
			if err != nil || maxPaths < 1 {
				http.Error(w, "Invalid maxPaths value", http.StatusBadRequest)
				return
			}
		}

		recipes, err := LoadRecipes("../recipes.json")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load recipes: %v", err), http.StatusInternalServerError)
			return
		}

		startElements := []string{"Air", "Earth", "Fire", "Water"}

		paths, duration, nodes := findPathDFS(recipes, startElements, target)

		response := map[string]interface{}{
			"paths":         paths,
			"duration":      duration.String(),
			"nodes_visited": nodes,
			"algorithm":     algorithm,
		}

		writeJSON(w, response)
	})

	// --- Start Server ---
	fmt.Println("🌐 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}

// Helper function to read input with default value
// func readInputWithDefault(defaultValue string) string {
// 	reader := bufio.NewReader(os.Stdin)
// 	input, err := reader.ReadString('\n')
// 	if err != nil {
// 		log.Fatalf("Error reading input: %v", err)
// 	}

// 	// Trim whitespace and newlines
// 	input = strings.TrimSpace(input)

// 	if input == "" {
// 		return defaultValue
// 	}
// 	return input

// }

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
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
