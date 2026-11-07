package services

import "slices"

var Admins = []string{
    "roastbeefer000@gmail.com",
    "rjudes123@gmail.com",
}

type State struct {
    User User
    Recipes []Recipe
    Favorites []string
}

type User struct {
    Email string
    Uid string
    DisplayName string
}

func (s *State) AddFavorite(id string) {
    if !slices.Contains(s.Favorites, id) {
        s.Favorites = append(s.Favorites, id)
    }

    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            s.Recipes[i].Favorite = true
            break
        }
    }
}

// AddRecipes lists the recipes that aren't listed yet. New recipes start without a
// day, even when they come from someone else's plan.
func (s *State) AddRecipes(recipes []Recipe) {
    existing := make(map[string]bool, len(s.Recipes))
    for _, r := range s.Recipes {
        existing[r.Id] = true
    }
    for _, r := range recipes {
        if !existing[r.Id] {
            r.Day = ""
            s.Recipes = append(s.Recipes, r)
            existing[r.Id] = true
        }
    }
}

// FindRecipe returns the listed recipe with the given ID.
func (s *State) FindRecipe(id string) (Recipe, bool) {
    for _, r := range s.Recipes {
        if r.Id == id {
            return r, true
        }
    }
    return Recipe{}, false
}

// SetDay plans a listed recipe for a date such as "2026-11-09", or unplans it
// when day is empty, and returns the updated recipe.
func (s *State) SetDay(id, day string) (Recipe, bool) {
    for i := range s.Recipes {
        if s.Recipes[i].Id == id {
            s.Recipes[i].Day = day
            return s.Recipes[i], true
        }
    }
    return Recipe{}, false
}

// SwapDays trades the days of two listed recipes and returns them updated.
func (s *State) SwapDays(a, b string) (Recipe, Recipe, bool) {
    ia := slices.IndexFunc(s.Recipes, func(r Recipe) bool { return r.Id == a })
    ib := slices.IndexFunc(s.Recipes, func(r Recipe) bool { return r.Id == b })
    if ia < 0 || ib < 0 || ia == ib {
        return Recipe{}, Recipe{}, false
    }
    s.Recipes[ia].Day, s.Recipes[ib].Day = s.Recipes[ib].Day, s.Recipes[ia].Day
    return s.Recipes[ia], s.Recipes[ib], true
}


// Unlisted returns the given recipes that aren't already listed, without repeats.
func (s *State) Unlisted(recipes []Recipe) []Recipe {
    unlisted := make([]Recipe, 0, len(recipes))
    seen := make(map[string]bool, len(recipes))
    for _, r := range recipes {
        if !seen[r.Id] && !s.HasRecipe(r.Id) {
            unlisted = append(unlisted, r)
            seen[r.Id] = true
        }
    }
    return unlisted
}

// DedupeRecipes removes repeated recipes, keeping the first of each. Swapping a
// recipe could previously list the same one twice, and those are still saved.
func (s *State) DedupeRecipes() {
    s.Recipes = (&State{}).Unlisted(s.Recipes)
}

// IncludeInGroceries puts a listed recipe back on the grocery list.
func (s *State) IncludeInGroceries(id string) {
    for i := range s.Recipes {
        if s.Recipes[i].Id == id {
            s.Recipes[i].SkipGroceries = false
            return
        }
    }
}

// ToggleGroceries flips whether a listed recipe counts toward the grocery list and
// returns the updated recipe.
func (s *State) ToggleGroceries(id string) (Recipe, bool) {
    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            s.Recipes[i].SkipGroceries = !recipe.SkipGroceries
            return s.Recipes[i], true
        }
    }
    return Recipe{}, false
}

func (s *State) HasRecipe(id string) bool {
    return slices.ContainsFunc(s.Recipes, func(r Recipe) bool { return r.Id == id })
}

func (s *State) DeleteRecipe(id string) {
    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            s.Recipes = append(s.Recipes[:i], s.Recipes[i+1:]...)
            break
        }
    }
}

func (s *State) DeleteFavorite(id string) {
    for i, favorite := range s.Favorites {
        if favorite == id {
            s.Favorites = append(s.Favorites[:i], s.Favorites[i+1:]...)
            break
        }
    }

    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            s.Recipes[i].Favorite = false
            break
        }
    }
}

// ReplaceRecipe swaps a listed recipe for another, which takes over its day.
func (s *State) ReplaceRecipe(id string, newRecipe Recipe) {
    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            newRecipe.Day = recipe.Day
            s.Recipes[i] = newRecipe
            break
        }
    }
}

func (s *State) IsFavorite(id string) bool {
    return slices.Contains(s.Favorites, id)
}

type Recipe struct {
	Name        string
	Time        string
	Ingredients []string
	Steps       []string
	Id          string
	Favorite    bool `firestore:"-"`
	// SkipGroceries leaves the recipe out of the grocery list. Unset means included,
	// which also covers recipes saved before this field existed.
	SkipGroceries bool `firestore:"SkipGroceries,omitempty"`
	// Day is the date the recipe is planned for, such as "2026-11-09", or empty
	// when it isn't planned yet. See days.go.
	Day string `firestore:"Day,omitempty"`
	// InWeek reports whether the recipe is in this week's plan, for favorites.
	InWeek bool `firestore:"-"`
}

