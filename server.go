package main

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

    "github.com/RoastBeefer00/carrot-firebase-server/views"
    "github.com/RoastBeefer00/carrot-firebase-server/handlers"
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
    index := views.Index()

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
    e.GET("/recipes", handlers.GetAllRecipes)
    e.POST("/recipes/replace/:id", handlers.ReplaceRecipe)
    e.GET("/recipes/random", handlers.GetRandomRecipe)
    e.POST("/recipes/random", handlers.GetRandomRecipesPost)
    e.GET("/recipes/random/:amount", handlers.GetRandomRecipes)
    e.GET("/recipes/name", handlers.SearchRecipesByName)
    e.GET("/recipes/ingredients", handlers.SearchRecipesByIngredient)
    e.GET("/recipes/filter", handlers.ChangeFilter)
    e.DELETE("/recipes/delete/:id", handlers.DeleteRecipe)
    e.DELETE("/recipes/delete/all", handlers.DeleteAllRecipes)
    e.GET("/groceries", handlers.CombineIngredients)

	e.Logger.Fatal(e.Start(":8080"))
}
