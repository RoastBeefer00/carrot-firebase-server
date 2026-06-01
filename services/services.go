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
}

type Ingredient struct {
	Quantity    string
	Measurement string
	Item        string
}
