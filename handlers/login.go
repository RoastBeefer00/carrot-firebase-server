package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/RoastBeefer00/carrot-firebase-server/views"
)

func Login(c echo.Context) error {
	state := GetStateFromContext(c)

	log.Printf("Logging in user %s with email %s", state.User.DisplayName, state.User.Email)
	return Render(c, http.StatusOK, views.User(state.User))
}
