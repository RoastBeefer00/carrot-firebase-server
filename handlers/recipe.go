package handlers

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/RoastBeefer00/carrot-firebase-server/database"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
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

func SearchRecipesByName(c echo.Context) error {
	filter := c.FormValue("search")
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

    for _, recipe := range filteredRecipes {
        services.AllRecipes.AddRecipe(recipe)
    }

    return Render(c, http.StatusOK, views.Recipes(filteredRecipes))
}

func SearchRecipesByIngredient(c echo.Context) error {
	filter := c.FormValue("search")
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

    for _, recipe := range filteredRecipes {
        services.AllRecipes.AddRecipe(recipe)
    }

    return Render(c, http.StatusOK, views.Recipes(filteredRecipes))
}

func GetAllRecipes(c echo.Context) error {
	recipes, err := getAll()
    if err != nil {
        return err
    }

    for _, recipe := range recipes {
        services.AllRecipes.AddRecipe(recipe)
    }
    return Render(c, http.StatusOK, views.Recipes(recipes))
}

func GetRandomRecipe(c echo.Context) error {
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
    services.AllRecipes.AddRecipe(recipe)
	return Render(c, http.StatusOK, views.Recipe(recipe, recipe.Id))
}

func GetRandomRecipes(c echo.Context) error {
	var randomRecipes []services.Recipe
	amount := c.Param("amount")
	fmt.Println(amount)
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

    for _, recipe := range randomRecipes {
        services.AllRecipes.AddRecipe(recipe)
    }
    return Render(c, http.StatusOK, views.Recipes(randomRecipes))
}

func DeleteRecipe(c echo.Context) error {
    param := c.Param("id")
    id, err := strconv.Atoi(param)
    if err != nil {
        return err
    }

    for i, recipe := range services.AllRecipes.Recipes {
        if recipe.Id == id {
            services.AllRecipes.Recipes = append(services.AllRecipes.Recipes[:i], services.AllRecipes.Recipes[i+1:]...)
            break
        }
    }
    return c.NoContent(200)
}

func DeleteAllRecipes(c echo.Context) error {
    services.AllRecipes = services.Recipes{}
    return c.NoContent(200)
}

func ChangeFilter(c echo.Context) error {
    filter := c.QueryParam("filter")
    fmt.Println(filter)
    services.SetFilter(filter)
    return Render(c, http.StatusOK, views.Search(filter))
}
