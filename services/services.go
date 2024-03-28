package services

var Id = 0
var Filter = "name"

type Recipes struct {
	recipes []Recipe
}

func (r *Recipes) AddRecipe(recipe Recipe) {
    r.recipes = append(r.recipes, recipe)
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

func SetFilter(filter string) {
    Filter = filter
}
