package handlers

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
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

func SearchRecipes(c echo.Context) error {
	state := GetStateFromContext(c)

	q := c.FormValue("search")
	filter := c.FormValue("filter")
	log.Printf(
		"Searching for recipes with filter=%s q=%s for user %s",
		filter, q, state.User.DisplayName,
	)

	var filteredRecipes []services.Recipe
	if filter == "ingredients" {
		filteredRecipes = Cache.SearchByIngredient(q)
	} else {
		filteredRecipes = Cache.SearchByName(q)
	}
	if len(filteredRecipes) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}
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
	if len(filteredRecipes) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}
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
	if len(filteredRecipes) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}
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

func TypeaheadRecipes(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("search"))
	if q == "" {
		return Render(c, http.StatusOK, views.TypeaheadList(nil))
	}
	filter := c.QueryParam("filter")
	if filter == "" {
		filter = "name"
	}
	var matches []services.Recipe
	if filter == "ingredients" {
		matches = Cache.SearchByIngredient(q)
	} else {
		matches = Cache.SearchByName(q)
	}
	if len(matches) > 6 {
		matches = matches[:6]
	}
	return Render(c, http.StatusOK, views.TypeaheadList(matches))
}

func PickRecipe(c echo.Context) error {
	state := GetStateFromContext(c)
	id := c.Param("id")
	recipe, ok := Cache.GetByID(id)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	if state.IsFavorite(recipe.Id) {
		recipe.Favorite = true
	}
	if err := Render(c, http.StatusOK, views.PickResponse(recipe)); err != nil {
		return err
	}
	state.AddRecipes([]services.Recipe{recipe})
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

	recipe.Name = formParams.Get("name")
	recipe.Time = formParams.Get("time")

	for _, ing := range formParams["ingredient[]"] {
		if strings.TrimSpace(ing) == "" {
			continue
		}
		recipe.Ingredients = append(recipe.Ingredients, services.NormalizeIngredient(ing))
	}
	for _, step := range formParams["step[]"] {
		if strings.TrimSpace(step) == "" {
			continue
		}
		recipe.Steps = append(recipe.Steps, step)
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
