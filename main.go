package main

import (
	"alchemy/recipe"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Element struct {
	Element string     `json:"element"`
	Recipes [][]string `json:"recipes"`
}

func LoadElements(path string) map[string][][]string {
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var raw []Element
	if err := json.Unmarshal(file, &raw); err != nil {
		panic(err)
	}
	result := make(map[string][][]string)
	for _, e := range raw {
		result[e.Element] = e.Recipes
	}
	return result
}

func ClearFramesFolder() {
	_ = os.Mkdir("frames", 0755)
	files, err := filepath.Glob("frames/*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		_ = os.Remove(f)
	}
}

func main() {
	// Set to true to enable step-by-step visualization, false for better performance
	recipe.VisualEnabled = true

	// Prepare visualization folder
	ClearFramesFolder()

	// Load data from elements.json
	elements := LoadElements("elements.json")

	// Basic elements
	basicElements := map[string]bool{
		"Air": true, "Water": true, "Earth": true, "Fire": true, "Time": true,
	}

	// Shortest Path
	fmt.Println("== Shortest Recipe ==")
	shortest := recipe.FindShortestRecipe("Rain", elements, basicElements)
	if shortest != nil {
		fmt.Printf("Found recipe with %d steps: %v\n", len(shortest), shortest)
	} else {
		fmt.Println("No recipe found")
	}

	// Multiple Paths
	fmt.Println("\n== Multiple Recipes (max 3) ==")
	fmt.Println("Finding multiple recipes for Rain...")
	multiple := recipe.FindMultipleRecipesConcurrent("Rain", elements, basicElements, 3)
	if len(multiple) > 0 {
		for i, path := range multiple {
			fmt.Printf("Recipe %d (%d steps): %v\n", i+1, len(path), path)
		}
	} else {
		fmt.Println("No recipes found")
	}

	if recipe.VisualEnabled {
		fmt.Println("\nVisualization files saved to the 'frames' folder")

		// Convert DOT files to PNG using Graphviz
		fmt.Println("Converting DOT files to PNG...")
		files, err := filepath.Glob("frames/*.dot")
		if err == nil {
			for _, file := range files {
				outPng := file[:len(file)-4] + ".png"
				cmd := exec.Command("dot", "-Tpng", file, "-o", outPng)
				err := cmd.Run()
				if err != nil {
					fmt.Printf("Error converting %s: %v\n", file, err)
				}
			}
		}

		fmt.Println("Done! You can view the step-by-step visualizations as PNG files")
	}
}
