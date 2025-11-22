package handlers

import (
	"cloud.google.com/go/firestore"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/labstack/echo/v4"
)

// Helper function to retrieve state from context in your handlers
func GetStateFromContext(c echo.Context) services.State {
	return c.Get("state").(services.State)
}

// Helper to get client from context
func GetDbClient(c echo.Context) *firestore.Client {
	return c.Get("db").(*firestore.Client)
}
