// File: main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	var recipesFile string
	var targetElement string
	var mode string
	var maxRecipes int

	// Get recipes file path
	fmt.Print("Enter recipes JSON file path (default: recipes.json): ")
	recipesFile = readInputWithDefault("../recipes.json")

	// Get target element
	for {
		fmt.Print("Enter target element to create: ")
		targetElement = readInputWithDefault("")
		if targetElement != "" {
			break
		}
		fmt.Println("Target element cannot be empty!")
	}

	// Get starting elements
	startingElements := []string{
		"Air", "Earth", "Fire", "Water", "Time",
	}

	// Get mode
	for {
		fmt.Print("Enter mode - 'single' for one recipe or 'multiple' for multiple recipes (default: single): ")
		mode = readInputWithDefault("single")
		if mode == "single" || mode == "multiple" {
			break
		}
		fmt.Println("Invalid mode! Please enter 'single' or 'multiple'")
	}

	// If multiple mode, get max recipes count
	if mode == "multiple" {
		for {
			fmt.Print("Enter maximum number of recipes to find (default: 5): ")
			maxRecipesStr := readInputWithDefault("5")
			var err error
			maxRecipes, err = strconv.Atoi(maxRecipesStr)
			if err == nil && maxRecipes > 0 {
				break
			}
			fmt.Println("Please enter a valid positive number!")
		}
	}

	// Execute based on selected mode
	if mode == "single" {
		fmt.Println("Finding a single recipe...")
		FindSingleRecipe(recipesFile, targetElement, startingElements)
	} else {
		fmt.Printf("Finding up to %d recipes...\n", maxRecipes)
		FindMultipleRecipesConcurrent(recipesFile, targetElement, startingElements, maxRecipes)
	}
}

// Helper function to read input with default value
func readInputWithDefault(defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	// Trim whitespace and newlines
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}
