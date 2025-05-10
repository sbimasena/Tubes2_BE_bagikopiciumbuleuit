package main

import (
	"container/list"
	"fmt"
	"maps"
	"time"
)

type BiState struct {
	Element        string
	Path           []string
	PathSteps      map[string][]string
	AvailableElems map[string]bool
}

// Helper function to check if an element is valid
func isValidElement(element string, elements map[string][][]string, tiers map[string]int) bool {
	// 1. Harus punya info tier
	if _, hasTier := tiers[element]; !hasTier {
		return false
	}

	// 2. Harus muncul sebagai hasil (elements map) atau bahan dalam recipe
	if _, isResult := elements[element]; isResult {
		return true
	}

	return false
}

// Implementasi BFS Bidirectional dengan constraint tier dan return time duration
func BiSearchBFS(target string, elements map[string][][]string, basicElements map[string]bool, tiers map[string]int) ([]string, map[string][]string, int, time.Duration) {
	// Catat waktu mulai
	startTime := time.Now()

	// Struktur data untuk BFS
	forwardQueue := list.New()
	backwardQueue := list.New()
	forwardVisited := make(map[string]BiState)
	backwardVisited := make(map[string]BiState)
	nodesExplored := 0

	// Inisialisasi pencarian maju
	for element := range basicElements {
		availableElems := make(map[string]bool)
		for e := range basicElements {
			availableElems[e] = true
		}

		state := BiState{
			Element:        element,
			Path:           []string{element},
			PathSteps:      make(map[string][]string),
			AvailableElems: availableElems,
		}

		forwardQueue.PushBack(state)
		forwardVisited[element] = state
	}

	// Inisialisasi pencarian mundur
	backwardState := BiState{
		Element:        target,
		Path:           []string{target},
		PathSteps:      make(map[string][]string),
		AvailableElems: map[string]bool{target: true},
	}
	backwardQueue.PushBack(backwardState)
	backwardVisited[target] = backwardState

	// Loop utama BFS
	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		// Forward BFS
		size := forwardQueue.Len()
		for i := 0; i < size; i++ {
			current := forwardQueue.Remove(forwardQueue.Front()).(BiState)
			nodesExplored++

			// Cek pertemuan
			if bwState, ok := backwardVisited[current.Element]; ok {

				// Gabungkan steps
				allSteps := make(map[string][]string)
				for k, v := range current.PathSteps {
					allSteps[k] = v
				}
				for k, v := range bwState.PathSteps {
					allSteps[k] = v
				}

				// Rekonstruksi jalur secara iteratif
				path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}

			// Coba buat elemen baru
			for resultElem, recipes := range elements {
				// Skip jika sudah dikunjungi
				if !isValidElement(resultElem, elements, tiers) {
					continue
				}

				// Dapatkan tier dari elemen hasil
				resultTier, hasTier := tiers[resultElem]
				if !hasTier {
					// Skip jika tidak punya tier info
					continue
				}

				// Coba setiap resep
				for _, recipe := range recipes {
					if len(recipe) != 2 {
						continue
					}

					ing1 := recipe[0]
					ing2 := recipe[1]
					if !isValidElement(ing1, elements, tiers) || !isValidElement(ing2, elements, tiers) {
						continue
					}

					// Dapatkan tier dari bahan-bahan
					ing1Tier, hasIng1Tier := tiers[ing1]
					ing2Tier, hasIng2Tier := tiers[ing2]

					// Skip jika tidak punya tier info
					if !hasIng1Tier || !hasIng2Tier {
						continue
					}

					// Constraint tier: bahan harus tier lebih rendah dari hasil
					if ing1Tier >= resultTier || ing2Tier >= resultTier {
						continue
					}

					// Cek apakah kedua bahan tersedia
					ing1Available := current.AvailableElems[ing1] || current.Element == ing1
					ing2Available := current.AvailableElems[ing2] || current.Element == ing2

					if ing1Available && ing2Available {
						// Buat state baru
						newAvailable := make(map[string]bool)
						for k, v := range current.AvailableElems {
							newAvailable[k] = v
						}
						newAvailable[resultElem] = true

						newPath := make([]string, len(current.Path))
						copy(newPath, current.Path)
						newPath = append(newPath, resultElem)

						newSteps := make(map[string][]string)
						for k, v := range current.PathSteps {
							newSteps[k] = make([]string, len(v))
							copy(newSteps[k], v)
						}
						newSteps[resultElem] = []string{ing1, ing2}

						// Buat state baru
						newState := BiState{
							Element:        resultElem,
							Path:           newPath,
							PathSteps:      newSteps,
							AvailableElems: newAvailable,
						}

						// Tambahkan ke queue dan visited
						forwardQueue.PushBack(newState)
						forwardVisited[resultElem] = newState

						// Cek apakah ini target
						if resultElem == target {

							// Buat jalur final
							finalPath := filterBasicElements(newPath, basicElements)
							return finalPath, newSteps, nodesExplored, time.Since(startTime)
						}

						// Cek pertemuan
						if _, ok := backwardVisited[resultElem]; ok {

							// Gabungkan steps
							allSteps := make(map[string][]string)
							for k, v := range newSteps {
								allSteps[k] = v
							}
							for k, v := range backwardVisited[resultElem].PathSteps {
								allSteps[k] = v
							}

							// Rekonstruksi jalur
							path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
							if path != nil {
								return path, allSteps, nodesExplored, time.Since(startTime)
							}
						}
					}
				}
			}
		}

		// Backward BFS
		size = backwardQueue.Len()
		for i := 0; i < size; i++ {
			current := backwardQueue.Remove(backwardQueue.Front()).(BiState)
			nodesExplored++

			// Dapatkan tier dari elemen saat ini
			currentTier, hasTier := tiers[current.Element]
			if !hasTier {
				// Skip jika tidak punya tier info
				continue
			}

			// Cek pertemuan
			if fwState, ok := forwardVisited[current.Element]; ok {

				// Gabungkan steps
				allSteps := make(map[string][]string)
				for k, v := range fwState.PathSteps {
					allSteps[k] = v
				}
				for k, v := range current.PathSteps {
					allSteps[k] = v
				}

				// Rekonstruksi jalur
				path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}

			// Cari resep yang menghasilkan elemen ini
			recipes, hasRecipes := elements[current.Element]
			if !hasRecipes {
				continue
			}

			// Coba setiap resep
			for _, recipe := range recipes {
				if len(recipe) != 2 {
					continue
				}

				// Coba setiap bahan
				for _, ingredient := range recipe {
					if !isValidElement(ingredient, elements, tiers) {
						continue
					}
					// Skip jika sudah dikunjungi
					if _, visited := backwardVisited[ingredient]; visited {
						continue
					}

					// Dapatkan tier dari bahan
					ingTier, hasIngTier := tiers[ingredient]
					if !hasIngTier {
						// Skip jika tidak punya tier info
						continue
					}

					// Constraint tier: bahan harus tier lebih rendah dari hasil
					if ingTier >= currentTier {
						continue
					}

					// Buat state baru
					newAvailable := make(map[string]bool)
					for k, v := range current.AvailableElems {
						newAvailable[k] = v
					}
					newAvailable[ingredient] = true

					newPath := make([]string, len(current.Path))
					copy(newPath, current.Path)
					newPath = append(newPath, ingredient)

					newSteps := make(map[string][]string)
					for k, v := range current.PathSteps {
						newSteps[k] = make([]string, len(v))
						copy(newSteps[k], v)
					}
					newSteps[current.Element] = recipe

					// Buat state baru
					newState := BiState{
						Element:        ingredient,
						Path:           newPath,
						PathSteps:      newSteps,
						AvailableElems: newAvailable,
					}

					// Tambahkan ke queue dan visited
					backwardQueue.PushBack(newState)
					backwardVisited[ingredient] = newState

					// Cek apakah ini elemen dasar
					if basicElements[ingredient] {

						// Gabungkan steps
						allSteps := make(map[string][]string)
						for k, v := range forwardVisited[ingredient].PathSteps {
							allSteps[k] = v
						}
						for k, v := range newSteps {
							allSteps[k] = v
						}

						// Rekonstruksi jalur
						path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
						if path != nil {
							return path, allSteps, nodesExplored, time.Since(startTime)
						}
					}

					// Cek pertemuan
					if _, ok := forwardVisited[ingredient]; ok {

						// Gabungkan steps
						allSteps := make(map[string][]string)
						for k, v := range forwardVisited[ingredient].PathSteps {
							allSteps[k] = v
						}
						for k, v := range newSteps {
							allSteps[k] = v
						}

						// Rekonstruksi jalur
						path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
						if path != nil {
							return path, allSteps, nodesExplored, time.Since(startTime)
						}
					}
				}
			}
		}

	}

	fmt.Println("No path found!")
	return nil, nil, nodesExplored, time.Since(startTime)
}

