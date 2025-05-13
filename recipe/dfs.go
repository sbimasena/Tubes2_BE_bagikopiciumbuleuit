package recipe

import (
	"encoding/json"
	"os"
	"time"
)

type ElementRecipe struct {
	Element  string      `json:"element"`
	ImageURL string      `json:"image_url"`
	Recipes  [][2]string `json:"recipes"`
	Tier     int         `json:"tier"`
}

type Path struct {
	Steps     []Step `json:"steps"`
	FinalItem string `json:"final_item"`
}

/*
*

	Path{
	  FinalItem: "Barn",
	  Steps: []Step{
	    {Ingredients: [2]string{"Wall", "Wall"}, Result: "House"},
	    {Ingredients: [2]string{"Plow", "Earth"}, Result: "Field"},
	    {Ingredients: [2]string{"House", "Field"}, Result: "Barn"},
	  },
	}

*
*/
type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

func findPathDFS(recipes []ElementRecipe, startElements []string, target string) ([]Path, time.Duration, int) {
	startTime := time.Now()
	elementMap := make(map[string]ElementRecipe) // just like the json
	tierMap := make(map[string]int)              // element -> tier
	recipeMap := make(map[string][][2]string)    // element -> all recipe

	for _, recipe := range recipes {
		elementMap[recipe.Element] = recipe
		tierMap[recipe.Element] = recipe.Tier
		recipeMap[recipe.Element] = append(recipeMap[recipe.Element], recipe.Recipes...)
	}

	basics := make(map[string]bool)
	for _, e := range startElements {
		basics[e] = true
	}

	memo := make(map[string]bool)
	visitedCounter := make(map[string]bool)

	// Tambahkan counter untuk menghitung total node yang dieksplorasi
	nodesExplored := 0

	// Hitung target sebagai node yang dieksplorasi
	nodesExplored++

	// Hitung semua elemen dasar
	for range startElements {
		nodesExplored++
	}

	var dfs func(string) *Path
	dfs = func(current string) *Path {
		// Hitung current element sebagai node yang dieksplorasi
		// Kecuali jika sudah pernah dihitung sebelumnya
		if !visitedCounter[current] {
			nodesExplored++
		}

		// Tandai sudah mengunjungi current element
		visitedCounter[current] = true

		if basics[current] { // return if target = basic elements
			return &Path{Steps: []Step{}, FinalItem: current}
		}

		// check if element has been viisted but failed
		if success, ok := memo[current]; ok && !success {
			return nil
		}

		combos, ok := recipeMap[current] // look for the recipe in recipeMap
		if !ok {
			memo[current] = false
			return nil
		}

		// check the possible combo
		for _, combo := range combos {
			// Hitung setiap resep yang diperiksa
			nodesExplored++

			a, b := combo[0], combo[1]
			aTier, aOk := tierMap[a]
			bTier, bOk := tierMap[b]
			resultTier := tierMap[current]

			//continue to the next combo if not valid
			if !aOk || !bOk {
				continue
			}
			if resultTier <= max(aTier, bTier) {
				continue
			}

			//continue to the next combo if elements cant be crafted
			//dfs karena ngabisin path nya A dulu baru ke B
			pathA := dfs(a)
			if pathA == nil {
				continue
			}

			pathB := dfs(b)
			if pathB == nil {
				continue
			}

			// merge steps without duplicate (pathA and pathB)
			stepSet := make(map[[3]string]bool)
			var steps []Step
			for _, s := range pathA.Steps {
				k := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
				if !stepSet[k] {
					stepSet[k] = true
					steps = append(steps, s)
				}
			}
			for _, s := range pathB.Steps {
				k := [3]string{s.Ingredients[0], s.Ingredients[1], s.Result}
				if !stepSet[k] {
					stepSet[k] = true
					steps = append(steps, s)
				}
			}

			// add the last step to make the current element
			k := [3]string{a, b, current}
			if !stepSet[k] {
				steps = append(steps, Step{Ingredients: [2]string{a, b}, Result: current})
				// Hitung setiap step baru yang ditambahkan
				nodesExplored++
			}

			memo[current] = true
			return &Path{Steps: steps, FinalItem: current}
		}

		memo[current] = false
		return nil
	}

	path := dfs(target)
	duration := time.Since(startTime)

	if path != nil {
		return []Path{*path}, duration, nodesExplored
	}
	return nil, duration, nodesExplored
}

// LoadRecipes loads element recipes from a JSON file
func LoadRecipes(filename string) ([]ElementRecipe, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var recipes []ElementRecipe
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&recipes); err != nil {
		return nil, err
	}

	return recipes, nil
}
