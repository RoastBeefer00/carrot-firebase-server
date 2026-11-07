package services

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

func combined(lines ...string) []string {
	out := []string{}
	for _, item := range CombineIngredients(lines) {
		out = append(out, item.Format(1))
	}
	return out
}

func TestCombineIngredients(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		// Things that must stay apart.
		{"name inside another name", []string{"1 egg", "1 eggplant"}, []string{"1 egg", "1 eggplant"}},
		{"different milks", []string{"1 cup milk", "1 cup coconut milk"}, []string{"1 cup coconut milk", "1 cup milk"}},
		{"different meats", []string{"1 lb ground beef", "1 lb ground turkey"}, []string{"1 lb ground beef", "1 lb ground turkey"}},
		{"canned product and fresh", []string{"1 can diced tomatoes", "3 tomatoes, chopped"}, []string{"1 can diced tomatoes", "3 tomatoes"}},
		{"spice and garlic", []string{"4 cloves garlic", "1 tsp ground cloves"}, []string{"4 cloves garlic", "1 tsp ground cloves"}},
		{"whole spice isn't garlic", []string{"2 whole cloves", "2 cloves garlic"}, []string{"2 cloves garlic", "2 whole cloves"}},

		// Quantities parse fully.
		{"mixed fraction", []string{"1 1/2 cups flour"}, []string{"1½ cups flour"}},
		{"unicode fraction", []string{"1½ cups rice", "1 cup rice"}, []string{"2½ cups rice"}},
		{"decimal", []string{"0.5 cup sugar"}, []string{"½ cup sugar"}},
		{"range takes the upper end", []string{"1-2 jalapeños", "1 jalapeño"}, []string{"3 jalapeños"}},
		{"worded range", []string{"2 to 3 cups stock"}, []string{"3 cups stock"}},
		{"name starting with to", []string{"2 tomatoes"}, []string{"2 tomatoes"}},
		{"a and an", []string{"a pinch of salt", "an onion"}, []string{"1 onion", "1 pinch salt"}},
		{"unit inside a word", []string{"Pecans", "1/2 cup pecans"}, []string{"½ cup pecans"}},

		// Units convert within a measure.
		{"spoons", []string{"2 tbsp olive oil", "1 tsp olive oil"}, []string{"7 tsp olive oil"}},
		{"too many teaspoons", []string{"2 tbsp olive oil", "1 tsp olive oil", "1/4 cup olive oil"}, []string{"7 tbsp olive oil"}},
		{"to cups", []string{"8 tbsp sugar", "1/2 cup sugar"}, []string{"1 cup sugar"}},
		{"sixteen tablespoons", []string{"16 tbsp sugar"}, []string{"1 cup sugar"}},
		{"to pounds", []string{"1 lb ground beef", "8 oz ground beef"}, []string{"1½ lb ground beef"}},
		{"abbreviated with a period", []string{"1 lb. beef", "1 pound beef"}, []string{"2 lb beef"}},
		{"case matters for t", []string{"1 T sugar", "1 t sugar"}, []string{"4 tsp sugar"}},
		{"metric stays metric", []string{"200 g flour", "300 g flour"}, []string{"500 g flour"}},
		{"metric adds to imperial", []string{"1 lb flour", "454 g flour"}, []string{"2 lb flour"}},
		{"rounds up, never down", []string{"1 cup milk", "1 tsp milk"}, []string{"1⅛ cups milk"}},

		// Amounts that can't be added stay visible.
		{"weight and volume", []string{"1 lb butter", "2 tbsp butter"}, []string{"1 lb + 2 tbsp butter"}},
		{"stick and spoons", []string{"1 stick butter", "2 tbsp butter"}, []string{"2 tbsp + 1 stick butter"}},
		{"count and volume", []string{"2 carrots", "1 cup chopped carrots"}, []string{"1 cup + 2 carrots"}},
		{"head and cloves", []string{"1 head garlic", "2 cloves garlic"}, []string{"1 head + 2 cloves garlic"}},
		{"can sizes", []string{"2 (14 oz) cans tomatoes", "1 (28 oz) can tomatoes", "1 14-oz can tomatoes"}, []string{"3 (14 oz) cans + 1 (28 oz) can tomatoes"}},

		// Names normalize.
		{"plural", []string{"1 egg", "2 eggs"}, []string{"3 eggs"}},
		{"plural added", []string{"1 egg", "1 egg"}, []string{"2 eggs"}},
		{"size words", []string{"3 large eggs", "1 small egg"}, []string{"4 eggs"}},
		{"preparation", []string{"1 onion, diced", "2 onions"}, []string{"3 onions"}},
		{"different preparations", []string{"1 lemon, zested", "1 lemon, juiced"}, []string{"2 lemons"}},
		{"preparation before", []string{"1/4 cup chopped parsley", "2 tbsp fresh parsley"}, []string{"6 tbsp parsley"}},
		{"parenthetical note", []string{"2 onions (about 1 lb)", "1 onion"}, []string{"3 onions"}},
		{"comma in description", []string{"2 boneless, skinless chicken breasts", "1 boneless skinless chicken breast"}, []string{"3 boneless skinless chicken breasts"}},
		{"garlic either way", []string{"4 cloves garlic", "2 garlic cloves, minced"}, []string{"6 cloves garlic"}},
		{"same thing", []string{"3 scallions", "2 green onions"}, []string{"5 green onions"}},
		{"oil", []string{"1/4 cup extra virgin olive oil", "1/4 cup olive oil"}, []string{"½ cup olive oil"}},
		{"plural irregular", []string{"1 potato", "1 tomato", "1 potato", "1 tomato"}, []string{"2 potatoes", "2 tomatoes"}},
		{"not actually plural", []string{"1 cup hummus", "1 lb asparagus"}, []string{"1 lb asparagus", "1 cup hummus"}},

		// No amount means enough is already covered, or it just needs to be on the list.
		{"no amount", []string{"Salt"}, []string{"Salt"}},
		{"no amount with amount", []string{"Salt", "1 tsp kosher salt"}, []string{"1 tsp salt"}},
		{"to taste", []string{"Salt and pepper to taste", "1 tsp salt"}, []string{"pepper", "1 tsp salt"}},
		{"optional", []string{"1 cup walnuts, optional", "1 cup walnuts"}, []string{"2 cups walnuts"}},
		{"unrecognized stays as written", []string{"Juice of 1 lemon"}, []string{"Juice of 1 lemon"}},

		// Shapes from real recipes.
		{"no leading zero", []string{".5 cup brown lentils", ".25 cup brown lentils"}, []string{"¾ cup brown lentils"}},
		{"spaced decimal", []string{"1 .5 tsp cumin"}, []string{"1½ tsp cumin"}},
		{"pantry star", []string{"olive oil*", "1 tbsp olive oil"}, []string{"1 tbsp olive oil"}},
		{"pantry star on a pair", []string{"Salt and pepper*"}, []string{"pepper", "salt"}},
		{"double star", []string{"Garlic**", "2 cloves garlic"}, []string{"2 cloves garlic"}},
		{"not ingredients", []string{"**Allergens: Soy", "*Optional", "**one head of garlic included in one of the meal bags", "Directions for 4 servings indicated in parentheses.", "Optional toppings: berries", "For the sauce:"}, []string{}},
		{"size before unit", []string{"4 large cloves of garlic", "1 clove garlic"}, []string{"5 cloves garlic"}},
		{"unit after name", []string{"1 celery stalk", "2 celery stalks, diced", "1 stalk celery"}, []string{"4 stalks celery"}},
		{"trailing preparation", []string{"4 cloves garlic minced", "1 tbsp ginger minced", "1 tbsp ginger"}, []string{"4 cloves garlic", "2 tbsp ginger"}},
		{"drained and rinsed", []string{"1 15-ounce can chickpeas drained & rinsed", "2 (15 oz) cans of chickpeas, drained and rinsed"}, []string{"3 (15 oz) cans chickpeas"}},
		{"size with metric", []string{"3 (13 oz/400g) cans of chickpeas"}, []string{"3 (13 oz) cans chickpeas"}},
		{"fluid ounces", []string{"24 fl oz vegetable broth", "1 cup vegetable broth"}, []string{"4 cups vegetable broth"}},
		{"accents", []string{"1 jalapeno", "1 jalapeño"}, []string{"2 jalapeños"}},
		{"hyphens and marks", []string{"1 tbsp Soy-Free Vegenaise", "1 tbsp Soy Free Vegenaise®"}, []string{"2 tbsp Soy Free Vegenaise®"}},
		{"organic", []string{"1 lb organic carrots", "1 lb carrots"}, []string{"2 lb carrots"}},
		{"word amount", []string{"Pinch of salt", "1 pinch salt"}, []string{"2 pinches salt"}},
		{"if needed", []string{"salt if needed"}, []string{"salt"}},
		{"see notes", []string{"1 pint vegan ice cream see Notes for my favorites"}, []string{"2 cups vegan ice cream"}},

		{"leading star on an ingredient", []string{"*Salt and pepper", "**Garlic"}, []string{"Garlic", "pepper", "salt"}},
		{"unit against the amount", []string{"50g flour", "16oz tofu", "1lb tofu", "500-gram flour"}, []string{"550 g flour", "2 lb tofu"}},
		{"amount with a stray period", []string{"14. oz hearts of palm"}, []string{"14 oz hearts of palm"}},
		{"fraction slash", []string{"1⁄2 teaspoon salt"}, []string{"½ tsp salt"}},

		{"package weights stay as written", []string{"8.8 oz precooked brown rice", "13.4 oz chickpeas"}, []string{"13.4 oz chickpeas", "8.8 oz precooked brown rice"}},
		{"too many for teaspoons", []string{"2 tbsp olive oil", "5 tsp olive oil"}, []string{"4 tbsp olive oil"}},
		{"spellings", []string{"1 tbsp chili garlic sauce", "1 tbsp chile garlic sauce", "1 cup veggie broth", "1 cup vegetable broth"}, []string{"2 tbsp chile garlic sauce", "2 cups vegetable broth"}},
		{"salt and black pepper", []string{"salt and black pepper to taste"}, []string{"pepper", "salt"}},

		{"blank lines", []string{"", "  "}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := combined(tc.in...)
			if strings.Join(got, " | ") != strings.Join(tc.want, " | ") {
				t.Errorf("combined(%q)\n got %q\nwant %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatMultiplier(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"1 egg"}, "2 eggs"},
		{[]string{"1 cup milk"}, "2 cups milk"},
		{[]string{"1/2 cup milk"}, "1 cup milk"},
		{[]string{"1 can beans"}, "2 cans beans"},
		{[]string{"Salt"}, "Salt"},
		{[]string{"1 lb butter", "2 tbsp butter"}, "2 lb + ¼ cup butter"},
	}
	for _, tc := range cases {
		items := CombineIngredients(tc.in)
		if got := items[0].Format(2); got != tc.want {
			t.Errorf("%q doubled = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// The real recipes aren't in the repo, so this only runs where they've been added.
func realIngredients(t *testing.T) (lines []string, recipes []Recipe) {
	t.Helper()
	b, err := os.ReadFile("testdata/recipes.json")
	if os.IsNotExist(err) {
		t.Skip("testdata/recipes.json isn't present")
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &recipes); err != nil {
		t.Fatal(err)
	}
	for _, r := range recipes {
		lines = append(lines, r.Ingredients...)
	}
	return lines, recipes
}

func TestRealRecipesParse(t *testing.T) {
	lines, _ := realIngredients(t)
	startsWithAmount := regexp.MustCompile(`^\s*[\d.½¼¾⅓⅔⅛]`)
	var failures []string
	for _, line := range lines {
		for _, p := range parseIngredient(line) {
			switch {
			case startsWithAmount.MatchString(line) && p.amount == 0 && !strings.Contains(p.name, " x "):
				failures = append(failures, "no amount: "+line)
			case strings.Contains(p.name, "*"):
				failures = append(failures, "star in name: "+line)
			case startsWithAmount.MatchString(p.name):
				failures = append(failures, "amount in name: "+line)
			}
		}
	}
	const allowed = 20 // Hand-typed oddities like "1 1 Roma crunch lettuce head".
	if len(failures) > allowed {
		t.Errorf("%d of %d lines misparsed, more than %d allowed, for example:\n%s", len(failures), len(lines), allowed, strings.Join(failures[:allowed], "\n"))
	}
	t.Logf("%d of %d lines have an oddity", len(failures), len(lines))
}
