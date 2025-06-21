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
	handlers.EncryptionKey, err = base64.URLEncoding.DecodeString(handlers.EncryptionKeyBase64)
	if err != nil {
		log.Fatalf("Failed to decode encryption key: %v", err)
	}
	if len(handlers.EncryptionKey) != 32 { // AES-256 requires a 32-byte key
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

	// This will initiate our template renderer
	e.GET("/", handlers.HandleIndex)
	e.GET("/login", handlers.HandleLogin)
	e.GET("/oauth2/callback", handlers.HandleOAuth2Callback)
	e.GET("/admin", handlers.AdminHandler)
	e.GET("/refresh", handlers.GetRecipes)
	e.GET("/recipes/replace/:id", handlers.ReplaceRecipe)
	e.GET("/recipes/random", handlers.GetRandomRecipes)
	e.GET("/recipes/add", handlers.AddRecipeToDatabase)
	e.GET("/recipes/all", handlers.GetAllRecipes)
	e.GET("/recipes/name", handlers.SearchRecipesByName)
	e.GET("/recipes/ingredients", handlers.SearchRecipesByIngredient)
	e.POST("/recipes/file", handlers.ProcessRecipeFile)
	e.GET("/recipes/filter", handlers.ChangeFilter)
	e.GET("/recipes/delete/:id", handlers.DeleteRecipe)
	e.GET("/recipes/delete/all", handlers.DeleteAllRecipes)
	e.GET("/recipes/favorites", handlers.Favorites)
	e.GET("/recipes/favorites/add/:id", handlers.AddFavorite)
	e.GET("/recipes/favorites/delete/:id", handlers.DeleteFavorite)
	e.GET("/groceries", handlers.CombineIngredients)
	e.GET("/ingredient/add/:id", handlers.AddIngredient)
	e.GET("/ingredient/delete/:id", handlers.DeleteIngredient)
	e.GET("/step/add/:id", handlers.AddStep)
	e.GET("/step/delete/:id", handlers.DeleteStep)

	e.Logger.Info("Starting server at localhost:8080...")
	e.Logger.Fatal(e.Start(":8080"))
}
