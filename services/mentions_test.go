package services

import (
	"strings"
	"testing"
)

// mentions renders a step with each mention as [text=ingredient line].
func mentions(ingredients []string, step string) string {
	var b strings.Builder
	for _, p := range StepParts(ingredients, step) {
		if p.Ingredient < 0 {
			b.WriteString(p.Text)
		} else {
			b.WriteString("[" + p.Text + "=" + ingredients[p.Ingredient] + "]")
		}
	}
	return b.String()
}

func TestStepParts(t *testing.T) {
	ribs := []string{
		"4 lb bone-in short ribs", "2 tbsp olive oil", "1 onion, diced", "3 carrots, chopped",
		"4 cloves garlic", "2 cups dry red wine", "2 cups beef stock", "2 sprigs rosemary",
		"1 lemon, zested", "1/4 cup chopped parsley", "Salt", "Pepper",
	}
	chicken := []string{"6 chicken thighs", "1/2 cup chicken stock", "2 tbsp butter", "Salt and pepper to taste"}
	stocks := []string{"1 cup beef stock", "1 cup chicken stock"}

	tests := []struct {
		name        string
		ingredients []string
		step, want  string
	}{
		{"short and plural names", ribs,
			"Add the onion and Carrots with a pinch of salt.",
			"Add the [onion=1 onion, diced] and [Carrots=3 carrots, chopped] with a pinch of [salt=Salt]."},
		{"last words of a name", ribs,
			"Heat the oil, then sear the short ribs. Pour in the wine and the stock.",
			"Heat the [oil=2 tbsp olive oil], then sear the [short ribs=4 lb bone-in short ribs]. Pour in the [wine=2 cups dry red wine] and the [stock=2 cups beef stock]."},
		{"whole name wins over a shorter one", ribs,
			"Add red wine, then lemon zest and parsley.",
			"Add [red wine=2 cups dry red wine], then [lemon=1 lemon, zested] zest and [parsley=1/4 cup chopped parsley]."},
		{"vague words don't match alone", ribs,
			"Bring to a boil until dry and red.",
			"Bring to a boil until dry and red."},
		{"shared first word matches neither", chicken,
			"Season the chicken, then add the chicken stock and butter.",
			"Season the chicken, then add the [chicken stock=1/2 cup chicken stock] and [butter=2 tbsp butter]."},
		{"first word alone", chicken[:1],
			"Season the chicken.",
			"Season the [chicken=6 chicken thighs]."},
		{"shared last word matches neither", stocks,
			"Add the stock and the beef stock.",
			"Add the stock and the [beef stock=1 cup beef stock]."},
		{"a pairing splits", chicken,
			"Season with salt and pepper.",
			"Season with [salt=Salt and pepper to taste] and [pepper=Salt and pepper to taste]."},
		{"no phrase across punctuation", []string{"1 tbsp olive oil", "1 olive"},
			"Add the olive, oil.",
			"Add the [olive=1 olive], [oil=1 tbsp olive oil]."},
		{"no ingredients", nil, "Preheat the oven.", "Preheat the oven."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mentions(tt.ingredients, tt.step); got != tt.want {
				t.Errorf("\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
