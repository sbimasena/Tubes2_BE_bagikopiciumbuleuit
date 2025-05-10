package main

import (
	"alchemy/recipe"
	"encoding/json"
	"fmt"
	"os"
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
	// Load data from elements.json
	elements := LoadElements("recipes.json")

	// Basic elements
	startingElements := []string{"Air", "Water", "Earth", "Fire", "Time"}

	// Tanya elemen tujuan
	fmt.Print("What is the target element?: ")
	var target string
	fmt.Scanln(&target)

	if _, ok := elements[target]; !ok {
		fmt.Println("Element not found in the database.")
		return
	}

	// Pilih jenis pencarian
	fmt.Print("Single (s) or Multiple (m) recipes?: ")
	var mode string
	fmt.Scanln(&mode)

	if mode == "m" {
		fmt.Print("Maximum number of recipes to find?: ")
		var max int
		fmt.Scanln(&max)
		fmt.Println("== Multiple Recipes ==")
		recipe.FindMultipleRecipesConcurrent("recipes.json", target, startingElements, max)
	} else {
		fmt.Print("BFS or DFS?: ")
		var choice string
		fmt.Scanln(&choice)
		if choice == "b" {
			fmt.Println("== BFS ==")
			recipe.FindSingleRecipeBFS("recipes.json", target, startingElements)
		} else {
			fmt.Println("== Single Recipe ==")
			recipe.FindSingleRecipe("recipes.json", target, startingElements)
		}

	}
}
