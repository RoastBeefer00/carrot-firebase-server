package middlewares

import (
	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
)

func DatabaseMiddleware(client *firestore.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", client)
			return next(c)
		}
	}
}
