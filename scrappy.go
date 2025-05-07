package main

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ElementRecipe struct {
	Element  string      `json:"element"`
	ImageURL string      `json:"image_url"`
	Recipes  [][2]string `json:"recipes"`
}

const url = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

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

// func main() {
// 	start := time.Now()
// 	doc, err := getDoc()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	elementImages := make(map[string]string)

// 	doc.Find(".wikia-gallery-item, .wikia-gallery-caption, .gallery-image-wrapper").Each(func(i int, item *goquery.Selection) {
// 		caption := item.Find(".wikia-gallery-caption, .gallery-image-caption")
// 		elemName := strings.TrimSpace(caption.Text())

// 		if elemName == "" {
// 			return
// 		}

// 		elemName = strings.TrimSpace(strings.Split(elemName, "\n")[0])
// 		img := item.Find("img")
// 		if img.Length() > 0 {
// 			imageURL := ""
// 			if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
// 				imageURL = dataSrc
// 			} else if srcSet, exists := img.Attr("srcset"); exists {
// 				srcSetParts := strings.Split(srcSet, ", ")
// 				if len(srcSetParts) > 0 {
// 					lastSrc := srcSetParts[len(srcSetParts)-1]
// 					urlParts := strings.Split(lastSrc, " ")
// 					if len(urlParts) > 0 {
// 						imageURL = urlParts[0]
// 					}
// 				}
// 			} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
// 				imageURL = src
// 			}

// 			if imageURL != "" {
// 				fmt.Printf("Found image for element %s: %s\n", elemName, imageURL)
// 				elementImages[elemName] = imageURL
// 			}
// 		}
// 	})

// 	doc.Find("table.article-table tr").Each(func(i int, row *goquery.Selection) {
// 		cells := row.Find("td")
// 		if cells.Length() >= 2 {
// 			firstCell := cells.First()
// 			elemName := strings.TrimSpace(firstCell.Text())

// 			if elemName == "" && cells.Length() > 1 {
// 				elemName = strings.TrimSpace(cells.Eq(1).Text())
// 			}

// 			if elemName == "" {
// 				return
// 			}
// 			img := firstCell.Find("img")

// 			if img.Length() > 0 {
// 				imageURL := ""
// 				if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
// 					imageURL = dataSrc
// 				} else if srcSet, exists := img.Attr("srcset"); exists {
// 					srcSetParts := strings.Split(srcSet, ", ")
// 					if len(srcSetParts) > 0 {
// 						lastSrc := srcSetParts[len(srcSetParts)-1]
// 						urlParts := strings.Split(lastSrc, " ")
// 						if len(urlParts) > 0 {
// 							imageURL = urlParts[0]
// 						}
// 					}
// 				} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
// 					imageURL = src
// 				}

// 				if imageURL != "" {
// 					fmt.Printf("Found image from table for element %s: %s\n", elemName, imageURL)
// 					elementImages[elemName] = imageURL
// 				}
// 			}
// 		}
// 	})

// 	results := []ElementRecipe{}

// 	doc.Find(".mw-parser-output").Each(func(i int, content *goquery.Selection) {
// 		content.Find("h2, h3").Each(func(j int, header *goquery.Selection) {
// 			headerText := strings.TrimSpace(header.Text())

// 			if headerText == "" || strings.Contains(headerText, "See also") ||
// 				strings.Contains(headerText, "Navigation") || strings.Contains(headerText, "Contents") {
// 				return
// 			}

// 			recipeList := header.NextUntil("h2, h3").FilterFunction(func(i int, s *goquery.Selection) bool {
// 				return goquery.NodeName(s) == "ul"
// 			})
// 			if recipeList.Length() == 0 {
// 				recipeList = header.NextUntil("h2, h3").Find("ul")
// 			}

// 			if recipeList.Length() == 0 {
// 				return
// 			}

// 			recipes := [][2]string{}
// 			seen := map[string]bool{}

