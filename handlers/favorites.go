package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

func Favorites(c echo.Context) error {
	ctx := c.Request().Context()
	state := GetStateFromContext(c)
	client := GetDbClient(c)

	var favorites []services.Recipe
	var wg sync.WaitGroup
	for _, id := range state.Favorites {
		wg.Add(1)
		go func(id string, client *firestore.Client, ctx context.Context) {
			defer wg.Done()
			var recipe services.Recipe
			doc, err := client.Collection("recipes").Doc(id).Get(ctx)
			if err != nil {
				log.Printf("Failed to get recipe: %s", err)
				return
			}

			err = doc.DataTo(&recipe)
			if err != nil {
				log.Printf("Failed to get recipe: %s", err)
				return
			}

			recipe.Favorite = true
			favorites = append(favorites, recipe)
		}(id, client, ctx)
	}
	wg.Wait()

	header := c.Request().Header

	if header["Hx-Request"] == nil {
		return Render(c, http.StatusOK, views.Index(views.Favorites(favorites), state))
	}

	return Render(c, http.StatusOK, views.Favorites(favorites))
}

func AddFavorite(c echo.Context) error {
	state := GetStateFromContext(c)

	id := c.Param("id")

	log.Printf(
		"Adding favorite recipe with id %s for user %s with email %s",
		id,
		state.User.DisplayName,
		state.User.Email,
	)
	err := Render(c, http.StatusOK, views.FavoriteButton(true, id))
	if err != nil {
		return err
	}

	state.AddFavorite(id)
	return db.UpdateState(state, c)
}

func DeleteFavorite(c echo.Context) error {
	state := GetStateFromContext(c)

	id := c.Param("id")

	log.Printf(
		"Removing favorite recipe with id %s for user %s with email %s",
		id,
		state.User.DisplayName,
		state.User.Email,
	)
	err := Render(c, http.StatusOK, views.FavoriteButton(false, id))
	if err != nil {
		return err
	}

	state.DeleteFavorite(id)
	return db.UpdateState(state, c)
}
