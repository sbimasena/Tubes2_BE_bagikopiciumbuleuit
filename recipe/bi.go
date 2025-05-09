package main

import (
	"container/list"
	"sync"
)

type BiState struct {
	Element   string
	Path      []string
	PathSteps map[string][]string
	Parent    string
}

func BiSearch(target string, elements map[string][][]string, basicElements map[string]bool) ([]string, map[string][]string, int) {
	visitedForward := make(map[string]BiState)
	visitedBackward := make(map[string]BiState)
	forwardQueue := list.New()
	backwardQueue := list.New()

	for element := range basicElements {
		state := BiState{Element: element, Path: []string{element}, PathSteps: map[string][]string{}}
		forwardQueue.PushBack(state)
		visitedForward[element] = state
	}

	targetState := BiState{Element: target, Path: []string{target}, PathSteps: map[string][]string{}}
	backwardQueue.PushBack(targetState)
	visitedBackward[target] = targetState

	nodesExplored := 0

	checkIntersection := func() (string, bool) {
		for elem := range visitedForward {
			if _, ok := visitedBackward[elem]; ok {
				return elem, true
			}
		}
		return "", false
	}

	reconstructPath := func(mid string) ([]string, map[string][]string) {
		forward := visitedForward[mid]
		backward := visitedBackward[mid]

		steps := copyPathSteps(forward.PathSteps)
		for k, v := range backward.PathSteps {
			steps[k] = v
		}

		path := []string{}
		for _, e := range forward.Path {
			path = append(path, e)
		}
		for i := len(backward.Path) - 1; i >= 0; i-- {
			if backward.Path[i] != mid {
				path = append(path, backward.Path[i])
			}
		}
		return path, steps
	}

	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		if mid, found := checkIntersection(); found {
			p, s := reconstructPath(mid)
			return p, s, nodesExplored
		}

		for i := 0; i < forwardQueue.Len(); i++ {
			elem := forwardQueue.Remove(forwardQueue.Front()).(BiState)
			nodesExplored++

			for other := range visitedForward {
				for result, recipes := range elements {
					for _, recipe := range recipes {
						if len(recipe) != 2 {
							continue
						}
						if (recipe[0] == elem.Element && recipe[1] == other) || (recipe[1] == elem.Element && recipe[0] == other) {
							newPath := append([]string{}, elem.Path...)
							if !contains(newPath, other) {
								newPath = append(newPath, other)
							}
							if !contains(newPath, result) {
								newPath = append(newPath, result)
							}

							steps := copyPathSteps(elem.PathSteps)
							steps[result] = []string{elem.Element, other}

							if _, ok := visitedForward[result]; !ok {
								state := BiState{Element: result, Path: newPath, PathSteps: steps, Parent: elem.Element}
								forwardQueue.PushBack(state)
								visitedForward[result] = state
							}
						}
					}
				}
			}
		}

		for i := 0; i < backwardQueue.Len(); i++ {
			elem := backwardQueue.Remove(backwardQueue.Front()).(BiState)
			nodesExplored++

			for _, recipe := range elements[elem.Element] {
				if len(recipe) != 2 {
					continue
				}
				for _, ing := range recipe {
					if _, ok := visitedBackward[ing]; ok {
						continue
					}
					newPath := append([]string{}, elem.Path...)
					newPath = append(newPath, ing)

					steps := copyPathSteps(elem.PathSteps)
					steps[elem.Element] = recipe

					state := BiState{Element: ing, Path: newPath, PathSteps: steps, Parent: elem.Element}
					backwardQueue.PushBack(state)
					visitedBackward[ing] = state

					if basicElements[ing] {
						if _, ok := visitedForward[ing]; ok {
							p, s := reconstructPath(ing)
							return p, s, nodesExplored
						}
					}
				}
			}
		}
	}

	return nil, nil, nodesExplored
}

func BiSearchMultiple(target string, elements map[string][][]string, basicElements map[string]bool, maxPaths int) ([][]string, []map[string][]string, int) {
	pathsChan := make(chan []string, maxPaths)
	stepsChan := make(chan map[string][]string, maxPaths)
	nodesChan := make(chan int, maxPaths)

	var wg sync.WaitGroup
	wg.Add(maxPaths)

	mutex := &sync.Mutex{}
	elementsCopy := copyElements(elements)

	for i := 0; i < maxPaths; i++ {
		go func() {
			defer wg.Done()
			localElements := copyElements(elementsCopy)
			p, s, n := BiSearch(target, localElements, basicElements)
			if p != nil {
				pathsChan <- p
				stepsChan <- s
				nodesChan <- n
				mutex.Lock()
				for k, v := range s {
					removeRecipe(elementsCopy, k, v)
					break
				}
				mutex.Unlock()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(pathsChan)
		close(stepsChan)
		close(nodesChan)
	}()

	var paths [][]string
	var steps []map[string][]string
	total := 0

	for p := range pathsChan {
		paths = append(paths, p)
	}
	for s := range stepsChan {
		steps = append(steps, s)
	}
	for n := range nodesChan {
		total += n
	}

	return paths, steps, total
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func copyPathSteps(original map[string][]string) map[string][]string {
	newMap := make(map[string][]string)
	for k, v := range original {
		cp := make([]string, len(v))
		copy(cp, v)
		newMap[k] = cp
	}
	return newMap
}

func copyElements(original map[string][][]string) map[string][][]string {
	newMap := make(map[string][][]string)
	for k, recipes := range original {
		var newRecipes [][]string
		for _, recipe := range recipes {
			cp := make([]string, len(recipe))
			copy(cp, recipe)
			newRecipes = append(newRecipes, cp)
		}
		newMap[k] = newRecipes
	}
	return newMap
}

func removeRecipe(elements map[string][][]string, elem string, recipe []string) {
	var filtered [][]string
	for _, r := range elements[elem] {
		if len(r) == len(recipe) && r[0] == recipe[0] && r[1] == recipe[1] {
			continue
		}
		filtered = append(filtered, r)
	}
	elements[elem] = filtered
}
