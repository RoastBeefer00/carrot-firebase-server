package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/labstack/echo/v4"
)

func Favorites(c echo.Context) error {
	ctx := c.Request().Context()
	state := GetStateFromContext(c)
	client := GetDbClient(c)

	results := make([]services.Recipe, len(state.Favorites))
	errs := make([]error, len(state.Favorites))
	var wg sync.WaitGroup
	for i, id := range state.Favorites {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			doc, err := client.Collection("recipes").Doc(id).Get(ctx)
			if err != nil {
				errs[i] = err
				return
			}
			if err := doc.DataTo(&results[i]); err != nil {
				errs[i] = err
				return
			}
			results[i].Favorite = true
		}(i, id)
	}
	wg.Wait()

	favorites := make([]services.Recipe, 0, len(results))
	for i, r := range results {
		if errs[i] != nil {
			log.Printf("Failed to get favorite %s: %v", state.Favorites[i], errs[i])
			continue
		}
		favorites = append(favorites, r)
	}

	if c.Request().Header.Get("Hx-Request") == "" {
		return Render(c, http.StatusOK, views.Index(views.Favorites(favorites), state, "favorites"))
	}

	return Render(c, http.StatusOK, views.Favorites(favorites))
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

	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
