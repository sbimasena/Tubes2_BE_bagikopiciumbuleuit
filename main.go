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
	recipe.VisualEnabled = false

	// Prepare visualization folder
	ClearFramesFolder()

	// Load data from elements.json
	elements := LoadElements("elements.json")

	// Basic elements
	basicElements := map[string]bool{
		"Air": true, "Water": true, "Earth": true, "Fire": true, "Time": true,
	}
	fmt.Print("What is the target element?: ")
	var target string
	fmt.Scanln(&target)
	if _, ok := elements[target]; !ok {
		fmt.Println("Element not found in the database.")
		return
	}
	fmt.Print("BFS or DFS? (b/d): ")
	var choice string
	fmt.Scanln(&choice)
	if choice != "b" && choice != "d" {
		fmt.Println("Invalid choice. Defaulting to BFS.")
		choice = "b"
	}
	// Set the search algorithm based on user choice
	if choice == "b" {
		// Shortest Path
		fmt.Println("== Shortest Recipe ==")
		shortest := recipe.FindShortestRecipe(target, elements, basicElements, false)
		if shortest == nil {
			fmt.Println("No recipe found")
		}
	} else {
		// Shortest Path
		fmt.Println("== Shortest Recipe ==")
		shortest := recipe.FindShortestRecipe(target, elements, basicElements, true)
		if shortest == nil {
			fmt.Println("No recipe found")
		}
	}
	// Multiple Paths
	// fmt.Println("\n== Multiple Recipes (max 3) ==")
	// fmt.Println("Finding multiple recipes for Brick...")
	// multiple := recipe.FindMultipleRecipesConcurrent("Brick", elements, basicElements, 3)
	// if len(multiple) == 0 {
	// 	fmt.Println("No recipes found")
	// }

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
