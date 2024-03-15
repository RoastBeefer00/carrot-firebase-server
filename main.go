package main

import (
	"fmt"
	"net/http"

	"github.com/RoastBeefer00/carrot-firebase-server/handlers"
)

func main() {
    mux := http.NewServeMux()


    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {fmt.Fprintf(w, "Hello World")})
    mux.HandleFunc("GET /recipes", handlers.GetAllRecipes)
    mux.HandleFunc("GET /recipes/random", handlers.GetRandomRecipe)
    mux.HandleFunc("GET /recipes/random/{amount}", handlers.GetRandomRecipes)
    //
    // mux.HandleFunc("POST /recipes", addRecipe)
    // mux.HandleFunc("DELETE /recipes", deleteRecipe)

    http.ListenAndServe(":8080", mux)
}
