package handlers

import (
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

var (
	quantityRE    = regexp.MustCompile(`^\d*[^a-zA-Z \*]?\d*`)
	measurementRE = regexp.MustCompile(`tbsps?|tsps?|cups?|cans?|packages?|packets?|ozs?|pounds?`)
	itemRE        = regexp.MustCompile(`\*?[a-zA-Z].*`)
)

func getIngredientQuantity(ingredient string) string {
	return quantityRE.FindString(ingredient)
}

func getIngredientMeasurement(ingredient string) string {
	return measurementRE.FindString(ingredient)
}

func getIngredientItem(ingredient string) string {
	measurement := getIngredientMeasurement(ingredient)
	if measurement == "" {
		return itemRE.FindString(ingredient)
	}
	return itemRE.FindString(strings.ReplaceAll(ingredient, measurement, ""))
}

func CombineIngredients(c echo.Context) error {
	state := GetStateFromContext(c)

	recipes := state.Recipes
	var ingredients []services.Ingredient

	for _, recipe := range recipes {
		for _, ing := range recipe.Ingredients {
			ingredients = append(ingredients, services.Ingredient{
				Quantity:    getIngredientQuantity(ing),
				Measurement: getIngredientMeasurement(ing),
				Item:        getIngredientItem(ing),
			})
		}
	}

	sort.Slice(ingredients, func(i, j int) bool {
		return strings.ToLower(ingredients[i].Item) < strings.ToLower(ingredients[j].Item)
	})
	i := 0
	max := len(ingredients) - 2

	for i < max {
		j := i + 1
		ingredientI := ingredients[i]
		ingredientJ := ingredients[j]

		if strings.Contains(ingredientI.Item, ingredientJ.Item) ||
			strings.Contains(ingredientJ.Item, ingredientI.Item) {
			if (ingredientI.Measurement == "" && ingredientJ.Measurement == "") ||
				((strings.Contains(ingredientI.Measurement, ingredientJ.Measurement) && ingredientJ.Measurement != "") || (strings.Contains(ingredientJ.Measurement, ingredientI.Measurement) && ingredientI.Measurement != "")) {
				insert := services.Ingredient{}
				if ingredientI.Quantity != "" {
					if len(ingredientI.Quantity) > len(ingredientJ.Quantity) {
						insert.Item = ingredientJ.Item
					} else {
						insert.Item = ingredientI.Item
					}

					if len(ingredientI.Measurement) > len(ingredientJ.Measurement) {
						insert.Measurement = ingredientJ.Measurement
					} else {
						insert.Measurement = ingredientI.Measurement
					}

					added := services.ParseQuantity(ingredientI.Quantity) + services.ParseQuantity(ingredientJ.Quantity)
					insert.Quantity = services.FormatDecimal(added)
				}
				ingredients = append(ingredients[:j], ingredients[j+1:]...)
				if insert.Item != "" {
					ingredients = slices.Insert(ingredients, j, insert)
					ingredients = append(ingredients[:i], ingredients[i+1:]...)
				}
				i--
				max--
			}
		}
		i++
	}

	// services.AllIngredients = ingredients
	return Render(c, http.StatusOK, views.Groceries(ingredients))
}
