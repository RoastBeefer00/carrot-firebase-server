package handlers

import (
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

func getIngredientQuantity(ingredient string) (string, error) {
	r, err := regexp.Compile("^\\d*[^a-zA-Z \\*]?\\d*")
	if err != nil {
		return "", err
	}

	match := r.FindString(ingredient)

	return match, nil
}

func getIngredientMeasurement(ingredient string) (string, error) {
	r, err := regexp.Compile("tbsps?|tsps?|cups?|cans?|packages?|packets?|ozs?|pounds?")
	if err != nil {
		return "", err
	}

	match := r.FindString(ingredient)

	return match, nil
}

func getIngredientItem(ingredient string) (string, error) {
	measurement, _ := getIngredientMeasurement(ingredient)
	if measurement == "" {
		r, err := regexp.Compile("\\*?[a-zA-Z].*")
		if err != nil {
			return "", err
		}

		match := r.FindString(ingredient)
		return match, nil
	} else {
		ing_wo_measurement := strings.ReplaceAll(ingredient, measurement, "")

		r, err := regexp.Compile("\\*?[a-zA-Z].*")
		if err != nil {
			return "", err
		}

		match := r.FindString(ing_wo_measurement)
		return match, nil
	}
}

func CombineIngredients(c echo.Context) error {
	state := GetStateFromContext(c)

	recipes := state.Recipes
	var ingredients []services.Ingredient

	for _, recipe := range recipes {
		for _, ing := range recipe.Ingredients {
			quantity, _ := getIngredientQuantity(ing)
			measurement, _ := getIngredientMeasurement(ing)
			item, _ := getIngredientItem(ing)
			ingredient := services.Ingredient{
				Quantity:    quantity,
				Measurement: measurement,
				Item:        item,
			}

			ingredients = append(ingredients, ingredient)
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

					if strings.Contains(ingredientI.Quantity, ".") ||
						strings.Contains(ingredientJ.Quantity, ".") {
						floatI, _ := strconv.ParseFloat(ingredientI.Quantity, 64)
						floatJ, _ := strconv.ParseFloat(ingredientJ.Quantity, 64)
						added := floatI + floatJ
						insert.Quantity = strconv.FormatFloat(added, 'f', 2, 64)
					} else {
						intI, _ := strconv.Atoi(ingredientI.Quantity)
						intJ, _ := strconv.Atoi(ingredientJ.Quantity)
						added := intI + intJ
						insert.Quantity = strconv.Itoa(added)
					}
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
