package recipe

import (
	"time"
)

type BFSNode struct {
	Remaining []string           // Elements still to be resolved
	Steps     []Step             // Steps collected so far
	Visited   map[string]bool    // Track visited elements in this path
	StepSet   map[[3]string]bool // Deduplicate steps by ingredients and result
}

func findPathBFS(recipes []ElementRecipe, startElements []string, target string) ([]Path, time.Duration, int) {
	startTime := time.Now()
	iterations := 0

	// Build lookup maps
	tierMap := make(map[string]int)
	recipeMap := make(map[string][][2]string)

	for _, recipe := range recipes {
		tierMap[recipe.Element] = recipe.Tier
		recipeMap[recipe.Element] = recipe.Recipes
		for _, combo := range recipe.Recipes {
			for _, ing := range combo {
				if _, exists := tierMap[ing]; !exists {
					tierMap[ing] = 1 // default for undefined tier
				}
			}
		}
	}

	// Set tier for basic elements
	for _, elem := range startElements {
		if _, exists := tierMap[elem]; !exists {
			tierMap[elem] = 1
		}
	}

	// Identify basic elements
	basics := make(map[string]bool)
	for _, e := range startElements {
		basics[e] = true
	}

	visitedCounter := make(map[string]bool)

	queue := []BFSNode{
		{
			Remaining: []string{target},
			Steps:     []Step{},
			Visited:   make(map[string]bool),
			StepSet:   make(map[[3]string]bool),
		},
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// First, check if all elements in `Remaining` are basic
		allBasic := true
		var elemToExpand string

		for _, elem := range curr.Remaining {
			visitedCounter[elem] = true
			if !basics[elem] {
				allBasic = false
				elemToExpand = elem
			}
		}

		// fmt.Printf("Exploring: %v, Steps: %d\n", curr.Remaining, len(curr.Steps))
		if allBasic {
			reversedSteps := make([]Step, len(curr.Steps))
			for i, step := range curr.Steps {
				reversedSteps[len(curr.Steps)-1-i] = step
			}
			return []Path{{
				Steps:     reversedSteps,
				FinalItem: target,
			}}, time.Since(startTime), iterations
		}

		// Skip cycles
		if curr.Visited[elemToExpand] {
			newRemaining := removeElement(curr.Remaining, elemToExpand)
			queue = append(queue, BFSNode{
				Remaining: newRemaining,
				Steps:     curr.Steps,
				Visited:   curr.Visited,
				StepSet:   curr.StepSet,
			})
			continue
		}

		newVisited := copyVisitedMap(curr.Visited)
		newVisited[elemToExpand] = true

		combos, ok := recipeMap[elemToExpand]
		if !ok {
			continue
		}

		for _, combo := range combos {
			iterations++
			a, b := combo[0], combo[1]

			// Check tier constraint
			aTier, aOk := tierMap[a]
			bTier, bOk := tierMap[b]
			resultTier := tierMap[elemToExpand]
			if !aOk || !bOk {
				continue
			}
			maxTier := aTier
			if bTier > maxTier {
				maxTier = bTier
			}
			if resultTier <= maxTier {
				continue
			}

			// Make a new step
			step := Step{Ingredients: [2]string{a, b}, Result: elemToExpand}
			stepKey := [3]string{a, b, elemToExpand}

			newSteps := make([]Step, len(curr.Steps))
			copy(newSteps, curr.Steps)
			newStepSet := copyStepSet(curr.StepSet)

			if !newStepSet[stepKey] {
				newSteps = append(newSteps, step)
				newStepSet[stepKey] = true
			}

			// New Remaining list: replace elemToExpand with [a, b], drop already basic ones
			newRemaining := []string{}
			for _, r := range curr.Remaining {
				if r == elemToExpand {
					if !basics[a] {
						newRemaining = append(newRemaining, a)
					}
					if !basics[b] {
						newRemaining = append(newRemaining, b)
					}
				} else if !basics[r] {
					newRemaining = append(newRemaining, r)
				}
			}

			queue = append(queue, BFSNode{
				Remaining: newRemaining,
				Steps:     newSteps,
				Visited:   newVisited,
				StepSet:   newStepSet,
			})
		}
	}

	return nil, time.Since(startTime), iterations
}

// Helper function to create a deep copy of the visited map
func copyVisitedMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}

// Helper function to create a deep copy of step set
func copyStepSet(original map[[3]string]bool) map[[3]string]bool {
	newMap := make(map[[3]string]bool)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}

func removeElement(slice []string, element string) []string {
	newSlice := []string{}
	for _, e := range slice {
		if e != element {
			newSlice = append(newSlice, e)
		}
	}
	return newSlice
}
