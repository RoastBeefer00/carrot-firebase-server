package handlers

import (
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

// CombineIngredients renders the grocery list drawer.
func CombineIngredients(c echo.Context) error {
	state := GetStateFromContext(c)
	return Render(c, http.StatusOK, views.GroceryDrawer(groceryItems(state), groceryMeals(state)))
}
