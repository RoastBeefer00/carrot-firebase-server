package views

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
)

// ShellData is what the header, bottom navigation, and toast need on every page.
type ShellData struct {
	User         services.User
	Active       string // "plan", "recipes", "groceries", or "" for other pages
	GroceryCount int
	IsAdmin      bool
	// Toast is a short message about what the last action did, if anything.
	Toast string
}

// PlanData is what the plan page shows.
type PlanData struct {
	Recipes  []services.Recipe
	ShowWife bool
}

// RecipesData is what the recipes page shows.
type RecipesData struct {
	List      string // "week" or "fav"
	Week      []services.Recipe
	Favorites []services.Recipe
	Selected  *services.Recipe
	// Detail shows the selected recipe instead of the list on a phone.
	Detail bool
}

// Shown returns the recipes in the list that's currently shown.
func (d RecipesData) Shown() []services.Recipe {
	if d.List == "fav" {
		return d.Favorites
	}
	return d.Week
}

func (d RecipesData) selectedID() string {
	if d.Selected == nil || !d.Detail {
		return ""
	}
	return d.Selected.Id
}

// RecipesURL is the address of the recipes page for a list and selected recipe.
func RecipesURL(list, id string) string {
	q := url.Values{}
	if id != "" {
		q.Set("id", id)
	}
	if list == "fav" {
		q.Set("list", "fav")
	}
	if len(q) == 0 {
		return "/recipes"
	}
	return "/recipes?" + q.Encode()
}

func recipesOn(recipes []services.Recipe, day string) []services.Recipe {
	var out []services.Recipe
	for _, r := range recipes {
		if r.Day == day {
			out = append(out, r)
		}
	}
	return out
}

func planSummary(recipes []services.Recipe) string {
	if len(recipes) == 0 {
		return "No meals yet"
	}
	planned := len(recipes) - len(recipesOn(recipes, ""))
	return fmt.Sprintf("%s · %d planned", plural(len(recipes), "meal", "meals"), planned)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

func dayLabel(r services.Recipe) string {
	if r.Day == "" {
		return "This week · not planned"
	}
	return services.DayName(r.Day)
}

func vals(pairs ...string) string {
	m := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return toJSON(m)
}

// dayGroups are the days in week order followed by "" for unplanned meals.
var dayGroups = append(slices.Clone(services.Days), "")
