package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
)

//go:generate templ generate
//go:generate npm i
//go:generate npx tailwindcss -i ./dist/main.css -o ./dist/tailwind.css

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func main() {
	index := views.Index(false)

	e := echo.New()

	e.Static("/dist", "dist")
	// Little bit of middlewares for housekeeping
	e.Use(middleware.Logger())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))

	// This will initiate our template renderer
	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, index)
	})
    e.GET("/login", handlers.Login)
    e.GET("/admin", handlers.AdminHandler)
    e.GET("/refresh", handlers.GetRecipes)
	e.GET("/recipes/replace/:id", handlers.ReplaceRecipe)
	e.GET("/recipes/random", handlers.GetRandomRecipes)
	e.GET("/recipes/add", handlers.AddRecipeToDatabase)
	e.GET("/recipes/name", handlers.SearchRecipesByName)
	e.GET("/recipes/ingredients", handlers.SearchRecipesByIngredient)
	e.GET("/recipes/filter", handlers.ChangeFilter)
	e.GET("/recipes/delete/:id", handlers.DeleteRecipe)
	e.GET("/recipes/delete/all", handlers.DeleteAllRecipes)
	e.GET("/groceries", handlers.CombineIngredients)
    e.GET("/ingredient/add/:id", handlers.AddIngredient)
    e.GET("/ingredient/delete/:id", handlers.DeleteIngredient)
    e.GET("/step/add/:id", handlers.AddStep)
    e.GET("/step/delete/:id", handlers.DeleteStep)

	e.Logger.Fatal(e.Start(":8080"))
}
