package recipe

import (
	"container/list"
	"fmt"
	"time"
)

type BiState struct {
	Element        string
	Path           []string
	PathSteps      map[string][]string
	AvailableElems map[string]bool
}

func isValidElement(element string, elements map[string][][]string, tiers map[string]int) bool {
	// Harus punya info tier
	if _, hasTier := tiers[element]; !hasTier {
		return false
	}

	// Harus muncul sebagai hasil (elements map) atau bahan dalam recipe
	if _, isResult := elements[element]; isResult {
		return true
	}

	return false
}

func BiSearchBFS(target string, elements map[string][][]string, basicElements map[string]bool, tiers map[string]int) ([]string, map[string][]string, int, time.Duration) {
	startTime := time.Now()
	nodesExplored := 0

	// forward queue (dari basic element)
	forwardQueue := list.New()
	forwardParent := make(map[string]string)   // element -> its immediate parent
	forwardRecipe := make(map[string][]string) // element -> ingredients used to create it
	forwardVisited := make(map[string]bool)

	// forward queue (dari target element)
	backwardQueue := list.New()
	backwardParent := make(map[string]string)       // element -> one of its possible results
	backwardIngredient := make(map[string][]string) // element -> the pair of ingredients it's part of
	backwardVisited := make(map[string]bool)

	for elem := range basicElements {
		forwardQueue.PushBack(elem)
		forwardVisited[elem] = true
	}

	backwardQueue.PushBack(target)
	backwardVisited[target] = true

	meetingPoints := make(map[string]bool)

	// Alternating BFS
	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		levelSize := forwardQueue.Len()
		for i := 0; i < levelSize; i++ {
			current := forwardQueue.Remove(forwardQueue.Front()).(string)
			nodesExplored++

			// cek kalo udh ketemu
			if backwardVisited[current] {
				meetingPoints[current] = true
			}

			// coba craft dr basic element dan yg udh dipunya
			for result, recipes := range elements {
				if forwardVisited[result] {
					continue
				}

				resultTier, hasTier := tiers[result]
				if !hasTier {
					continue
				}

				// loop utk cari variasi resep dari current element
				for _, recipe := range recipes {
					if len(recipe) != 2 {
						continue
					}

					ing1, ing2 := recipe[0], recipe[1]
					ing1Tier, hasIng1Tier := tiers[ing1]
					ing2Tier, hasIng2Tier := tiers[ing2]

					if !hasIng1Tier || !hasIng2Tier || ing1Tier >= resultTier || ing2Tier >= resultTier {
						continue
					}

					// cek ingredients udah pernah divisit blm
					if forwardVisited[ing1] && forwardVisited[ing2] && !forwardVisited[result] {
						forwardVisited[result] = true
						forwardParent[result] = current
						forwardRecipe[result] = []string{ing1, ing2}
						forwardQueue.PushBack(result)

						if backwardVisited[result] {
							meetingPoints[result] = true
						}

						if result == target {
							meetingPoints[result] = true
							break
						}
					}

				}
			}
		}

		// kalau udh ada yg ketemu
		if len(meetingPoints) > 0 {

			allSteps := make(map[string][]string)

			for elem, recipe := range forwardRecipe {
				allSteps[elem] = recipe
			}

			for elem, ingredients := range backwardIngredient {
				parent := backwardParent[elem]
				if parent != "" && !basicElements[elem] {
					allSteps[parent] = ingredients
				}
			}

			// reconstruct a path
			for meetingPoint := range meetingPoints {
				// kalo meeting point nya dr bw
				if recipe, ok := backwardIngredient[meetingPoint]; ok && meetingPoint != target {
					allSteps[meetingPoint] = recipe
				}

				path := reconstructPath(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}
		}

		levelSize = backwardQueue.Len()
		for i := 0; i < levelSize; i++ {
			current := backwardQueue.Remove(backwardQueue.Front()).(string)
			nodesExplored++

			if forwardVisited[current] {
				meetingPoints[current] = true
			}

			if basicElements[current] {
				meetingPoints[current] = true
				continue
			}

			// find recipes that can create this element
			currentTier, hasTier := tiers[current]
			if !hasTier {
				continue
			}

			for _, recipe := range elements[current] {
				if len(recipe) != 2 {
					continue
				}

				ing1, ing2 := recipe[0], recipe[1]
				ing1Tier, hasIng1Tier := tiers[ing1]
				ing2Tier, hasIng2Tier := tiers[ing2]

				if !hasIng1Tier || !hasIng2Tier || ing1Tier >= currentTier || ing2Tier >= currentTier {
					continue
				}

				// process each ingredient
				for _, ing := range []string{ing1, ing2} {
					if !backwardVisited[ing] {
						backwardVisited[ing] = true
						backwardQueue.PushBack(ing)
						backwardParent[ing] = current
						backwardIngredient[ing] = []string{ing1, ing2} // simpan resep dari elemen saat ini

						if forwardVisited[ing] {
							meetingPoints[ing] = true
						}
					}
				}
			}
		}

		if len(meetingPoints) > 0 {
			allSteps := make(map[string][]string)

			// Add all forward recipes
			for elem, recipe := range forwardRecipe {
				allSteps[elem] = recipe
			}

			// Add all backward recipes
			for _ = range meetingPoints {
				path := reconstructPath(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}
		}
	}

	fmt.Println("No path found!")
	return nil, nil, nodesExplored, time.Since(startTime)
}

func BiSearchDFS(target string, elements map[string][][]string, basicElements map[string]bool, tiers map[string]int) ([]string, map[string][]string, int, time.Duration) {
	startTime := time.Now()

	type SearchState struct {
		Element   string
		Available map[string]bool
		Steps     map[string][]string
	}

	forwardStack := list.New()
	backwardStack := list.New()
	forwardVisited := make(map[string]map[string][]string)
	backwardVisited := make(map[string]map[string][]string)
	nodesExplored := 0

	for element := range basicElements {
		available := make(map[string]bool)
		for e := range basicElements {
			available[e] = true
		}

		state := SearchState{
			Element:   element,
			Available: available,
			Steps:     make(map[string][]string),
		}
		forwardStack.PushBack(state)
	}

	backwardState := SearchState{
		Element:   target,
		Available: map[string]bool{target: true},
		Steps:     make(map[string][]string),
	}

	if recipes, ok := elements[target]; ok {
		for _, recipe := range recipes {
			if len(recipe) == 2 {
				targetTier, hasTier := tiers[target]
				ing1Tier, hasIng1 := tiers[recipe[0]]
				ing2Tier, hasIng2 := tiers[recipe[1]]

				if hasTier && hasIng1 && hasIng2 &&
					ing1Tier < targetTier && ing2Tier < targetTier {
					backwardState.Steps[target] = recipe
					break
				}
			}
		}
	}

	backwardStack.PushBack(backwardState)

	// forward dfs
	for forwardStack.Len() > 0 {
		current := forwardStack.Remove(forwardStack.Back()).(SearchState)
		nodesExplored++

		if _, visited := forwardVisited[current.Element]; visited {
			continue
		}

		stepsCopy := make(map[string][]string)
		for k, v := range current.Steps {
			stepsCopy[k] = append([]string{}, v...)
		}
		forwardVisited[current.Element] = stepsCopy

		if backwardSteps, found := backwardVisited[current.Element]; found {
			allSteps := make(map[string][]string)
			for k, v := range stepsCopy {
				allSteps[k] = v
			}
			for k, v := range backwardSteps {
				allSteps[k] = v
			}
			if isStepsComplete(allSteps, basicElements) {
				path := reconstructPath(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}
		}

		if current.Element == target && isStepsComplete(current.Steps, basicElements) {
			path := reconstructPath(target, current.Steps, basicElements, elements, tiers)
			if path != nil {
				return path, current.Steps, nodesExplored, time.Since(startTime)
			}
		}

		newAvailable := make(map[string]bool)
		for k, v := range current.Available {
			newAvailable[k] = v
		}
		newAvailable[current.Element] = true

		for resultElem, recipes := range elements {
			if _, visited := forwardVisited[resultElem]; visited {
				continue
			}
			resultTier, ok := tiers[resultElem]
			if !ok {
				continue
			}

			for _, recipe := range recipes {
				if len(recipe) != 2 {
					continue
				}
				ing1, ing2 := recipe[0], recipe[1]
				t1, ok1 := tiers[ing1]
				t2, ok2 := tiers[ing2]
				if !ok1 || !ok2 || t1 >= resultTier || t2 >= resultTier {
					continue
				}

				ing1Avail := newAvailable[ing1] || current.Element == ing1
				ing2Avail := newAvailable[ing2] || current.Element == ing2

				if ing1Avail && ing2Avail {
					newSteps := make(map[string][]string)
					for k, v := range current.Steps {
						newSteps[k] = append([]string{}, v...)
					}
					newSteps[resultElem] = []string{ing1, ing2}

					state := SearchState{
						Element:   resultElem,
						Available: newAvailable,
						Steps:     newSteps,
					}
					forwardStack.PushBack(state)
				}
			}
		}
	}

	// backward dfs
	for backwardStack.Len() > 0 {
		current := backwardStack.Remove(backwardStack.Back()).(SearchState)
		nodesExplored++

		if _, visited := backwardVisited[current.Element]; visited {
			continue
		}

		currentTier, ok := tiers[current.Element]
		if !ok {
			continue
		}

		stepsCopy := make(map[string][]string)
		for k, v := range current.Steps {
			stepsCopy[k] = append([]string{}, v...)
		}
		backwardVisited[current.Element] = stepsCopy

		if forwardSteps, found := forwardVisited[current.Element]; found {
			allSteps := make(map[string][]string)
			for k, v := range forwardSteps {
				allSteps[k] = v
			}
			for k, v := range stepsCopy {
				allSteps[k] = v
			}
			if isStepsComplete(allSteps, basicElements) {
				path := reconstructPath(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}
		}

		recipes, hasRecipes := elements[current.Element]
		if !hasRecipes {
			continue
		}

		for _, recipe := range recipes {
			if len(recipe) != 2 {
				continue
			}

			ing1, ing2 := recipe[0], recipe[1]
			t1, ok1 := tiers[ing1]
			t2, ok2 := tiers[ing2]
			if !ok1 || !ok2 || t1 >= currentTier || t2 >= currentTier {
				continue
			}

			newSteps := make(map[string][]string)
			for k, v := range current.Steps {
				newSteps[k] = append([]string{}, v...)
			}
			newSteps[current.Element] = recipe

			// push ing1
			if isValidElement(ing1, elements, tiers) {
				if _, visited := backwardVisited[ing1]; !visited {
					state := SearchState{
						Element:   ing1,
						Available: cloneMap(current.Available),
						Steps:     newSteps,
					}
					state.Available[ing1] = true
					backwardStack.PushBack(state)
				}
			}
			// push ing2
			if isValidElement(ing2, elements, tiers) {
				if _, visited := backwardVisited[ing2]; !visited {
					state := SearchState{
						Element:   ing2,
						Available: cloneMap(current.Available),
						Steps:     newSteps,
					}
					state.Available[ing2] = true
					backwardStack.PushBack(state)
				}
			}
		}
	}

	fmt.Println("No path found after exploring", nodesExplored, "nodes")
	return nil, nil, nodesExplored, time.Since(startTime)
}
func isStepsComplete(steps map[string][]string, basicElements map[string]bool) bool {
	for _, ingredients := range steps {
		for _, ing := range ingredients {
			if !basicElements[ing] && steps[ing] == nil {
				return false
			}
		}
	}
	return true
}

func cloneMap(src map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range src {
		copy[k] = v
	}
	return copy
}

func reconstructPath(target string, steps map[string][]string, basicElements map[string]bool,
	_ map[string][][]string, tiers map[string]int) []string {

	dependsOn := make(map[string][]string) // element -> ingredients1, ingredients2
	inDegree := make(map[string]int)       // berapa banyak ingredient yang dibutuhkan untuk membuat element

	// untuk setiap elemen di steps, simpan dua bahan pembuatnya
	for elem, ingredients := range steps {
		if len(ingredients) == 2 {
			ing1, ing2 := ingredients[0], ingredients[1]

			// Validasi tier
			elemTier, hasTier := tiers[elem]
			ing1Tier, hasIng1Tier := tiers[ing1]
			ing2Tier, hasIng2Tier := tiers[ing2]

			if !hasTier || !hasIng1Tier || !hasIng2Tier {
				continue // skip jika tidak ada info tier
			}

			if ing1Tier >= elemTier || ing2Tier >= elemTier {
				continue // skip jika melanggar constraint tier
			}

			dependsOn[elem] = []string{ing1, ing2}
			inDegree[elem] = 2
		}
	}

	var queue []string

	// Tambahkan elemen dasar ke queue
	for elem := range basicElements {
		queue = append(queue, elem)
	}

	// Elemen yang sudah tersedia
	available := make(map[string]bool)
	for elem := range basicElements {
		available[elem] = true
	}

	var buildPath []string

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// skip elemen dasar dalam path final
		if !basicElements[current] {
			buildPath = append(buildPath, current)
		}

		available[current] = true

		if current == target {
			break
		}

		for elem, ingredients := range dependsOn {
			if available[elem] {
				continue
			}

			if ingredients[0] == current || ingredients[1] == current {
				inDegree[elem]--
			}

			// kalau bahan sudah tersedia, masukkin ke jalur
			if inDegree[elem] == 0 {
				queue = append(queue, elem)
			}
		}
	}

	// cek apakah target ditemukan
	if !available[target] {
		return nil
	}

	return buildPath
}