// 			recipeList.Find("li").Each(func(k int, li *goquery.Selection) {
// 				links := li.Find("a")
// 				if links.Length() >= 2 {
// 					a := strings.TrimSpace(links.Eq(0).Text())
// 					b := strings.TrimSpace(links.Eq(1).Text())

// 					if a != "" && b != "" {
// 						key := a + "+" + b
// 						if !seen[key] {
// 							recipes = append(recipes, [2]string{a, b})
// 							seen[key] = true
// 						}
// 					}
// 				} else {
// 					text := strings.TrimSpace(li.Text())
// 					parts := strings.Split(text, "+")

// 					if len(parts) == 2 {
// 						a := strings.TrimSpace(parts[0])
// 						b := strings.TrimSpace(parts[1])

// 						if a != "" && b != "" {
// 							key := a + "+" + b
// 							if !seen[key] {
// 								recipes = append(recipes, [2]string{a, b})
// 								seen[key] = true
// 							}
// 						}
// 					}
// 				}
// 			})

// 			if len(recipes) > 0 {
// 				imageURL := elementImages[headerText]

// 				results = append(results, ElementRecipe{
// 					Element:  headerText,
// 					Recipes:  recipes,
// 					ImageURL: imageURL,
// 				})
// 			}
// 		})
// 	})

// 	if len(results) == 0 {
// 		doc.Find("table").Each(func(i int, table *goquery.Selection) {
// 			table.Find("tr").Each(func(j int, row *goquery.Selection) {
// 				if j == 0 {
// 					return
// 				}
// 				cells := row.Find("td")
// 				if cells.Length() >= 2 {
// 					element := strings.TrimSpace(cells.Eq(0).Text())
// 					recipesText := strings.TrimSpace(cells.Eq(1).Text())

// 					if element == "" || recipesText == "" {
// 						return
// 					}

// 					recipes := [][2]string{}
// 					seen := map[string]bool{}

// 					for _, recipeText := range strings.Split(recipesText, "\n") {
// 						parts := strings.Split(recipeText, "+")
// 						if len(parts) == 2 {
// 							a := strings.TrimSpace(parts[0])
// 							b := strings.TrimSpace(parts[1])

// 							if a != "" && b != "" {
// 								key := a + "+" + b
// 								if !seen[key] {
// 									recipes = append(recipes, [2]string{a, b})
// 									seen[key] = true
// 								}
// 							}
// 						}
// 					}

// 					if len(recipes) > 0 {
// 						imageURL := elementImages[element]
// 						if imageURL == "" {
// 							img := cells.Eq(0).Find("img")
// 							if img.Length() > 0 {
// 								if dataSrc, exists := img.Attr("data-src"); exists && !strings.Contains(dataSrc, "data:image/gif;base64") {
// 									imageURL = dataSrc
// 								} else if srcSet, exists := img.Attr("srcset"); exists {
// 									srcSetParts := strings.Split(srcSet, ", ")
// 									if len(srcSetParts) > 0 {
// 										lastSrc := srcSetParts[len(srcSetParts)-1]
// 										urlParts := strings.Split(lastSrc, " ")
// 										if len(urlParts) > 0 {
// 											imageURL = urlParts[0]
// 										}
// 									}
// 								} else if src, exists := img.Attr("src"); exists && !strings.Contains(src, "data:image/gif;base64") {
// 									imageURL = src
// 								}
// 							}
// 						}

// 						results = append(results, ElementRecipe{
// 							Element:  element,
// 							Recipes:  recipes,
// 							ImageURL: imageURL,
// 						})
// 					}
// 				}
// 			})
// 		})
// 	}
// 	f, err := os.Create("recipes.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()

// 	enc := json.NewEncoder(f)
// 	enc.SetIndent("", "  ")
// 	enc.Encode(results)

// 	fmt.Printf("✅ Selesai dalam %s, total elemen: %d\n", time.Since(start), len(results))
// }
