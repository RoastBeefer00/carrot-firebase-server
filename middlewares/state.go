package middlewares

import (
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/labstack/echo/v4"
)

// Middleware that loads user state and adds it to the context
func StateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		state, err := db.GetState(c)
		if err != nil {
			// Handle error appropriately - redirect to login, return error page, etc.
			// return c.String(http.StatusUnauthorized, "Unauthorized")
			// User is not logged in
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		// Store state in context for handlers to access
		c.Set("state", state)
		return next(c)
	}
}
