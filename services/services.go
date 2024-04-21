package services

var Id = 0
var Filter = "name"
var AllRecipes = Recipes{}
var AllIngredients = []Ingredient{}

type State struct {
    User User
    Recipes Recipes
}

type User struct {
    Email string
    Uid string
    DisplayName string
}

type Recipes struct {
	Recipes []Recipe
}

func (r *Recipes) AddRecipe(recipe Recipe) {
    r.Recipes = append(r.Recipes, recipe)
}

func (r *Recipes) DeleteRecipe(id int) {
    for i, recipe := range r.Recipes {
        if recipe.Id == id {
            r.Recipes = append(r.Recipes[:i], r.Recipes[i+1:]...)
            break
        }
    }
}

func (r *Recipes) ReplaceRecipe(id int, newRecipe Recipe) {
    for i, recipe := range r.Recipes {
        if recipe.Id == id {
            r.Recipes[i] = newRecipe
            break
        }
    }
}

type Recipe struct {
	Name        string
	Time        string
	Ingredients []string
	Steps       []string
	Id          int
}

func (r *Recipe) AddId() {
    Id++
    r.Id = Id
}

type Ingredient struct {
	Quantity    string
	Measurement string
	Item        string
}

func SetFilter(filter string) {
    Filter = filter
}
