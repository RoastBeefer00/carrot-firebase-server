package services

import "slices"

var Filter = "name"
var AllIngredients = []Ingredient{}

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

func (s *State) AddRecipe(recipe Recipe) {
    s.Recipes = append(s.Recipes, recipe)
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
    s.Recipes = append(s.Recipes, recipes...)
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
	Favorite    bool
}

func (r *Recipe) AddId(id string) {
    r.Id = id
}

type Ingredient struct {
	Quantity    string
	Measurement string
	Item        string
}

func SetFilter(filter string) {
    Filter = filter
}
