package services

var Id = 0
var Filter = "name"
var AllRecipes = Recipes{}
var AllIngredients = []Ingredient{}

type User struct {
    Email string
    Uid string
    Token string
    DisplayName string
}

type Recipes struct {
	Recipes []Recipe
}

func (r *Recipes) AddRecipe(recipe Recipe) {
    r.Recipes = append(r.Recipes, recipe)
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
