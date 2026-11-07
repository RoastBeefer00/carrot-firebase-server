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

func (s *State) AddRecipes(recipes []Recipe) {
    existing := make(map[string]bool, len(s.Recipes))
    for _, r := range s.Recipes {
        existing[r.Id] = true
    }
    for _, r := range recipes {
        if !existing[r.Id] {
            s.Recipes = append(s.Recipes, r)
            existing[r.Id] = true
        }
    }
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

func (s *State) ReplaceRecipe(id string, newRecipe Recipe) {
    for i, recipe := range s.Recipes {
        if recipe.Id == id {
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
	// InWeek reports whether the recipe is on the homepage, for the favorites page.
	InWeek bool `firestore:"-"`
}

