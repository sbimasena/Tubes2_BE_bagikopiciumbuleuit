package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// ElementData represents the structure of element data from JSON
type ElementData struct {
	Element  string     `json:"element"`
	ImageURL string     `json:"image_url"`
	Recipes  [][]string `json:"recipes"`
	Tier     int        `json:"tier"`
}

// SearchResult represents the result of a search
type SearchResult struct {
	Paths        [][]string             `json:"paths"`
	Steps        []map[string][]string  `json:"steps"`
	NodesVisited int                    `json:"nodes_visited"`
	Duration     string                 `json:"duration"`
	Algorithm    string                 `json:"algorithm"`
}

// LoadElements loads elements from JSON file
func LoadElements(filename string) ([]ElementData, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var elements []ElementData
	err = json.Unmarshal(data, &elements)
	if err != nil {
		return nil, err
	}

	return elements, nil
}

// PrepareElementMaps converts element data to the format needed for search algorithms
func PrepareElementMaps(elements []ElementData) (map[string][][]string, map[string]int, map[string]bool) {
	recipeMap := make(map[string][][]string)
	tierMap := make(map[string]int)

	basicElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	for _, elem := range elements {
		recipeMap[elem.Element] = elem.Recipes
		tierMap[elem.Element] = elem.Tier
	}

	return recipeMap, tierMap, basicElements
}

// BidirectionalSearch performs a bidirectional search to find a path from basic elements to target
func BidirectionalSearch(elements []ElementData, target string, maxPaths int) (SearchResult, error) {
	startTime := time.Now()

	recipeMap, _, basicElements := PrepareElementMaps(elements)

	var paths [][]string
	var steps []map[string][]string
	var nodesVisited int

	if maxPaths > 1 {
		paths, steps, nodesVisited = BiSearchMultiple(target, recipeMap, basicElements, maxPaths)
	} else {
		path, step, nodes := BiSearch(target, recipeMap, basicElements)
		if path != nil {
			paths = [][]string{path}
			steps = []map[string][]string{step}
			nodesVisited = nodes
		}
	}

	duration := time.Since(startTime)

	result := SearchResult{
		Paths:        paths,
		Steps:        steps,
		NodesVisited: nodesVisited,
		Duration:     duration.String(),
		Algorithm:    "bidirectional",
	}

	return result, nil
}

// SearchRecipe is a unified function that can use different search algorithms based on parameters
func SearchRecipe(elementsFile string, target string, algorithm string, maxPaths int) (SearchResult, error) {
	elements, err := LoadElements(elementsFile)
	if err != nil {
		return SearchResult{}, fmt.Errorf("failed to load elements: %v", err)
	}

	switch algorithm {
	case "bfs":
		// return BreadthFirstSearch(elements, target, maxPaths)
		return SearchResult{}, fmt.Errorf("BFS implementation not connected yet")
	case "dfs":
		// return DepthFirstSearch(elements, target, maxPaths)
		return SearchResult{}, fmt.Errorf("DFS implementation not connected yet")
	case "bidirectional":
		return BidirectionalSearch(elements, target, maxPaths)
	default:
		return SearchResult{}, fmt.Errorf("unknown algorithm: %s", algorithm)
	}
}
