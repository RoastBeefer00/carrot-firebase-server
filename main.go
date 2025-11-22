package main

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
	"github.com/RoastBeefer00/carrot-firebase-server/middlewares"
)

//go:generate templ generate
//go:generate tailwindcss -i ./dist/main.css -o ./dist/tailwind.css

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func main() {
	// --- Load configuration from .env ---
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, falling back to environment variables")
	}

	// --- Read environment variables ---
	handlers.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	handlers.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	handlers.RedirectURL = os.Getenv("REDIRECT_URL")
	handlers.EncryptionKeyBase64 = os.Getenv("ENCRYPTION_KEY_BASE64")
	e := echo.New()

	// --- Validate Configuration ---
	if handlers.GoogleClientID == "" || handlers.GoogleClientSecret == "" ||
		handlers.RedirectURL == "" ||
		handlers.EncryptionKeyBase64 == "" {
		log.Fatal(
			"Missing required environment variables (GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, REDIRECT_URL, ENCRYPTION_KEY_BASE64)",
		)
	}

	// --- Decode Encryption Key ---
	db.EncryptionKey, err = base64.URLEncoding.DecodeString(handlers.EncryptionKeyBase64)
	if err != nil {
		log.Fatalf("Failed to decode encryption key: %v", err)
	}
	if len(db.EncryptionKey) != 32 { // AES-256 requires a 32-byte key
		log.Fatalf("Encryption key must be 32 bytes long (decoded)")
	}

	// --- Initialize OAuth Config ---
	handlers.OauthConfig = &oauth2.Config{
		ClientID:     handlers.GoogleClientID,
		ClientSecret: handlers.GoogleClientSecret,
		RedirectURL:  handlers.RedirectURL,
		Scopes:       []string{"openid", "profile", "email"}, // Request basic user info
		Endpoint:     google.Endpoint,
	}

	e.Static("/dist", "dist")
	// Little bit of middlewares for housekeeping
	e.Use(middleware.Logger())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))

	client, _, err := db.GetClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Page routes
	pages := e.Group("")
	pages.Use(middlewares.DatabaseMiddleware(client))
	pages.Use(middlewares.StateMiddleware)
	pages.GET("/", handlers.HandleIndex)
	pages.GET("/admin", handlers.AdminHandler)

	// Authentication routes
	e.GET("/login", handlers.HandleLogin)
	e.GET("/oauth2/callback", handlers.HandleOAuth2Callback, middlewares.DatabaseMiddleware(client))

	// API routes
	apis := e.Group("/api")
	apis.Use(middlewares.DatabaseMiddleware(client))
	apis.Use(middlewares.StateMiddleware)

	recipes := apis.Group("/recipes")
	recipes.GET("/replace/:id", handlers.ReplaceRecipe)
	recipes.GET("/random", handlers.GetRandomRecipes)
	recipes.GET("/add", handlers.AddRecipeToDatabase)
	recipes.GET("/all", handlers.GetAllRecipes)
	recipes.GET("/name", handlers.SearchRecipesByName)
	recipes.GET("/ingredients", handlers.SearchRecipesByIngredient)
	recipes.POST("/file", handlers.ProcessRecipeFile)
	recipes.GET("/filter", handlers.ChangeFilter)
	recipes.GET("/delete/:id", handlers.DeleteRecipe)
	recipes.GET("/delete/all", handlers.DeleteAllRecipes)
	recipes.GET("/favorites", handlers.Favorites)
	recipes.GET("/favorites/add/:id", handlers.AddFavorite)
	recipes.GET("/favorites/delete/:id", handlers.DeleteFavorite)

	apis.GET("/groceries", handlers.CombineIngredients)

	ingredient := apis.Group("/ingredient")
	ingredient.GET("/add/:id", handlers.AddIngredient)
	ingredient.GET("/delete/:id", handlers.DeleteIngredient)

	step := apis.Group("/step")
	step.GET("/add/:id", handlers.AddStep)
	step.GET("/delete/:id", handlers.DeleteStep)

	e.Logger.Info("Starting server at localhost:8080...")
	e.Logger.Fatal(e.Start(":8080"))
}
