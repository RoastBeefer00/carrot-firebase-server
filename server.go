package main

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
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
	index := views.Index()

	e := echo.New()
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

    store := sessions.NewFilesystemStore(
		"",
		authKeyOne,
		encryptionKeyOne,
	)

    store.MaxLength(10000000)

	e.Use(session.Middleware(store))
	gob.Register(services.Recipe{})
	gob.Register([]services.Recipe{})

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
	e.GET("/login", func(c echo.Context) error {
		token := c.QueryParam("token")
		displayName := c.QueryParam("displayName")
		email := c.QueryParam("email")
		uid := c.QueryParam("uid")
		fmt.Println("Token: " + token)
		fmt.Println("DisplayName: " + displayName)
		fmt.Println("Email: " + email)
		fmt.Println("Uid: " + uid)

		user := services.User{
			Token:       token,
			DisplayName: displayName,
			Email:       email,
			Uid:         uid,
		}

		sess, _ := session.Get(uid, c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			HttpOnly: true,
		}
		sess.Values["foo"] = "bar"
		sess.Save(c.Request(), c.Response())
		return Render(c, http.StatusOK, views.Page(user))
	})
	e.GET("/recipes", handlers.GetAllRecipes)
	e.POST("/recipes/replace/:id", handlers.ReplaceRecipe)
	// e.GET("/recipes/random", handlers.GetRandomRecipe)
	e.GET("/recipes/random", handlers.GetRandomRecipes)
	e.GET("/recipes/name", handlers.SearchRecipesByName)
	e.GET("/recipes/ingredients", handlers.SearchRecipesByIngredient)
	e.GET("/recipes/filter", handlers.ChangeFilter)
	e.DELETE("/recipes/delete/:id", handlers.DeleteRecipe)
	e.DELETE("/recipes/delete/all", handlers.DeleteAllRecipes)
	e.GET("/groceries", handlers.CombineIngredients)

	e.Logger.Fatal(e.Start(":8080"))
}
