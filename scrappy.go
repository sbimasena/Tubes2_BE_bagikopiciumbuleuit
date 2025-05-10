package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ElementRecipe struct {
	Element  string      `json:"element"`
	ImageURL string      `json:"image_url"`
	Recipes  [][2]string `json:"recipes"`
	Tier     int         `json:"tier"` // Added tier information
}

const url = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

func normalizeName(s string) string {
	// Remove anything after a newline
	parts := strings.Split(s, "\n")
	s = parts[0]

	// Remove special characters and lowercase
	s = strings.TrimSpace(strings.ToLower(s))

	return s
}

func getDoc() (*goquery.Document, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	return goquery.NewDocumentFromReader(res.Body)
}

func main() {
	start := time.Now()
	doc, err := getDoc()
	if err != nil {
		log.Fatal(err)
	}

	// STEP 2: Complete reset of approach - use raw DOM inspection and build tier map methodically
	elementTiers := make(map[string]int)

	// Print the structure of the document for debugging
	fmt.Println("Starting DOM inspection...")

	// Find all spans with mw-headline class that match Tier pattern
	var tierHeadings []struct {
		Tier int
		Elem *goquery.Selection
	}

	// First collect all tier headings with their numbers
	doc.Find("span.mw-headline").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			return
		}

		// Check if this is a tier heading
		tierMatch := regexp.MustCompile(`Tier_(\d+)_elements`).FindStringSubmatch(id)
		if len(tierMatch) < 2 {
			return
		}

		tierNum, _ := strconv.Atoi(tierMatch[1])
		tierHeadings = append(tierHeadings, struct {
			Tier int
			Elem *goquery.Selection
		}{
			Tier: tierNum,
			Elem: s.Parent(), // Get the h2 that contains this span
		})

		fmt.Printf("Found tier heading %d: %s (id=%s)\n", tierNum, s.Text(), id)
	})

	// Sort tier headings by their position in the document (not by tier number)
	// This will help us determine section boundaries

	// Now iterate through all elements after each tier heading until the next tier heading
	for i, th := range tierHeadings {
		currentTier := th.Tier
		currentHeading := th.Elem

		fmt.Printf("\n==== PROCESSING TIER %d ====\n", currentTier)

		// Find all element names until the next tier heading or end of document
		var nextHeading *goquery.Selection
		if i < len(tierHeadings)-1 {
			nextHeading = tierHeadings[i+1].Elem
		}

		// Get all h3s (element names) between current heading and next heading
		var elementsInThisTier []string

		// Process this tier's section
		currentNode := currentHeading.Next()
		for currentNode.Length() > 0 {
			// Stop if we've reached the next tier heading
			if nextHeading != nil && currentNode.Is(nextHeading.Nodes[0].Data) && currentNode.HasClass(nextHeading.AttrOr("class", "")) {
				break
			}

			// Check if this node is an h3 (element name)
			if goquery.NodeName(currentNode) == "h3" {
				elementName := strings.TrimSpace(currentNode.Text())
				if elementName != "" {
					elementsInThisTier = append(elementsInThisTier, elementName)
					normalizedName := normalizeName(elementName)
					elementTiers[normalizedName] = currentTier
					fmt.Printf("  📦 [h3] %-30s → Tier %d\n", elementName, currentTier)
				}
			}

			// Also look for tables with elements
			if goquery.NodeName(currentNode) == "table" || currentNode.Find("table").Length() > 0 {
				tables := currentNode
				if goquery.NodeName(currentNode) != "table" {
					tables = currentNode.Find("table")
				}

				tables.Each(func(j int, table *goquery.Selection) {
					table.Find("tr").Each(func(k int, row *goquery.Selection) {
						if k == 0 {
							return // Skip header row
						}

						cells := row.Find("td")
						if cells.Length() >= 1 {
							elementName := strings.TrimSpace(cells.First().Text())
							if elementName == "" && cells.Length() > 1 {
								elementName = strings.TrimSpace(cells.Eq(1).Text())
							}

							if elementName != "" {
								elementsInThisTier = append(elementsInThisTier, elementName)
								normalizedName := normalizeName(elementName)
								elementTiers[normalizedName] = currentTier
								fmt.Printf("  📦 [table] %-30s → Tier %d\n", elementName, currentTier)
							}
						}
					})
				})
			}

			// Look for divs that might contain lists of elements
			currentNode.Find("h3").Each(func(j int, elem *goquery.Selection) {
				elementName := strings.TrimSpace(elem.Text())
				if elementName != "" {
					elementsInThisTier = append(elementsInThisTier, elementName)
					normalizedName := normalizeName(elementName)
					elementTiers[normalizedName] = currentTier
					fmt.Printf("  📦 [div>h3] %-30s → Tier %d\n", elementName, currentTier)
				}
			})

			// Move to next sibling
			currentNode = currentNode.Next()
		}

		fmt.Printf("Found %d elements in Tier %d\n", len(elementsInThisTier), currentTier)
	}

	// STEP 3: Alternative approach - Use the content div directly and track transitions
	// between tier sections
	currentTier := 0
	inTierSection := false

	fmt.Println("\n==== DIRECT CONTENT DIV APPROACH ====")

	doc.Find(".mw-parser-output").Children().Each(func(i int, node *goquery.Selection) {
		// Check if this is a tier heading
		if goquery.NodeName(node) == "h2" {
			headline := node.Find(".mw-headline")
			if headline.Length() > 0 {
				id, exists := headline.Attr("id")
				if exists {
					tierMatch := regexp.MustCompile(`Tier_(\d+)_elements`).FindStringSubmatch(id)
					if len(tierMatch) >= 2 {
						tierNum, _ := strconv.Atoi(tierMatch[1])
						currentTier = tierNum
						inTierSection = true
						fmt.Printf("\n--- Entering Tier %d Section ---\n", currentTier)
						return
					}
				}
			}

			// If it's any other h2, we're no longer in a tier section
			if inTierSection {
				fmt.Printf("\n--- Exiting Tier Section ---\n")
				inTierSection = false
			}
			return
		}

		// Process elements only if we're in a tier section
		if !inTierSection || currentTier == 0 {
			return
		}

		// Process h3 elements (element names)
		if goquery.NodeName(node) == "h3" {
			elementName := strings.TrimSpace(node.Text())
			if elementName != "" {
				normalizedName := normalizeName(elementName)
				elementTiers[normalizedName] = currentTier
				fmt.Printf("  📦 [direct h3] %-30s → Tier %d\n", elementName, currentTier)
			}
			return
		}

		// Process tables
		node.Find("table").Each(func(j int, table *goquery.Selection) {
			table.Find("tr").Each(func(k int, row *goquery.Selection) {
				if k == 0 {
					return // Skip header row
				}

				cells := row.Find("td")
				if cells.Length() >= 1 {
					elementName := strings.TrimSpace(cells.First().Text())
					if elementName == "" && cells.Length() > 1 {
						elementName = strings.TrimSpace(cells.Eq(1).Text())
					}

					if elementName != "" {
						normalizedName := normalizeName(elementName)
						elementTiers[normalizedName] = currentTier
						fmt.Printf("  📦 [direct table] %-30s → Tier %d\n", elementName, currentTier)
					}
				}
			})
		})

		// Process h3 elements inside divs
		node.Find("h3").Each(func(j int, elem *goquery.Selection) {
			elementName := strings.TrimSpace(elem.Text())
			if elementName != "" {
				normalizedName := normalizeName(elementName)
				elementTiers[normalizedName] = currentTier
				fmt.Printf("  📦 [direct div>h3] %-30s → Tier %d\n", elementName, currentTier)
			}
		})
	})

	// Print tier distribution statistics
	tierCounts := make(map[int]int)
	for _, tier := range elementTiers {
		tierCounts[tier]++
	}

	fmt.Println("\n------ Tier Distribution ------")
	for tier, count := range tierCounts {
		fmt.Printf("Tier %d: %d elements\n", tier, count)
	}
	fmt.Println("------------------------------\n")

	// Continue with your existing image extraction logic
	elementImages := make(map[string]string)

	doc.Find(".wikia-gallery-item, .wikia-gallery-caption, .gallery-image-wrapper").Each(func(i int, item *goquery.Selection) {
		caption := item.Find(".wikia-gallery-caption, .gallery-image-caption")
		elemName := strings.TrimSpace(caption.Text())

		if elemName == "" {
			return
		}

		elemName = strings.TrimSpace(strings.Split(elemName, "\n")[0])
		img := item.Find("img")
		if img.Length() > 0 {
			imageURL := ""
			if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
				imageURL = dataSrc
			} else if srcSet, exists := img.Attr("srcset"); exists {
				srcSetParts := strings.Split(srcSet, ", ")
				if len(srcSetParts) > 0 {
					lastSrc := srcSetParts[len(srcSetParts)-1]
					urlParts := strings.Split(lastSrc, " ")
					if len(urlParts) > 0 {
						imageURL = urlParts[0]
					}
				}
			} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
				imageURL = src
			}

			if imageURL != "" {
				fmt.Printf("Found image for element %s: %s\n", elemName, imageURL)
				elementImages[elemName] = imageURL
			}
		}
	})

	// Extract images from tables
	doc.Find("table.article-table tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() >= 2 {
			firstCell := cells.First()
			elemName := strings.TrimSpace(firstCell.Text())

			if elemName == "" && cells.Length() > 1 {
				elemName = strings.TrimSpace(cells.Eq(1).Text())
			}

			if elemName == "" {
				return
			}
			img := firstCell.Find("img")

			if img.Length() > 0 {
				imageURL := ""
				if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
					imageURL = dataSrc
				} else if srcSet, exists := img.Attr("srcset"); exists {
					srcSetParts := strings.Split(srcSet, ", ")
					if len(srcSetParts) > 0 {
						lastSrc := srcSetParts[len(srcSetParts)-1]
						urlParts := strings.Split(lastSrc, " ")
						if len(urlParts) > 0 {
							imageURL = urlParts[0]
						}
					}
				} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
					imageURL = src
				}

				if imageURL != "" {
					fmt.Printf("Found image from table for element %s: %s\n", elemName, imageURL)
					elementImages[elemName] = imageURL
				}
			}
		}
	})

	// Parse recipes
	results := []ElementRecipe{}

	doc.Find(".mw-parser-output").Each(func(i int, content *goquery.Selection) {
		content.Find("h2, h3").Each(func(j int, header *goquery.Selection) {
			headerText := strings.TrimSpace(header.Text())

			if headerText == "" || strings.Contains(headerText, "See also") ||
				strings.Contains(headerText, "Navigation") || strings.Contains(headerText, "Contents") {
				return
			}

			recipeList := header.NextUntil("h2, h3").FilterFunction(func(i int, s *goquery.Selection) bool {
				return goquery.NodeName(s) == "ul"
			})
			if recipeList.Length() == 0 {
				recipeList = header.NextUntil("h2, h3").Find("ul")
			}

			if recipeList.Length() == 0 {
				return
			}

			recipes := [][2]string{}
			seen := map[string]bool{}

			recipeList.Find("li").Each(func(k int, li *goquery.Selection) {
				links := li.Find("a")
				if links.Length() >= 2 {
					a := strings.TrimSpace(links.Eq(0).Text())
					b := strings.TrimSpace(links.Eq(1).Text())

					if a != "" && b != "" {
						key := a + "+" + b
						if !seen[key] {
							recipes = append(recipes, [2]string{a, b})
							seen[key] = true
						}
					}
				} else {
					text := strings.TrimSpace(li.Text())
					parts := strings.Split(text, "+")

					if len(parts) == 2 {
						a := strings.TrimSpace(parts[0])
						b := strings.TrimSpace(parts[1])

						if a != "" && b != "" {
							key := a + "+" + b
							if !seen[key] {
								recipes = append(recipes, [2]string{a, b})
								seen[key] = true
							}
						}
					}
				}
			})

			if len(recipes) > 0 {
				imageURL := elementImages[headerText]
				// Use the normalized name to look up the tier
				tier := elementTiers[normalizeName(headerText)]

				// Log for debugging
				fmt.Printf("⚗️ Element: %-30s → Found tier: %d\n", headerText, tier)

				results = append(results, ElementRecipe{
					Element:  headerText,
					Recipes:  recipes,
					ImageURL: imageURL,
					Tier:     tier,
				})
			}
		})
	})

	// If no results found from headers, try tables
	if len(results) == 0 {
		doc.Find("table").Each(func(i int, table *goquery.Selection) {
			table.Find("tr").Each(func(j int, row *goquery.Selection) {
				if j == 0 {
					return
				}
				cells := row.Find("td")
				if cells.Length() >= 2 {
					element := strings.TrimSpace(cells.Eq(0).Text())
					recipesText := strings.TrimSpace(cells.Eq(1).Text())

					if element == "" || recipesText == "" {
						return
					}

					recipes := [][2]string{}
					seen := map[string]bool{}

					for _, recipeText := range strings.Split(recipesText, "\n") {
						parts := strings.Split(recipeText, "+")
						if len(parts) == 2 {
							a := strings.TrimSpace(parts[0])
							b := strings.TrimSpace(parts[1])

							if a != "" && b != "" {
								key := a + "+" + b
								if !seen[key] {
									recipes = append(recipes, [2]string{a, b})
									seen[key] = true
								}
							}
						}
					}

					if len(recipes) > 0 {
						imageURL := elementImages[element]
						if imageURL == "" {
							img := cells.Eq(0).Find("img")
							if img.Length() > 0 {
								if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
									imageURL = dataSrc
								} else if srcSet, exists := img.Attr("srcset"); exists {
									srcSetParts := strings.Split(srcSet, ", ")
									if len(srcSetParts) > 0 {
										lastSrc := srcSetParts[len(srcSetParts)-1]
										urlParts := strings.Split(lastSrc, " ")
										if len(urlParts) > 0 {
											imageURL = urlParts[0]
										}
									}
								} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
									imageURL = src
								}
							}
						}

						// Use the normalized name to look up the tier
						tier := elementTiers[normalizeName(element)]

						// Log for debugging
						fmt.Printf("⚗️ Element (table): %-30s → Found tier: %d\n", element, tier)

						results = append(results, ElementRecipe{
							Element:  element,
							Recipes:  recipes,
							ImageURL: imageURL,
							Tier:     tier,
						})
					}
				}
			})
		})
	}
	// Tambahkan elemen Earth dan Time secara manual
	manualBasics := []ElementRecipe{
		{
			Element:  "Earth",
			ImageURL: "https://static.wikia.nocookie.net/little-alchemy/images/2/21/Earth_2.svg/revision/latest?cb=20210827132928",
			Recipes:  [][2]string{},
			Tier:     0,
		},
		{
			Element:  "Time",
			ImageURL: "https://static.wikia.nocookie.net/little-alchemy/images/6/63/Time_2.svg/revision/latest?cb=20210827124225",
			Recipes:  [][2]string{},
			Tier:     0,
		},
	}

	results = append(results, manualBasics...)

	f, err := os.Create("recipes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(results)

	fmt.Printf("✅ Selesai dalam %s, total elemen: %d\n", time.Since(start), len(results))
}
