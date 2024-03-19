package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/RoastBeefer00/carrot-firebase-server/database"
)

type Recipe struct {
    Name string `json:"name"`
    Time string `json:"time"`
    Ingredients []string `json:"ingredients"`
    Steps []string `json:"steps"`
}

type IDs struct {
    IDs []string `json:"ids"`
}

func getAll() []Recipe {
    client, ctx, err := database.GetClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    docs, err := client.Collection("recipes").Documents(ctx).GetAll()
    if err != nil {
        log.Fatal(err)
    }

    var recipes []Recipe
    for _, doc := range docs {
        var recipe Recipe
        err = doc.DataTo(&recipe)
        if err != nil {
            log.Fatal(err)
        }

        recipes = append(recipes, recipe)
    }

    return recipes
}

func filterRecipes(recipes []Recipe, function func(Recipe) bool) []Recipe {
    var filteredRecipes []Recipe

    for _, recipe := range recipes {
        if function(recipe) {
            filteredRecipes = append(filteredRecipes, recipe)
        }
    }

    return filteredRecipes
}

func SearchRecipesByName(w http.ResponseWriter, r *http.Request) {
    name := r.PathValue("name")
    recipes := getAll()
    var filteredRecipes []Recipe

    filterFunc := func(recipe Recipe) bool {
        if strings.Contains(strings.ToLower(recipe.Name), strings.ToLower(name)) {
            return true
        }
        return false
    }

    filteredRecipes = filterRecipes(recipes, filterFunc)

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(filteredRecipes)
    if err != nil {
        log.Fatal(err)
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func SearchRecipesByIngredient(w http.ResponseWriter, r *http.Request) {
    filter := r.PathValue("ingredient")
    recipes := getAll()
    var filteredRecipes []Recipe

    filterFunc := func(recipe Recipe) bool {
        for _, ingredient := range recipe.Ingredients {
            if strings.Contains(strings.ToLower(ingredient), strings.ToLower(filter)) {
                return true
            }
        }
        return false
    }

    filteredRecipes = filterRecipes(recipes, filterFunc)

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(filteredRecipes)
    if err != nil {
        log.Fatal(err)
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func GetAllRecipes(w http.ResponseWriter, r *http.Request) {
    recipes := getAll()

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(recipes)
    if err != nil {
        log.Fatal(err)
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func GetRandomRecipe(w http.ResponseWriter, r *http.Request) {
    client, ctx, err := database.GetClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    docs, err := client.Collection("ids").Documents(ctx).GetAll()
    if err != nil {
        log.Fatalln(err)
    }
    var ids IDs
    docs[0].DataTo(&ids)
    randomId := ids.IDs[rand.IntN(len(ids.IDs))]
    doc, err := client.Collection("recipes").Doc(randomId).Get(ctx)
    if err != nil {
        log.Fatalln(err)
    }
    var recipe Recipe
    doc.DataTo(&recipe)

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(recipe)
    if err != nil {
        log.Fatal(err)
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func GetRandomRecipes(w http.ResponseWriter, r *http.Request) {
    var randomRecipes []Recipe
    amount := r.PathValue("amount")
    fmt.Println(amount)
    amountInt, err := strconv.Atoi(amount)
    if err != nil {
        log.Fatal(err)
    }

    client, ctx, err := database.GetClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    docs, err := client.Collection("ids").Documents(ctx).GetAll()
    if err != nil {
        log.Fatalln(err)
    }
    var ids IDs
    docs[0].DataTo(&ids)

    var wg sync.WaitGroup
    for _ = range amountInt {
        wg.Add(1)
        go func() {
            defer wg.Done()
            randomId := ids.IDs[rand.IntN(len(ids.IDs))]
            doc, err := client.Collection("recipes").Doc(randomId).Get(ctx)
            if err != nil {
                log.Fatalln(err)
            }
            var recipe Recipe
            doc.DataTo(&recipe)

            randomRecipes = append(randomRecipes, recipe)
        }()
    }
    wg.Wait()

    w.Header().Set("Content-Type", "application/json")
    data, err := json.Marshal(randomRecipes)
    if err != nil {
        log.Fatal(err)
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
