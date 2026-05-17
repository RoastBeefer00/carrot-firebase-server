package handlers

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type IDs struct {
	IDs []string `json:"ids"`
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	ctx.Response().Writer.WriteHeader(statusCode)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func GetRecipes(c echo.Context) error {
	state := GetStateFromContext(c)

	log.Printf(
		"Refreshing %d recipes for user %s with email %s",
		len(state.Recipes),
		state.User.DisplayName,
		state.User.Email,
	)
	return Render(c, http.StatusOK, views.Recipes(state.Recipes, false))
}

func GetAllRecipes(c echo.Context) error {
	return c.JSON(http.StatusOK, Cache.All())
}

func SearchRecipesByName(c echo.Context) error {
	state := GetStateFromContext(c)

	filter := c.FormValue("search")
	log.Printf(
		"Searching for recipes with name %s for user %s with email %s",
		filter,
		state.User.DisplayName,
		state.User.Email,
	)

	filteredRecipes := Cache.SearchByName(filter)
	for i, recipe := range filteredRecipes {
		if state.IsFavorite(recipe.Id) {
			filteredRecipes[i].Favorite = true
		}
	}
	if err := Render(c, http.StatusOK, views.Recipes(filteredRecipes, false)); err != nil {
		return err
	}

	state.AddRecipes(filteredRecipes)
	return db.UpdateState(state, c)
}

func SearchRecipesByIngredient(c echo.Context) error {
	state := GetStateFromContext(c)

	filter := c.FormValue("search")
	log.Printf(
		"Searching for recipes with ingredient %s for user %s with email %s",
		filter,
		state.User.DisplayName,
		state.User.Email,
	)

	filteredRecipes := Cache.SearchByIngredient(filter)
	for i, recipe := range filteredRecipes {
		if state.IsFavorite(recipe.Id) {
			filteredRecipes[i].Favorite = true
		}
	}

	if err := Render(c, http.StatusOK, views.Recipes(filteredRecipes, false)); err != nil {
		return err
	}
	state.AddRecipes(filteredRecipes)
	return db.UpdateState(state, c)
}

func ReplaceRecipe(c echo.Context) error {
	state := GetStateFromContext(c)
	client := GetDbClient(c)
	ctx := c.Request().Context()

	id := c.Param("id")

	docs, err := client.Collection("ids").Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no ids document found")
	}
	var ids IDs
	if err := docs[0].DataTo(&ids); err != nil {
		return err
	}
	if len(ids.IDs) == 0 {
		return fmt.Errorf("ids document empty")
	}
	randomId := ids.IDs[rand.IntN(len(ids.IDs))]
	doc, err := client.Collection("recipes").Doc(randomId).Get(ctx)
	if err != nil {
		return err
	}
	var recipe services.Recipe
	if err := doc.DataTo(&recipe); err != nil {
		return err
	}
	if state.IsFavorite(recipe.Id) {
		recipe.Favorite = true
	}

	log.Printf(
		"Replacing recipe with id %s with recipe %s for user %s with email %s",
		id,
		recipe.Name,
		state.User.DisplayName,
		state.User.Email,
	)
	err = Render(c, http.StatusOK, views.Recipe(recipe, false))
	if err != nil {
		return err
	}

	state.ReplaceRecipe(id, recipe)
	return db.UpdateState(state, c)
}

func GetRandomRecipes(c echo.Context) error {
	ctx := c.Request().Context()
	state := GetStateFromContext(c)
	client := GetDbClient(c)

	amount := c.FormValue("amount")
	log.Printf(
		"Fetching %s random recipes for user %s with email %s",
		amount,
		state.User.DisplayName,
		state.User.Email,
	)

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return err
	}

	docs, err := client.Collection("ids").Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no ids document found")
	}
	var ids IDs
	if err := docs[0].DataTo(&ids); err != nil {
		return err
	}
	if len(ids.IDs) == 0 {
		return fmt.Errorf("ids document empty")
	}

	if amountInt > len(ids.IDs) {
		amountInt = len(ids.IDs)
	}

	shuffled := make([]string, len(ids.IDs))
	copy(shuffled, ids.IDs)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	pick := shuffled[:amountInt]

	randomRecipes := make([]services.Recipe, amountInt)
	errs := make([]error, amountInt)
	var wg sync.WaitGroup
	for i, id := range pick {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			doc, err := client.Collection("recipes").Doc(id).Get(ctx)
			if err != nil {
				errs[i] = err
				return
			}
			if err := doc.DataTo(&randomRecipes[i]); err != nil {
				errs[i] = err
			}
		}(i, id)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	for i, recipe := range randomRecipes {
		if state.IsFavorite(recipe.Id) {
			randomRecipes[i].Favorite = true
		}
	}

	err = Render(c, http.StatusOK, views.Recipes(randomRecipes, false))
	if err != nil {
		return err
	}
	state.AddRecipes(randomRecipes)
	return db.UpdateState(state, c)
}

