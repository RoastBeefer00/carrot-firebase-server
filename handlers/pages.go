package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// husbandEmail is the user who gets the button that copies their wife's plan.
const husbandEmail = "roastbeefer000@gmail.com"

// viewCtx says which page is showing. Every request from inside #page sends it
// (see views.viewCtx), so an action can answer with that page re-rendered.
type viewCtx struct {
	View string // "plan", "recipes", or "groceries"
	Sel  string // the selected recipe on the recipes page
	List string // "week" or "fav" on the recipes page
	Week string // the Monday of the week shown, or empty for the current week
}

func (v viewCtx) normalized() viewCtx {
	if v.View != "recipes" && v.View != "groceries" {
		v.View = "plan"
	}
	if v.List != "fav" {
		v.List = "week"
	}
	return v
}

func viewFrom(c echo.Context) viewCtx {
	return viewCtx{
		View: c.FormValue("view"),
		Sel:  c.FormValue("sel"),
		List: c.FormValue("list"),
		Week: c.FormValue("week"),
	}.normalized()
}

// today is the current date in the user's time zone (see StateMiddleware).
func today(c echo.Context) time.Time {
	if t, ok := c.Get("today").(time.Time); ok {
		return t
	}
	return services.Today("")
}

// week resolves the week a request is for, defaulting to the current one.
func week(c echo.Context, v viewCtx) views.Week {
	now := today(c)
	start := services.WeekStart(now)
	if t, ok := services.ParseDay(v.Week); ok {
		start = services.WeekStart(t)
	}
	return views.Week{Start: start, Today: now}
}

func HandleIndex(c echo.Context) error {
	return renderFull(c, viewCtx{View: "plan", Week: c.QueryParam("week")}.normalized())
}

func HandleRecipesPage(c echo.Context) error {
	return renderFull(c, viewCtx{View: "recipes", Sel: c.QueryParam("id"), List: c.QueryParam("list"), Week: c.QueryParam("week")}.normalized())
}

func HandleGroceriesPage(c echo.Context) error {
	return renderFull(c, viewCtx{View: "groceries"}.normalized())
}

// HandleView renders a page into #page, for moving around without a full load.
func HandleView(c echo.Context) error {
	return Render(c, http.StatusOK, pageContent(c, GetStateFromContext(c), viewFrom(c), ""))
}

func renderFull(c echo.Context, v viewCtx) error {
	return Render(c, http.StatusOK, views.Index(pageContent(c, GetStateFromContext(c), v, "")))
}

// respond answers an action with the current page re-rendered and a toast
// saying what happened. On the recipes page the address follows the selection.
func respond(c echo.Context, state services.State, v viewCtx, toast string) error {
	if v.View == "recipes" {
		c.Response().Header().Set("HX-Replace-Url", views.RecipesURL(v.List, v.Sel, week(c, v).Param()))
	}
	return Render(c, http.StatusOK, pageContent(c, state, v, toast))
}

// saveAndRespond saves the state, then responds as respond does.
func saveAndRespond(c echo.Context, state services.State, v viewCtx, toast string) error {
	if err := db.UpdateState(state, c); err != nil {
		return err
	}
	return respond(c, state, v, toast)
}

func shellData(state services.State, active, toast string) views.ShellData {
	return views.ShellData{
		User:         state.User,
		Active:       active,
		GroceryCount: len(groceryItems(state)),
		IsAdmin:      slices.Contains(services.Admins, state.User.Email),
		Toast:        toast,
	}
}

func pageContent(c echo.Context, state services.State, v viewCtx, toast string) templ.Component {
	// Favorite isn't saved with the recipe, and recipes just added don't have it.
	state.Recipes = slices.Clone(state.Recipes)
	for i, r := range state.Recipes {
		state.Recipes[i].Favorite = state.IsFavorite(r.Id)
	}
	shell := shellData(state, v.View, toast)
	switch v.View {
	case "recipes":
		return views.RecipesPage(shell, recipesData(state, v, week(c, v)))
	case "groceries":
		return views.GroceriesPage(shell, groceryItems(state), groceryMeals(state))
	default:
		return views.PlanPage(shell, views.PlanData{
			Recipes:  state.Recipes,
			ShowWife: state.User.Email == husbandEmail,
			Week:     week(c, v),
		})
	}
}

func recipesData(state services.State, v viewCtx, w views.Week) views.RecipesData {
	meals := services.ByDay(state.Recipes)
	for i := range meals {
		meals[i].InWeek = true
	}
	d := views.RecipesData{List: v.List, Meals: meals, Favorites: favoriteRecipes(state), Week: w}
	shown := d.Shown()
	for i := range shown {
		if shown[i].Id == v.Sel {
			d.Selected = &shown[i]
			d.Detail = true
		}
	}
	// With nothing selected, the desktop layout shows the first recipe beside
	// the list, while a phone shows only the list.
	if d.Selected == nil && len(shown) > 0 {
		d.Selected = &shown[0]
	}
	return d
}

// favoriteRecipes returns the user's favorites, using this week's copy of any
// that are planned so they carry its day and grocery setting.
func favoriteRecipes(state services.State) []services.Recipe {
	favorites := make([]services.Recipe, 0, len(state.Favorites))
	for _, id := range state.Favorites {
		r, inWeek := state.FindRecipe(id)
		if !inWeek {
			var ok bool
			if r, ok = Cache.GetByID(id); !ok {
				continue
			}
		}
		r.Favorite = true
		r.InWeek = inWeek
		favorites = append(favorites, r)
	}
	return favorites
}

func groceryItems(state services.State) []services.GroceryItem {
	var lines []string
	for _, r := range state.Recipes {
		if !r.SkipGroceries {
			lines = append(lines, r.Ingredients...)
		}
	}
	return services.CombineIngredients(lines)
}

func groceryMeals(state services.State) int {
	n := 0
	for _, r := range state.Recipes {
		if !r.SkipGroceries {
			n++
		}
	}
	return n
}

// SetRecipeDay plans a recipe for a day, or unplans it when day is empty.
func SetRecipeDay(c echo.Context) error {
	state := GetStateFromContext(c)
	day := c.FormValue("day")
	if !services.ValidDay(day) {
		return c.NoContent(http.StatusBadRequest)
	}
	r, ok := state.SetDay(c.Param("id"), day)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	toast := fmt.Sprintf("%s moved to Not planned", r.Name)
	if day != "" {
		toast = fmt.Sprintf("%s planned for %s", r.Name, views.DayLong(day))
	}
	return saveAndRespond(c, state, viewFrom(c), toast)
}

// SwapRecipeDays trades the days of two recipes.
func SwapRecipeDays(c echo.Context) error {
	state := GetStateFromContext(c)
	a, b, ok := state.SwapDays(c.Param("id"), c.FormValue("with"))
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	return saveAndRespond(c, state, viewFrom(c), fmt.Sprintf("Swapped %s and %s", a.Name, b.Name))
}
