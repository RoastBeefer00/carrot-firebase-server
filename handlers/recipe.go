package handlers

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/RoastBeefer00/carrot-firebase-server/database"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type IDs struct {
	IDs []string `json:"ids"`
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func getAll() ([]services.Recipe, error) {
	client, ctx, err := database.GetClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	docs, err := client.Collection("recipes").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var recipes []services.Recipe
	for _, doc := range docs {
		var recipe services.Recipe
		err = doc.DataTo(&recipe)
		if err != nil {
			return nil, err
		}

		recipe.AddId()
		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func filterRecipes(recipes []services.Recipe, function func(services.Recipe) bool) []services.Recipe {
	var filteredRecipes []services.Recipe

	for _, recipe := range recipes {
		if function(recipe) {
			recipe.AddId()
			filteredRecipes = append(filteredRecipes, recipe)
		}
	}

	return filteredRecipes
}

func GetRecipes(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

    log.Printf("Refreshing %d recipes for user %s with email %s", len(state.Recipes), state.User.DisplayName, state.User.Email)
    return Render(c, http.StatusOK, views.Recipes(state.Recipes, false))
}

func SearchRecipesByName(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

	filter := c.FormValue("search")
    log.Printf("Searching for recipes with name %s for user %s with email %s", filter, state.User.DisplayName, state.User.Email)

	recipes, err := getAll()
	if err != nil {
		return err
	}
	var filteredRecipes []services.Recipe

	filterFunc := func(recipe services.Recipe) bool {
		if strings.Contains(strings.ToLower(recipe.Name), strings.ToLower(filter)) {
			return true
		}
		return false
	}

	filteredRecipes = filterRecipes(recipes, filterFunc)
    err = Render(c, http.StatusOK, views.Recipes(filteredRecipes, false))
    if err != nil {
        return err
    }

    state.Recipes = append(state.Recipes, filteredRecipes...)
    return database.UpdateState(state)
}

func SearchRecipesByIngredient(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

	filter := c.FormValue("search")
    log.Printf("Searching for recipes with ingredient %s for user %s with email %s", filter, state.User.DisplayName, state.User.Email)

	recipes, err := getAll()
	if err != nil {
		return err
	}
	var filteredRecipes []services.Recipe

	filterFunc := func(recipe services.Recipe) bool {
		for _, ingredient := range recipe.Ingredients {
			if strings.Contains(strings.ToLower(ingredient), strings.ToLower(filter)) {
				return true
			}
		}
		return false
	}

	filteredRecipes = filterRecipes(recipes, filterFunc)

    err = Render(c, http.StatusOK, views.Recipes(filteredRecipes, false))
    if err != nil {
        return err
    }
    state.AddRecipes(filteredRecipes)
    return database.UpdateState(state)
}

func ReplaceRecipe(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

	param := c.Param("id")
    fmt.Println(param)

	id, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

	client, ctx, err := database.GetClient()
	if err != nil {
		return err
	}
	defer client.Close()

	docs, err := client.Collection("ids").Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	var ids IDs
	docs[0].DataTo(&ids)
	randomId := ids.IDs[rand.IntN(len(ids.IDs))]
	doc, err := client.Collection("recipes").Doc(randomId).Get(ctx)
	if err != nil {
		return err
	}
	var recipe services.Recipe
	doc.DataTo(&recipe)

	recipe.AddId()

    log.Printf("Replacing recipe with id %d with recipe %s for user %s with email %s", id, recipe.Name, state.User.DisplayName, state.User.Email)
    err = Render(c, http.StatusOK, views.Recipe(recipe, recipe.Id, false))
    if err != nil {
        return err
    }

    state.ReplaceRecipe(id, recipe)
    return database.UpdateState(state)
}

func GetRandomRecipes(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

	var randomRecipes []services.Recipe
	amount := c.FormValue("amount")
    log.Printf("Fetching %s random recipes for user %s with email %s", amount, state.User.DisplayName, state.User.Email)

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return err
	}

	client, ctx, err := database.GetClient()
	if err != nil {
		return err
	}
	defer client.Close()

	docs, err := client.Collection("ids").Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	var ids IDs
	docs[0].DataTo(&ids)

	var wg sync.WaitGroup
	for range amountInt {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			randomId := ids.IDs[rand.IntN(len(ids.IDs))]
			doc, err := client.Collection("recipes").Doc(randomId).Get(ctx)
			if err != nil {
				return err
			}
			var recipe services.Recipe
			doc.DataTo(&recipe)

			recipe.AddId()
			randomRecipes = append(randomRecipes, recipe)
			return nil
		}()
	}
	wg.Wait()

    err = Render(c, http.StatusOK, views.Recipes(randomRecipes, false))
    if err != nil {
        return err
    }
    state.AddRecipes(randomRecipes)
    return database.UpdateState(state)
}

func DeleteRecipe(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

	param := c.Param("id")
    fmt.Println(param)

	id, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

    log.Printf("Deleting recipe with id %d for user %s with email %s", id, state.User.DisplayName, state.User.Email)
    err = c.NoContent(200)
    if err != nil {
        return err
    }
    state.DeleteRecipe(id)
    return database.UpdateState(state)
}

func DeleteAllRecipes(c echo.Context) error {
    state, err := database.GetState(c)
    if err != nil {
        return err
    }

    log.Printf("Deleting all recipes for user %s with email %s", state.User.DisplayName, state.User.Email)
    err = c.NoContent(200)
    if err != nil {
        return err
    }
    state.Recipes = make([]services.Recipe, 0)
    return database.UpdateState(state)
}

func ChangeFilter(c echo.Context) error {
	filter := c.QueryParam("filter")
	fmt.Println(filter)
	services.SetFilter(filter)
	return Render(c, http.StatusOK, views.Search(filter))
}
