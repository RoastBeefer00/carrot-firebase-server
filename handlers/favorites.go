package handlers

import (
	"log"
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

// Favorites now lives on the recipes page; this keeps old links working.
func Favorites(c echo.Context) error {
	return c.Redirect(http.StatusFound, views.RecipesURL("fav", ""))
}

func ToggleFavorite(c echo.Context) error {
	state := GetStateFromContext(c)
	id := c.Param("id")

	if state.IsFavorite(id) {
		log.Printf("Removing favorite %s for %s", id, state.User.Email)
		state.DeleteFavorite(id)
	} else {
		log.Printf("Adding favorite %s for %s", id, state.User.Email)
		state.AddFavorite(id)
	}

	return saveAndRespond(c, state, viewFrom(c), "")
}
