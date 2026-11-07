package services

import "testing"

func TestToggleGroceries(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a"}, {Id: "b"}}}

	if r, ok := s.ToggleGroceries("b"); !ok || !r.SkipGroceries {
		t.Fatalf("ToggleGroceries(b) = %+v, %v, want skipped", r, ok)
	}
	if s.Recipes[0].SkipGroceries {
		t.Error("toggling b changed a")
	}
	if r, _ := s.ToggleGroceries("b"); r.SkipGroceries {
		t.Error("second toggle should include b again")
	}
	if _, ok := s.ToggleGroceries("missing"); ok {
		t.Error("ToggleGroceries(missing) = true, want false")
	}
}

func TestIncludeInGroceries(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", SkipGroceries: true}, {Id: "b", SkipGroceries: true}}}

	s.IncludeInGroceries("a")

	if s.Recipes[0].SkipGroceries {
		t.Error("a should be back on the grocery list")
	}
	if !s.Recipes[1].SkipGroceries {
		t.Error("b should be unchanged")
	}
}

func TestUnlisted(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a"}}}

	got := s.Unlisted([]Recipe{{Id: "a"}, {Id: "b"}, {Id: "b"}, {Id: "c"}})

	if len(got) != 2 || got[0].Id != "b" || got[1].Id != "c" {
		t.Errorf("Unlisted = %+v, want b and c", got)
	}
}

func TestDedupeRecipesKeepsFirst(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", SkipGroceries: true}, {Id: "b"}, {Id: "a"}}}

	s.DedupeRecipes()

	if len(s.Recipes) != 2 || s.Recipes[0].Id != "a" || !s.Recipes[0].SkipGroceries || s.Recipes[1].Id != "b" {
		t.Errorf("DedupeRecipes = %+v, want the first a, then b", s.Recipes)
	}
}

func TestReplaceRecipeKeepsDay(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", Day: "2026-11-11"}}}

	s.ReplaceRecipe("a", Recipe{Id: "b", Day: "2026-11-13"})

	if s.Recipes[0].Id != "b" || s.Recipes[0].Day != "2026-11-11" {
		t.Errorf("after replace got %+v, want b on Nov 11", s.Recipes[0])
	}
}

func TestAddRecipesStartUnplanned(t *testing.T) {
	var s State
	s.AddRecipes([]Recipe{{Id: "a", Day: "2026-11-14"}})

	if s.Recipes[0].Day != "" {
		t.Errorf("added recipe has Day %q, want unplanned", s.Recipes[0].Day)
	}
}

func ids(recipes []Recipe) []string {
	out := make([]string, len(recipes))
	for i, r := range recipes {
		out[i] = r.Id
	}
	return out
}
