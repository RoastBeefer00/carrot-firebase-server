package handlers

import (
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

func CombineIngredients(c echo.Context) error {
	state := GetStateFromContext(c)

	var lines []string
	for _, recipe := range state.Recipes {
		if recipe.SkipGroceries {
			continue
		}
		lines = append(lines, recipe.Ingredients...)
	}

	return Render(c, http.StatusOK, views.Groceries(services.CombineIngredients(lines)))
}
