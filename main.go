package main

import (
	"fmt"
    "flag"
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
    "github.com/RoastBeefer00/carrot-firebase-server/frontend"
)

func cors(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        handler.ServeHTTP(w, r)
    })
}

func main() {
    devMode := false
    flag.BoolVar(&devMode, "dev", devMode, "Enable dev mode")
    flag.Parse()

    mux := http.NewServeMux()


    // mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {fmt.Fprintf(w, "Hello World")})
    mux.Handle("/", frontend.SvelteKitHandler("/"))
    mux.HandleFunc("GET /recipes", handlers.GetAllRecipes)
    mux.HandleFunc("GET /recipes/random", handlers.GetRandomRecipe)
    mux.HandleFunc("GET /recipes/random/{amount}", handlers.GetRandomRecipes)
    mux.HandleFunc("GET /recipes/name/{name}", handlers.SearchRecipesByName)
    mux.HandleFunc("GET /recipes/ingredient/{ingredient}", handlers.SearchRecipesByIngredient)
    mux.HandleFunc("POST /groceries", handlers.CombineIngredients)
    //
    // mux.HandleFunc("POST /recipes", addRecipe)
    // mux.HandleFunc("DELETE /recipes", deleteRecipe)


    var handler http.Handler = mux

    if devMode {
        handler = cors(handler)
        fmt.Println("Serving in dev mode")
    }

    http.ListenAndServe(":8080", handler)
}
