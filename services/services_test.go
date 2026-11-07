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

func TestSetDay(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a"}, {Id: "b", Day: "Tue"}}}

	if r, ok := s.SetDay("a", "Mon"); !ok || r.Day != "Mon" {
		t.Fatalf("SetDay(a, Mon) = %+v, %v, want Mon", r, ok)
	}
	if r, _ := s.SetDay("b", ""); r.Day != "" {
		t.Errorf("SetDay(b, \"\") left Day = %q, want unplanned", r.Day)
	}
	if _, ok := s.SetDay("missing", "Mon"); ok {
		t.Error("SetDay(missing) = true, want false")
	}
}

func TestSwapDays(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", Day: "Mon"}, {Id: "b"}}}

	a, b, ok := s.SwapDays("a", "b")
	if !ok || a.Day != "" || b.Day != "Mon" {
		t.Fatalf("SwapDays = %+v, %+v, %v, want a unplanned and b on Mon", a, b, ok)
	}
	if _, _, ok := s.SwapDays("a", "a"); ok {
		t.Error("swapping a recipe with itself should fail")
	}
	if _, _, ok := s.SwapDays("a", "missing"); ok {
		t.Error("swapping with a missing recipe should fail")
	}
}

func TestReplaceRecipeKeepsDay(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", Day: "Wed"}}}

	s.ReplaceRecipe("a", Recipe{Id: "b", Day: "Fri"})

	if s.Recipes[0].Id != "b" || s.Recipes[0].Day != "Wed" {
		t.Errorf("after replace got %+v, want b on Wed", s.Recipes[0])
	}
}

func TestAddRecipesStartUnplanned(t *testing.T) {
	var s State
	s.AddRecipes([]Recipe{{Id: "a", Day: "Sat"}})

	if s.Recipes[0].Day != "" {
		t.Errorf("added recipe has Day %q, want unplanned", s.Recipes[0].Day)
	}
}

func TestByDay(t *testing.T) {
	got := ByDay([]Recipe{{Id: "none"}, {Id: "sun", Day: "Sun"}, {Id: "mon1", Day: "Mon"}, {Id: "mon2", Day: "Mon"}})

	want := []string{"mon1", "mon2", "sun", "none"}
	for i, r := range got {
		if r.Id != want[i] {
			t.Fatalf("ByDay order = %v, want %v", ids(got), want)
		}
	}
}

func ids(recipes []Recipe) []string {
	out := make([]string, len(recipes))
	for i, r := range recipes {
		out[i] = r.Id
	}
	return out
}
