// Command seed loads sample recipes into the local Firestore emulator. It refuses
// to run anywhere else, and the production image never builds it.
package main

import (
	"log"
	"os"

	"github.com/RoastBeefer00/carrot-firebase-server/db"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
)

// idsDocID matches the document handlers.AddRecipeToDatabase appends to.
const idsDocID = "hHjqXrWMhH7WDTPEwlkN"

var recipes = []services.Recipe{
	{
		Id:   "seed-short-ribs",
		Name: "Red Wine Braised Short Ribs",
		Time: "3 hours",
		Ingredients: []string{
			"4 lb bone-in short ribs", "2 tbsp olive oil", "1 onion, diced", "3 carrots, chopped",
			"4 cloves garlic", "2 cups dry red wine", "2 cups beef stock", "2 sprigs rosemary",
			"1 lemon, zested", "1/4 cup chopped parsley", "Salt", "Pepper",
		},
		Steps: []string{
			"Preheat the oven to 325°F.",
			"Pat the short ribs dry and season generously on all sides with salt and pepper. Heat the oil in a large Dutch oven over medium-high heat until shimmering, then sear the ribs in batches, about 3 minutes per side, until deeply browned. Don't crowd the pot or the meat will steam instead of browning. Transfer the ribs to a plate and pour off all but two tablespoons of fat. Add the onion and carrots with a pinch of salt and cook, stirring often and scraping up the browned bits, until softened and starting to color, about 8 minutes. Add the garlic and cook until fragrant. Pour in the wine, bring to a boil, and reduce by half, about 10 minutes.",
			"Return the ribs to the pot, add the stock and rosemary, cover, and braise in the oven for 2½ hours.",
			"Mix the lemon zest and parsley, then scatter over the ribs and serve.",
		},
	},
	{
		Id:   "seed-lemon-chicken",
		Name: "Lemon Garlic Chicken Thighs",
		Time: "40 min",
		Ingredients: []string{
			"6 chicken thighs", "2 tbsp butter", "4 cloves garlic", "1 lemon, sliced",
			"1/2 cup chicken stock", "1 tsp dried thyme", "Salt",
		},
		Steps: []string{
			"Season the chicken with salt and thyme.",
			"Melt the butter in a skillet over medium-high heat and brown the chicken skin side down for 7 minutes.",
			"Flip the chicken, add the garlic, lemon, and stock, then simmer for 20 minutes until cooked through.",
		},
	},
	{
		Id:   "seed-black-bean-tacos",
		Name: "Black Bean Tacos",
		Time: "20 min",
		Ingredients: []string{
			"2 cans black beans", "1 tbsp olive oil", "1 tsp cumin", "8 corn tortillas",
			"1 avocado, sliced", "1 lime", "1/2 cup crumbled cotija",
		},
		Steps: []string{
			"Warm the oil in a saucepan, add the cumin and beans, and simmer for 10 minutes, mashing some of the beans.",
			"Char the tortillas directly over a burner or in a dry skillet.",
			"Fill the tortillas with beans, avocado, and cotija, then finish with lime juice.",
		},
	},
	{
		Id:   "seed-tomato-soup",
		Name: "Roasted Tomato Soup",
		Time: "1 hour",
		Ingredients: []string{
			"3 lb roma tomatoes", "1 onion, quartered", "6 cloves garlic", "3 tbsp olive oil",
			"3 cups vegetable stock", "1/2 cup heavy cream", "1 cup basil leaves",
		},
		Steps: []string{
			"Heat the oven to 425°F and roast the tomatoes, onion, and garlic with the oil for 40 minutes.",
			"Transfer to a pot, add the stock and basil, and simmer for 10 minutes.",
			"Blend until smooth and stir in the cream.",
		},
	},
	{
		Id:   "seed-banana-pancakes",
		Name: "Banana Pancakes",
		Time: "25 min",
		Ingredients: []string{
			"2 ripe bananas", "2 eggs", "1 1/2 cups flour", "1 cup milk", "2 tsp baking powder", "1 tbsp butter",
		},
		Steps: []string{
			"Mash the bananas, then whisk in the eggs and milk.",
			"Fold in the flour and baking powder until just combined.",
			"Cook in a buttered skillet over medium heat, about 2 minutes per side.",
		},
	},
}

func main() {
	if !db.DevSandbox() {
		log.Fatal("refusing to seed: needs APP_ENV=development, FIRESTORE_EMULATOR_HOST, and a demo- FIRESTORE_PROJECT_ID")
	}

	client, ctx, err := db.GetClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ids := make([]string, 0, len(recipes))
	for _, recipe := range recipes {
		for i, ingredient := range recipe.Ingredients {
			recipe.Ingredients[i] = services.NormalizeIngredient(ingredient)
		}
		if _, err := client.Collection("recipes").Doc(recipe.Id).Set(ctx, recipe); err != nil {
			log.Fatal(err)
		}
		ids = append(ids, recipe.Id)
	}
	// Same shape as handlers.IDs.
	idsDoc := struct{ IDs []string }{ids}
	if _, err := client.Collection("ids").Doc(idsDocID).Set(ctx, idsDoc); err != nil {
		log.Fatal(err)
	}
	log.Printf("Seeded %d recipes into emulator at %s", len(recipes), os.Getenv("FIRESTORE_EMULATOR_HOST"))
}