// Implementasi DFS Bidirectional dengan constraint tier dan return time duration
func BiSearchDFS(target string, elements map[string][][]string, basicElements map[string]bool, tiers map[string]int) ([]string, map[string][]string, int, time.Duration) {
	// Catat waktu mulai
	startTime := time.Now()

	// Struktur state untuk stack
	type SearchState struct {
		Element   string
		Available map[string]bool
		Steps     map[string][]string
	}

	// Struktur data untuk DFS
	forwardStack := list.New()
	backwardStack := list.New()
	forwardVisited := make(map[string]map[string][]string)
	backwardVisited := make(map[string]map[string][]string)
	nodesExplored := 0

	// Inisialisasi forward stack dengan elemen dasar
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

	// Inisialisasi backward stack dengan target
	backwardState := SearchState{
		Element:   target,
		Available: map[string]bool{target: true},
		Steps:     make(map[string][]string),
	}
	backwardStack.PushBack(backwardState)

	// Forward DFS loop
	for forwardStack.Len() > 0 {
		// Pop dari stack
		current := forwardStack.Remove(forwardStack.Back()).(SearchState)
		nodesExplored++

		// Skip jika sudah dikunjungi
		if _, visited := forwardVisited[current.Element]; visited {
			continue
		}

		// Tandai sebagai dikunjungi
		stepsCopy := make(map[string][]string)
		for k, v := range current.Steps {
			stepsCopy[k] = make([]string, len(v))
			copy(stepsCopy[k], v)
		}
		forwardVisited[current.Element] = stepsCopy

		// Cek pertemuan dengan backward search
		if backwardSteps, found := backwardVisited[current.Element]; found {
			// Gabungkan steps
			allSteps := make(map[string][]string)
			for k, v := range stepsCopy {
				allSteps[k] = v
			}
			for k, v := range backwardSteps {
				allSteps[k] = v
			}

			// Rekonstruksi jalur
			path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
			if path != nil {
				return path, allSteps, nodesExplored, time.Since(startTime)
			}
		}

		// Cek apakah ini target
		if current.Element == target {
			path := reconstructPathIterative(target, current.Steps, basicElements, elements, tiers)
			if path != nil {
				return path, current.Steps, nodesExplored, time.Since(startTime)
			}
		}

		// Update available
		newAvailable := make(map[string]bool)
		for k, v := range current.Available {
			newAvailable[k] = v
		}
		newAvailable[current.Element] = true

		// Coba buat elemen baru
		for resultElem, recipes := range elements {
			// Skip jika sudah dikunjungi
			if _, visited := forwardVisited[resultElem]; visited {
				continue
			}

			// Dapatkan tier dari elemen hasil
			resultTier, hasTier := tiers[resultElem]
			if !hasTier {
				// Skip jika tidak punya tier info
				continue
			}

			// Coba setiap resep
			for _, recipe := range recipes {
				if len(recipe) != 2 {
					continue
				}

				ing1 := recipe[0]
				ing2 := recipe[1]

				// Dapatkan tier dari bahan-bahan
				ing1Tier, hasIng1Tier := tiers[ing1]
				ing2Tier, hasIng2Tier := tiers[ing2]

				// Skip jika tidak punya tier info
				if !hasIng1Tier || !hasIng2Tier {
					continue
				}

				// Constraint tier: bahan harus tier lebih rendah dari hasil
				if ing1Tier >= resultTier || ing2Tier >= resultTier {
					continue
				}

				// Cek apakah kedua bahan tersedia
				ing1Available := newAvailable[ing1] || current.Element == ing1
				ing2Available := newAvailable[ing2] || current.Element == ing2

				if ing1Available && ing2Available {
					// Buat state baru
					newSteps := make(map[string][]string)
					for k, v := range current.Steps {
						newSteps[k] = make([]string, len(v))
						copy(newSteps[k], v)
					}
					newSteps[resultElem] = []string{ing1, ing2}

					// Push ke stack
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

	// Backward DFS loop
	for backwardStack.Len() > 0 {
		// Pop dari stack
		current := backwardStack.Remove(backwardStack.Back()).(SearchState)
		nodesExplored++

		// Skip jika sudah dikunjungi
		if _, visited := backwardVisited[current.Element]; visited {
			continue
		}

		// Dapatkan tier dari elemen saat ini
		currentTier, hasTier := tiers[current.Element]
		if !hasTier {
			// Skip jika tidak punya tier info
			continue
		}

		// Tandai sebagai dikunjungi
		stepsCopy := make(map[string][]string)
		for k, v := range current.Steps {
			stepsCopy[k] = make([]string, len(v))
			copy(stepsCopy[k], v)
		}
		backwardVisited[current.Element] = stepsCopy

		// Cek pertemuan dengan forward search
		if forwardSteps, found := forwardVisited[current.Element]; found {
			// Gabungkan steps
			allSteps := make(map[string][]string)
			for k, v := range forwardSteps {
				allSteps[k] = v
			}
			for k, v := range stepsCopy {
				allSteps[k] = v
			}

			// Rekonstruksi jalur
			path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
			if path != nil {
				return path, allSteps, nodesExplored, time.Since(startTime)
			}
		}

		// Cek apakah ini elemen dasar
		if basicElements[current.Element] {
			if forwardSteps, found := forwardVisited[current.Element]; found {
				// Gabungkan steps
				allSteps := make(map[string][]string)
				maps.Copy(allSteps, forwardSteps)
				for k, v := range stepsCopy {
					allSteps[k] = v
				}

				// Rekonstruksi jalur
				path := reconstructPathIterative(target, allSteps, basicElements, elements, tiers)
				if path != nil {
					return path, allSteps, nodesExplored, time.Since(startTime)
				}
			}
		}

		// Cari resep yang menghasilkan elemen ini
		recipes, hasRecipes := elements[current.Element]
		if !hasRecipes {
			continue
		}

		// Coba setiap resep
		for _, recipe := range recipes {
			if len(recipe) != 2 {
				continue
			}

			// Simpan resep
			newSteps := make(map[string][]string)
			for k, v := range current.Steps {
				newSteps[k] = make([]string, len(v))
				copy(newSteps[k], v)
			}
			newSteps[current.Element] = recipe

			// Coba setiap bahan
			for _, ingredient := range recipe {
				// Skip jika sudah dikunjungi
				if !isValidElement(ingredient, elements, tiers) {
					continue
				}
				if _, visited := backwardVisited[ingredient]; visited {
					continue
				}

				// Dapatkan tier dari bahan
				ingTier, hasIngTier := tiers[ingredient]
				if !hasIngTier {
					// Skip jika tidak punya tier info
					continue
				}

				// Constraint tier: bahan harus tier lebih rendah dari hasil
				if ingTier >= currentTier {
					continue
				}

				// Buat state baru
				state := SearchState{
					Element:   ingredient,
					Available: make(map[string]bool),
					Steps:     newSteps,
				}

				// Copy available dan tambahkan ingredient
				for k, v := range current.Available {
					state.Available[k] = v
				}
				state.Available[ingredient] = true

				// Push ke stack
				backwardStack.PushBack(state)
			}
		}
	}

	fmt.Println("No path found after exploring", nodesExplored, "nodes")
	return nil, nil, nodesExplored, time.Since(startTime)
}

// Rekonstruksi jalur secara iteratif (tanpa rekursi)
func reconstructPathIterative(target string, steps map[string][]string, basicElements map[string]bool,
	elements map[string][][]string, tiers map[string]int) []string {
	if !isValidElement(target, elements, tiers) {
		fmt.Printf("Target is not a valid element: %s\n", target)
		return nil
	}

	// Cek apakah target punya resep
	if _, hasTargetRecipe := steps[target]; !hasTargetRecipe {
		recipes, ok := elements[target]
		if !ok || len(recipes) == 0 {
			fmt.Printf("No recipe found for target: %s\n", target)
			return nil
		}

		_, hasTier := tiers[target]
		if !hasTier {
			fmt.Printf("No tier info for target: %s\n", target)
			return nil
		}
		validFound := false
		for _, recipe := range recipes {
			if len(recipe) != 2 {
				continue
			}
			ing1, ing2 := recipe[0], recipe[1]
			t1, ok1 := tiers[ing1]
			t2, ok2 := tiers[ing2]
			targetTier, okTarget := tiers[target]
			if !isValidElement(ing1, elements, tiers) || !isValidElement(ing2, elements, tiers) {
				continue
			}
			if !ok1 || !ok2 || !okTarget {
				continue
			}
			if t1 >= targetTier || t2 >= targetTier {
				continue
			}

			steps[target] = recipe
			validFound = true
			break
		}

		if !validFound {
			return nil
		}

	}

	type StackItem struct {
		Element string
		Visited bool
	}

	stack := list.New()
	stack.PushBack(StackItem{Element: target, Visited: false})

	available := make(map[string]bool)
	for e := range basicElements {
		available[e] = true
	}
	visited := make(map[string]bool)
	var finalPath []string

	for stack.Len() > 0 {
		item := stack.Back().Value.(StackItem)
		stack.Remove(stack.Back())

		if available[item.Element] {
			continue
		}
		if !isValidElement(item.Element, elements, tiers) {
			continue
		}

		if !item.Visited {
			stack.PushBack(StackItem{Element: item.Element, Visited: true})

			recipe, hasRecipe := steps[item.Element]
			if !hasRecipe {
				recipes, ok := elements[item.Element]
				if !ok || len(recipes) == 0 {
					fmt.Printf("No recipe found for: %s\n", item.Element)
					continue
				}
				elemTier, hasTier := tiers[item.Element]
				if !hasTier {
					fmt.Printf("No tier info for: %s\n", item.Element)
					continue
				}
				for _, r := range recipes {
					if len(r) != 2 {
						continue
					}
					ing1, ing2 := r[0], r[1]
					t1, ok1 := tiers[ing1]
					t2, ok2 := tiers[ing2]
					if !ok1 || !ok2 {
						continue
					}
					if t1 >= elemTier || t2 >= elemTier {
						continue
					}
					recipe = r
					steps[item.Element] = r
					break
				}
			}

			for _, ingredient := range recipe {
				if !available[ingredient] && !visited[ingredient] {
					stack.PushBack(StackItem{Element: ingredient, Visited: false})
				}
				if !isValidElement(ingredient, elements, tiers) {
					return nil
				}
			}
		} else {
			allAvailable := true
			for _, ing := range steps[item.Element] {
				if !available[ing] {
					allAvailable = false
					break
				}
			}
			if allAvailable {
				finalPath = append(finalPath, item.Element)
				available[item.Element] = true
				visited[item.Element] = true
			} else {
				fmt.Printf("Not all ingredients available for: %s\n", item.Element)
				return nil
			}
		}
	}

	return filterBasicElements(finalPath, basicElements)
}

// Filter elemen dasar dari jalur
func filterBasicElements(path []string, basicElements map[string]bool) []string {
	var filtered []string
	basicAdded := make(map[string]bool)

	for _, elem := range path {
		if basicElements[elem] {
			// Hanya tambahkan elemen dasar sekali
			if !basicAdded[elem] {
				filtered = append(filtered, elem)
				basicAdded[elem] = true
			}
		} else {
			filtered = append(filtered, elem)
		}
	}

	return filtered
}
