package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
	"github.com/RoastBeefer00/carrot-firebase-server/middlewares"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
)

//go:generate templ generate
//go:generate tailwindcss -i ./dist/main.css -o ./dist/tailwind.css

type writeFunc func([]byte) (int, error)

func (f writeFunc) Write(p []byte) (int, error) { return f(p) }

func main() {
	// --- Load configuration from .env ---
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, falling back to environment variables")
	}

	dev := os.Getenv("APP_ENV") == "development"
	if dev {
		charmlog.SetReportTimestamp(true)
		charmlog.SetTimeFormat(time.Kitchen)
		charmlog.SetLevel(charmlog.DebugLevel)
		bridge := writeFunc(func(p []byte) (int, error) {
			charmlog.Info(strings.TrimRight(string(p), "\n"))
			return len(p), nil
		})
		log.SetFlags(0)
		log.SetOutput(bridge)
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
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Fatal("Missing required environment variable: ANTHROPIC_API_KEY")
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
	e.Pre(middleware.RemoveTrailingSlash())
	if dev {
		e.Logger.SetOutput(writeFunc(func(p []byte) (int, error) {
			charmlog.Info(strings.TrimRight(string(p), "\n"))
			return len(p), nil
		}))
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:  true,
			LogURI:     true,
			LogMethod:  true,
			LogLatency: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				charmlog.Info("req",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency.Round(time.Millisecond),
				)
				return nil
			},
		}))
	} else {
		e.Use(middleware.Logger())
	}
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

	// --- Start recipe cache listener ---
	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	defer cacheCancel()
	handlers.Cache = services.NewRecipeCache()
	handlers.Cache.Start(cacheCtx, client)
	if !handlers.Cache.WaitReady(10 * time.Second) {
		log.Println("Warning: recipe cache not ready after 10s; serving with empty cache")
	}

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
	recipes.POST("/replace/:id", handlers.ReplaceRecipe)
	recipes.GET("/random", handlers.GetRandomRecipes)
	recipes.GET("/add", handlers.AddRecipeToDatabase)
	recipes.GET("/all", handlers.GetAllRecipes)
	recipes.GET("/search", handlers.SearchRecipes)
	recipes.GET("/name", handlers.SearchRecipesByName)
	recipes.GET("/ingredients", handlers.SearchRecipesByIngredient)
	recipes.POST("/file", handlers.ProcessRecipeFile)
recipes.DELETE("/all", handlers.DeleteAllRecipes)
	recipes.DELETE("/:id", handlers.DeleteRecipe)
	recipes.GET("/favorites", handlers.Favorites)
	recipes.POST("/favorites/:id", handlers.ToggleFavorite)
	recipes.GET("/typeahead", handlers.TypeaheadRecipes)
	recipes.GET("/pick/:id", handlers.PickRecipe)

	apis.GET("/groceries", handlers.CombineIngredients)

e.Logger.Info("Starting server at localhost:8080...")
	e.Logger.Fatal(e.Start(":8080"))
}
