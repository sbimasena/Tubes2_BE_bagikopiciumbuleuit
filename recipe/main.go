package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	mux := http.NewServeMux()

	// Endpoint: /api/search
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target element is required", http.StatusBadRequest)
			return
		}

		algorithm := r.URL.Query().Get("algorithm")
		if algorithm == "" {
			algorithm = "bfs"
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

		result, err := SearchRecipe("recipes.json", target, algorithm, maxPaths)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSON(w, result)
	})

	// Endpoint: /api/bidirectional
	mux.HandleFunc("/api/bidirectional", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target element is required", http.StatusBadRequest)
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

		result, err := SearchRecipe("recipes.json", target, "bidirectional", maxPaths)

		if err != nil {
			http.Error(w, fmt.Sprintf("Bidirectional search failed: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSON(w, result)
	})

	// Serve static files if needed
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// Start server with CORS middleware
	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, withCORS(mux)))
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

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
