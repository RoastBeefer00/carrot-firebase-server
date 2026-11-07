package middlewares

import (
	"log"
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/labstack/echo/v4"
)

// TimeZoneCookie holds the browser's time zone, set by dist/app.js, so "today"
// and "this week" follow the user's clock rather than the server's.
const TimeZoneCookie = "tz"

// Middleware that loads user state and adds it to the context, along with
// today's date in the user's time zone.
func StateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		state, err := db.GetState(c)
		if err != nil {
			// User is not logged in
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		zone := ""
		if cookie, err := c.Cookie(TimeZoneCookie); err == nil {
			zone = cookie.Value
		}
		today := services.Today(zone)

		// Plans saved before days had dates are moved onto this week, once.
		if state.MigrateLegacyDays(services.WeekStart(today)) {
			if err := db.UpdateState(state, c); err != nil {
				log.Printf("Failed to save migrated days for %s: %v", state.User.Email, err)
			}
		}

		// Store state in context for handlers to access
		c.Set("state", state)
		c.Set("today", today)
		return next(c)
	}
}
