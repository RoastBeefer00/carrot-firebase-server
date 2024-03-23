package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Ingredient struct {
	Quantity    string `json:"quantity"`
	Measurement string `json:"measurement"`
	Item        string `json:"item"`
}

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

func CombineIngredients(w http.ResponseWriter, r *http.Request) {
    var recipes []Recipe
	var ingredients []Ingredient

    err := json.NewDecoder(r.Body).Decode(&recipes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	for _, recipe := range recipes {
		for _, ing := range recipe.Ingredients {
			quantity, _ := getIngredientQuantity(ing)
			measurement, _ := getIngredientMeasurement(ing)
			item, _ := getIngredientItem(ing)
			ingredient := Ingredient{
				Quantity:    quantity,
				Measurement: measurement,
				Item:        item,
			}

			ingredients = append(ingredients, ingredient)
		}
	}

	sort.Slice(ingredients, func(i, j int) bool { return ingredients[i].Item < ingredients[j].Item })
	i := 0
	max := len(ingredients) - 2

	for i < max {
		j := i + 1
		ingredientI := ingredients[i]
		ingredientJ := ingredients[j]

		if strings.Contains(ingredientI.Item, ingredientJ.Item) || strings.Contains(ingredientJ.Item, ingredientI.Item) {
			insert := Ingredient{}
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

				if strings.Contains(ingredientI.Quantity, ".") || strings.Contains(ingredientJ.Quantity, ".") {
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

				ingredients = append(ingredients[:j], ingredients[j+1:]...)
				if insert.Item != "" {
					ingredients = slices.Insert(ingredients, j, insert)
					ingredients = append(ingredients[:i], ingredients[i+1:]...)
				}
				i++
				max--
			}
		}
		i++
	}

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(ingredients)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
