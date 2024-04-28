package services

var Filter = "name"
var AllIngredients = []Ingredient{}

var Admins = []string{
    "roastbeefer000@gmail.com",
    "rjudes123@gmail.com",
}

type State struct {
    User User
    Recipes []Recipe
}

type User struct {
    Email string
    Uid string
    DisplayName string
}

func (s *State) AddRecipe(recipe Recipe) {
    s.Recipes = append(s.Recipes, recipe)
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

func (s *State) ReplaceRecipe(id string, newRecipe Recipe) {
    for i, recipe := range s.Recipes {
        if recipe.Id == id {
            s.Recipes[i] = newRecipe
            break
        }
    }
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
