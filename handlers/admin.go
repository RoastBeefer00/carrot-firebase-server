package handlers

import (
	"log"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"

	"github.com/RoastBeefer00/carrot-firebase-server/database"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
)

func AdminHandler(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

    header := c.Request().Header
    log.Println(header)
    log.Println(header["Hx-Request"] == nil)


    if slices.Contains(services.Admins, state.User.Email) {
        if header["Hx-Request"] == nil {
            return Render(c, http.StatusOK, views.Index(true))
        } else {
            return Render(c, http.StatusOK, views.Page(true))
        }
    } else {
        return c.NoContent(403)
    }
}