func DeleteRecipe(c echo.Context) error {
	state := GetStateFromContext(c)

	id := c.Param("id")

	log.Printf(
		"Deleting recipe with id %s for user %s with email %s",
		id,
		state.User.DisplayName,
		state.User.Email,
	)
	err := c.NoContent(200)
	if err != nil {
		return err
	}
	state.DeleteRecipe(id)
	return db.UpdateState(state, c)
}

func DeleteAllRecipes(c echo.Context) error {
	state := GetStateFromContext(c)

	log.Printf(
		"Deleting all recipes for user %s with email %s",
		state.User.DisplayName,
		state.User.Email,
	)
	err := c.NoContent(200)
	if err != nil {
		return err
	}
	state.Recipes = make([]services.Recipe, 0)
	return db.UpdateState(state, c)
}

func ChangeFilter(c echo.Context) error {
	state := GetStateFromContext(c)
	filter := c.QueryParam("filter")
	state.Filter = filter
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.Search(filter))
}

func AddRecipeToDatabase(c echo.Context) error {
	recipe := services.Recipe{}
	state := GetStateFromContext(c)

	if !slices.Contains(services.Admins, state.User.Email) {
		return c.NoContent(200)
	}

	formParams, err := c.FormParams()
	if err != nil {
		return err
	}

	ingredients := make(map[string]string)
	steps := make(map[string]string)
	for key, value := range formParams {
		if strings.HasPrefix(key, "name") {
			recipe.Name = value[0]
		} else if strings.HasPrefix(key, "time") {
			recipe.Time = value[0]
		} else if strings.HasPrefix(key, "ingredient") {
			ingredients[key] = value[0]
		} else if strings.HasPrefix(key, "step") {
			steps[key] = value[0]
		}
	}

	ingKeys := make([]string, 0, len(ingredients))
	for k := range ingredients {
		ingKeys = append(ingKeys, k)
	}
	sort.Strings(ingKeys)

	for _, k := range ingKeys {
		recipe.Ingredients = append(recipe.Ingredients, services.NormalizeIngredient(ingredients[k]))
	}

	stepKeys := make([]string, 0, len(steps))
	for k := range steps {
		stepKeys = append(stepKeys, k)
	}
	sort.Strings(stepKeys)

	for _, k := range stepKeys {
		recipe.Steps = append(recipe.Steps, steps[k])
	}

	log.Printf(
		"User %s with email %s is adding recipe %v",
		state.User.DisplayName,
		state.User.Email,
		recipe,
	)
	client, ctx, err := db.GetClient()
	if err != nil {
		return err
	}
	defer client.Close()

	doc, _, err := client.Collection("recipes").Add(ctx, recipe)
	if err != nil {
		return err
	}
	log.Println(doc.ID)

	recipe.Id = doc.ID
	_, err = client.Collection("recipes").Doc(doc.ID).Set(ctx, recipe)
	if err != nil {
		return err
	}

	idDoc, err := client.Collection("ids").Doc("hHjqXrWMhH7WDTPEwlkN").Get(ctx)
	if err != nil {
		return err
	}

	var ids IDs
	err = idDoc.DataTo(&ids)
	if err != nil {
		return err
	}

	ids.IDs = append(ids.IDs, doc.ID)

	_, err = client.Collection("ids").Doc("hHjqXrWMhH7WDTPEwlkN").Set(ctx, ids)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, views.Admin())
}
