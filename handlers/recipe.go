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

// addAndRender saves the recipes that aren't already listed, then renders their cards.
func addAndRender(c echo.Context, state services.State, recipes []services.Recipe) error {
	unlisted := state.Unlisted(recipes)
	if len(unlisted) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}
	state.AddRecipes(unlisted)
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.Recipes(unlisted, false))
}

// unlistedRecipeIDs returns the IDs of recipes that aren't on the user's homepage.
func unlistedRecipeIDs(c echo.Context, state services.State) ([]string, error) {
	docs, err := GetDbClient(c).Collection("ids").Documents(c.Request().Context()).GetAll()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no ids document found")
	}
	var ids IDs
	if err := docs[0].DataTo(&ids); err != nil {
		return nil, err
	}
	if len(ids.IDs) == 0 {
		return nil, fmt.Errorf("ids document empty")
	}
	return slices.DeleteFunc(ids.IDs, state.HasRecipe), nil
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
	return addAndRender(c, state, filteredRecipes)
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
	return addAndRender(c, state, filteredRecipes)
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

	return addAndRender(c, state, filteredRecipes)
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
	if state.HasRecipe(recipe.Id) {
		// Picking a recipe that's already listed means it's needed again, so refresh
		// its card rather than adding a second one.
		state.IncludeInGroceries(recipe.Id)
		c.Response().Header().Set("HX-Retarget", "#recipe-"+recipe.Id)
		c.Response().Header().Set("HX-Reswap", "outerHTML")
	} else {
		state.AddRecipes([]services.Recipe{recipe})
	}
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.PickResponse(recipe))
}

func ToggleRecipeGroceries(c echo.Context) error {
	state := GetStateFromContext(c)
	recipe, ok := state.ToggleGroceries(c.Param("id"))
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.GroceryToggle(recipe))
}

func ToggleRecipeInWeek(c echo.Context) error {
	state := GetStateFromContext(c)
	recipe, ok := Cache.GetByID(c.Param("id"))
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	if state.HasRecipe(recipe.Id) {
		state.DeleteRecipe(recipe.Id)
	} else {
		state.AddRecipes([]services.Recipe{recipe})
		recipe.InWeek = true
	}
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.WeekToggle(recipe))
}

func ReplaceRecipe(c echo.Context) error {
	state := GetStateFromContext(c)
	client := GetDbClient(c)
	ctx := c.Request().Context()

	id := c.Param("id")

	ids, err := unlistedRecipeIDs(c, state)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}
	randomId := ids[rand.IntN(len(ids))]
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
	state.ReplaceRecipe(id, recipe)
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return Render(c, http.StatusOK, views.Recipe(recipe, false))
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

	shuffled, err := unlistedRecipeIDs(c, state)
	if err != nil {
		return err
	}
	if len(shuffled) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}

	if amountInt > len(shuffled) {
		amountInt = len(shuffled)
	}

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

	return addAndRender(c, state, randomRecipes)
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
	state.DeleteRecipe(id)
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return c.NoContent(200)
}

func DeleteAllRecipes(c echo.Context) error {
	state := GetStateFromContext(c)

	log.Printf(
		"Deleting all recipes for user %s with email %s",
		state.User.DisplayName,
		state.User.Email,
	)
	state.Recipes = make([]services.Recipe, 0)
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return c.NoContent(200)
}


func GetWifeRecipes(c echo.Context) error {
	state := GetStateFromContext(c)

	if state.User.Email != "roastbeefer000@gmail.com" {
		return c.NoContent(http.StatusForbidden)
	}

	client := GetDbClient(c)
	ctx := c.Request().Context()

	docs, err := client.Collection("users").
		Where("User.Email", "==", "rjudes123@gmail.com").
		Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return c.NoContent(http.StatusOK)
	}

	var wife services.State
	if err := docs[0].DataTo(&wife); err != nil {
		return err
	}

	for i, r := range wife.Recipes {
		wife.Recipes[i].Favorite = wife.IsFavorite(r.Id)
		wife.Recipes[i].SkipGroceries = false
	}

	log.Printf("Serving %d recipes for wife (%s) to user %s", len(wife.Recipes), wife.User.Email, state.User.Email)
	return addAndRender(c, state, wife.Recipes)
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
