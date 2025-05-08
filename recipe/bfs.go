package recipe

type Node struct {
	Element string   // current element
	Path    []string // path from target down to this
}

// ReverseBFSBranch explores from target and returns the first fully basic-resolved branch
type bfsState struct {
	Remaining    []string            // Elements still to resolve
	Path         []string            // Flattened recipe path so far
	Used         map[string]bool     // For deduplication in path
	Combinations map[string][]string // To fill the combinations map
}

func bfsWithFlattenedRecipe(target string, elements map[string][][]string, basicElements map[string]bool) ([]string, map[string][]string) {
	// Initial state: try all recipes of target
	queue := []bfsState{}
	for _, recipe := range elements[target] {
		if len(recipe) != 2 {
			continue
		}
		initialUsed := make(map[string]bool)
		initialPath := []string{}
		for _, ing := range recipe {
			if !initialUsed[ing] {
				initialPath = append(initialPath, ing)
				initialUsed[ing] = true
			}
		}
		queue = append(queue, bfsState{
			Remaining: recipe,
			Path:      initialPath,
			Used:      initialUsed,
			Combinations: map[string][]string{
				target: recipe,
			},
		})
	}

	// Store the shortest recipe found
	var bestPath []string
	var bestCombinations map[string][]string

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		fullyResolved := true

		for _, ing := range curr.Remaining {
			if basicElements[ing] {
				continue
			}
			fullyResolved = false

			// Expand this non-basic element using its recipes
			for _, recipe := range elements[ing] {
				if len(recipe) != 2 {
					continue
				}

				// Clone current state
				newUsed := copySet(curr.Used)
				newPath := append([]string{}, curr.Path...)
				for _, r := range recipe {
					if !newUsed[r] {
						newPath = append(newPath, r)
						newUsed[r] = true
					}
				}

				newCombos := copyCombos(curr.Combinations)
				newCombos[ing] = recipe

				// Replace current ingredient with the ingredients of the recipe
				nextRemaining := []string{}
				for _, e := range curr.Remaining {
					if e == ing {
						nextRemaining = append(nextRemaining, recipe...)
					} else {
						nextRemaining = append(nextRemaining, e)
					}
				}

				queue = append(queue, bfsState{
					Remaining:    nextRemaining,
					Path:         newPath,
					Used:         newUsed,
					Combinations: newCombos,
				})
			}
			break // only expand one unresolved element per iteration
		}

		if fullyResolved {
			finalPath := append([]string{}, curr.Path...)
			if !curr.Used[target] {
				finalPath = append(finalPath, target)
			}
			return finalPath, curr.Combinations // exit immediately
		}

	}

	return bestPath, bestCombinations
}

func copySet(original map[string]bool) map[string]bool {
	newSet := make(map[string]bool)
	for k, v := range original {
		newSet[k] = v
	}
	return newSet
}

func copyCombos(original map[string][]string) map[string][]string {
	newMap := make(map[string][]string)
	for k, v := range original {
		cp := make([]string, len(v))
		copy(cp, v)
		newMap[k] = cp
	}
	return newMap
}
