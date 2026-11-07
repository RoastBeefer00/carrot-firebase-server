package services

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Combining ingredients favors leaving two lines over merging wrongly: a shopper can
// spot and combine two similar lines, but can't see a wrong total. So names only match
// exactly after normalizing, and amounts only add when they measure the same thing.

type measure int

const (
	weight measure = iota
	volume
	packaged
	counted
)

type unit struct {
	singular, plural string
	measure          measure
	// base is the unit's size in milliliters or grams, for volume and weight.
	base   float64
	metric bool
}

// Sizes derive from exact relationships, so 16 tablespoons is exactly 1 cup.
const (
	teaspoonML = 4.92892159375
	ounceG     = 28.349523125
)

var (
	teaspoon   = unit{"tsp", "tsp", volume, teaspoonML, false}
	tablespoon = unit{"tbsp", "tbsp", volume, 3 * teaspoonML, false}
	cup        = unit{"cup", "cups", volume, 48 * teaspoonML, false}
	fluidOunce = unit{"fl oz", "fl oz", volume, 6 * teaspoonML, false}
	pint       = unit{"pint", "pints", volume, 96 * teaspoonML, false}
	quart      = unit{"quart", "quarts", volume, 192 * teaspoonML, false}
	milliliter = unit{"ml", "ml", volume, 1, true}
	liter      = unit{"l", "l", volume, 1000, true}
	ounce      = unit{"oz", "oz", weight, ounceG, false}
	pound      = unit{"lb", "lb", weight, 16 * ounceG, false}
	gram       = unit{"g", "g", weight, 1, true}
	kilogram   = unit{"kg", "kg", weight, 1000, true}
)

func packageUnit(singular, plural string) unit {
	return unit{singular: singular, plural: plural, measure: packaged}
}

var units = map[string]unit{
	"tsp": teaspoon, "tsps": teaspoon, "teaspoon": teaspoon, "teaspoons": teaspoon,
	"tbsp": tablespoon, "tbsps": tablespoon, "tablespoon": tablespoon, "tablespoons": tablespoon,
	"cup": cup, "cups": cup,
	"ml": milliliter, "milliliter": milliliter, "milliliters": milliliter, "millilitre": milliliter, "millilitres": milliliter,
	"l": liter, "liter": liter, "liters": liter, "litre": liter, "litres": liter,
	"pint": pint, "pints": pint, "quart": quart, "quarts": quart,
	"oz": ounce, "ozs": ounce, "ounce": ounce, "ounces": ounce,
	"lb": pound, "lbs": pound, "pound": pound, "pounds": pound,
	"g": gram, "gram": gram, "grams": gram,
	"kg": kilogram, "kilogram": kilogram, "kilograms": kilogram,
}

func init() {
	for _, p := range [][2]string{
		{"can", "cans"}, {"jar", "jars"}, {"package", "packages"}, {"packet", "packets"},
		{"bottle", "bottles"}, {"box", "boxes"}, {"bag", "bags"}, {"clove", "cloves"},
		{"head", "heads"}, {"bunch", "bunches"}, {"sprig", "sprigs"}, {"stalk", "stalks"},
		{"stick", "sticks"}, {"slice", "slices"}, {"pinch", "pinches"}, {"dash", "dashes"},
		{"piece", "pieces"}, {"ear", "ears"}, {"block", "blocks"}, {"container", "containers"},
	} {
		units[p[0]] = packageUnit(p[0], p[1])
		units[p[1]] = packageUnit(p[0], p[1])
	}
	units["pkg"] = units["package"]
	units["pack"] = units["package"]
	units["packs"] = units["package"]
}

// droppedWords don't change what you'd buy, so they're ignored when they lead a name.
// Words like "diced", "crushed" and "ground" stay because they describe a different
// product: diced tomatoes aren't tomatoes, and ground beef isn't beef.
var droppedWords = map[string]bool{
	"large": true, "medium": true, "small": true, "extra-large": true, "ripe": true,
	"fresh": true, "freshly": true, "chopped": true, "minced": true, "finely": true,
	"roughly": true, "coarsely": true, "thinly": true, "grated": true, "shredded": true,
	"crumbled": true, "melted": true, "softened": true, "peeled": true, "organic": true,
}

// trailingWords are preparation written after a name without a comma, as in "4 cloves
// garlic minced". After the name they can't be part of a product's name.
var trailingWords = map[string]bool{
	"minced": true, "diced": true, "chopped": true, "sliced": true, "drained": true,
	"rinsed": true, "peeled": true, "juiced": true, "zested": true, "grated": true,
	"crushed": true, "cubed": true, "halved": true, "quartered": true, "trimmed": true,
	"thinly": true, "finely": true, "roughly": true, "and": true, "&": true,
}

// sizeWords can come before a unit, as in "1 small can" or "4 large cloves".
var sizeWords = map[string]bool{"small": true, "medium": true, "large": true, "extra-large": true}

// unitsAfterName are also written after the name, as in "2 celery stalks".
var unitsAfterName = map[string]bool{"stalk": true, "head": true, "sprig": true}

// keyFolding makes spellings of the same thing compare equal.
var keyFolding = strings.NewReplacer("ñ", "n", "é", "e", "®", "", "™", "", "-", " ", "’", "'")

// sameThing maps a normalized name to the one it's bought as.
var sameThing = map[string]string{
	"all-purpose flour":      "flour",
	"all purpose flour":      "flour",
	"kosher salt":            "salt",
	"sea salt":               "salt",
	"table salt":             "salt",
	"black pepper":           "pepper",
	"ground black pepper":    "pepper",
	"ground pepper":          "pepper",
	"extra virgin olive oil": "olive oil",
	"jalapeno pepper":        "jalapeno",
	"evoo":                   "olive oil",
	"scallion":               "green onion",
	"spring onion":           "green onion",
	"chili garlic sauce":     "chile garlic sauce",
	"veggie broth":           "vegetable broth",
	"cilantro leaf":          "cilantro",
	"parsley leaf":           "parsley",
	"salt and black pepper":  "salt and pepper",
	"granulated sugar":       "sugar",
	"white sugar":            "sugar",
	"powdered sugar":         "confectioners sugar",
	"confectioners' sugar":   "confectioners sugar",
}

var (
	unicodeFractions = strings.NewReplacer(
		"⅛", " 1/8", "¼", " 1/4", "⅓", " 1/3", "⅜", " 3/8", "½", " 1/2",
		"⅝", " 5/8", "⅔", " 2/3", "¾", " 3/4", "⅞", " 7/8", "⁄", "/",
	)
	number = `(?:\d+(?:\.\d+)?|\.\d+)(?:/\d+)?`
	// The fractional part of a mixed amount: the "1/2" of "1 1/2", or the ".5" of "1 .5".
	fractionPart = `(?:\s+(?:\d+/\d+|\.\d+))?`
	// A leading amount: "2", ".5", "1/2", "1 1/2", or a range such as "1-2" or "2 to 3".
	amountRE = regexp.MustCompile(`^(` + number + fractionPart + `)(?:\s*(?:-|–|—|to)\s*(` + number + fractionPart + `))?(?:\s+|$)`)
	// A package size after the amount, as in "2 (14 oz) cans", "1 14.5-oz can", or
	// "3 (13 oz/400g) cans".
	sizeRE    = regexp.MustCompile(`^\(?(` + number + `)[\s-]*(oz|ounces?|lbs?|pounds?|g|grams?|ml|l)\.?(?:\s*/[^)]*)?\)?\s+(\S+)\s`)
	fluidOzRE = regexp.MustCompile(`(?i)^(?:fl\.?\s*oz\.?|fluid\s+ounces?)\s+`)
	// An amount written as a word: "Pinch of salt".
	wordAmountRE = regexp.MustCompile(`(?i)^(pinch|dash) of `)
	// Lines that came along with an ingredient list but aren't ingredients.
	notIngredientRE = regexp.MustCompile(`(?i)^\**\s*(?:$|allergens?\b|directions? for\b|optional\b|not in|one head\b|for full\b|.*:$)`)
	// A unit written against its amount: "50g", "16oz", "500-gram", "14. oz".
	attachedUnitRE = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\.?-?\s*(g|grams?|kg|ml|oz|ounces?|lbs?|pounds?)\b`)
	parensRE       = regexp.MustCompile(`\([^)]*\)`)
	trailerRE      = regexp.MustCompile(`(?i)\s*(?:\bto taste\b|\bfor (?:garnish|serving)\b|\boptional\b|\bdivided\b|\bas needed\b|\bif needed\b|\bsee notes?\b.*|\bfor frying\b|\b(?:plus|or) more\b.*)\s*$`)
	// Commas inside a description rather than before a preparation.
	listedDescriptorRE = regexp.MustCompile(`(?i)\b(boneless|skinless|bone-in|skin-on|unsalted|salted|low-sodium|lean),\s*`)
)

// parsed is one ingredient line split into what's needed.
type parsed struct {
	amount float64
	unit   unit
	// size distinguishes package sizes, as in the "14 oz" of "2 (14 oz) cans".
	size string
	name string
}

func parseIngredient(line string) []parsed {
	text := strings.Join(strings.Fields(unicodeFractions.Replace(line)), " ")
	if notIngredientRE.MatchString(text) {
		return nil
	}
	text = strings.TrimLeft(text, "-*• ")
	// A trailing star marks a pantry staple in some recipes.
	text = strings.TrimRight(text, "* ")
	text = attachedUnitRE.ReplaceAllString(text, "$1 $2")

	p := parsed{unit: unit{measure: counted}}
	if m := amountRE.FindStringSubmatch(text); m != nil {
		// A range takes its upper end, so there's enough either way.
		amount := m[1]
		if m[2] != "" {
			amount = m[2]
		}
		p.amount = ParseQuantity(amount)
		text = text[len(m[0]):]
	} else if lower := strings.ToLower(text); strings.HasPrefix(lower, "a ") || strings.HasPrefix(lower, "an ") {
		p.amount = 1
		text = text[strings.Index(text, " ")+1:]
	} else if m := wordAmountRE.FindStringSubmatch(text); m != nil {
		p.amount = 1
		text = m[1] + " " + text[len(m[0]):]
	}

	if !strings.Contains(text, " ") {
		// Nothing can follow a unit, so this is only a name.
	} else if first, rest, _ := strings.Cut(text, " "); sizeWords[strings.ToLower(first)] {
		if u, _, _ := strings.Cut(rest, " "); isUnit(u) {
			text = rest
		}
	}

	if m := sizeRE.FindStringSubmatch(text + " "); m != nil {
		if u, ok := units[strings.ToLower(strings.TrimRight(m[3], "."))]; ok && u.measure == packaged {
			size, _ := units[strings.ToLower(m[2])]
			p.size = m[1] + " " + size.singular
			p.unit = u
			text = strings.TrimSpace((text + " ")[len(m[0]):])
		}
	}
	if m := fluidOzRE.FindString(text); m != "" && len(m) < len(text) {
		p.unit, text = fluidOunce, text[len(m):]
	} else if p.size == "" {
		fields := strings.SplitN(text, " ", 2)
		if u, ok := lookupUnit(fields[0]); ok && len(fields) == 2 {
			p.unit = u
			text = fields[1]
		}
	}
	if rest, ok := cutPrefixFold(text, "of "); ok && p.unit.measure != counted {
		text = rest
	}

	name := cleanName(text)
	if p.size == "" && p.unit.measure == counted {
		// "2 garlic cloves" means the same as "2 cloves garlic". Other cloves are a spice.
		if nameKey(name) == "garlic clove" {
			name, p.unit = name[:len("garlic")], units["clove"]
		} else if i := strings.LastIndex(name, " "); i > 0 && unitsAfterName[singularWord(strings.ToLower(name[i+1:]))] {
			p.unit = units[singularWord(strings.ToLower(name[i+1:]))]
			name = name[:i]
		}
	}
	if name == "" {
		name = strings.TrimSpace(line)
	}
	p.name = name

	// The one common pairing written as a single ingredient.
	if p.amount == 0 {
		switch nameKey(name) {
		case "salt and pepper", "salt & pepper":
			return []parsed{{unit: p.unit, name: "salt"}, {unit: p.unit, name: "pepper"}}
		}
	}
	return []parsed{p}
}

func isUnit(token string) bool {
	_, ok := lookupUnit(token)
	return ok
}

func lookupUnit(token string) (unit, bool) {
	token = strings.TrimRight(token, ".")
	// Recipes use case to tell these apart.
	switch token {
	case "T":
		return tablespoon, true
	case "t":
		return teaspoon, true
	}
	u, ok := units[strings.ToLower(token)]
	return u, ok
}

// cleanName removes what doesn't affect shopping: notes and preparation.
func cleanName(text string) string {
	text = parensRE.ReplaceAllString(text, " ")
	text = listedDescriptorRE.ReplaceAllString(text, "$1 ")
	if i := strings.Index(text, ","); i >= 0 {
		text = text[:i]
	}
	if i := strings.Index(text, " - "); i >= 0 {
		text = text[:i]
	}
	text = trailerRE.ReplaceAllString(text, "")
	fields := strings.Fields(text)
	for len(fields) > 1 && droppedWords[strings.ToLower(fields[0])] {
		fields = fields[1:]
	}
	for len(fields) > 1 && trailingWords[strings.ToLower(fields[len(fields)-1])] {
		fields = fields[:len(fields)-1]
	}
	return strings.Join(fields, " ")
}

func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return s, false
}

var (
	irregularSingular = map[string]string{"leaves": "leaf", "loaves": "loaf", "cookies": "cookie", "pies": "pie", "brownies": "brownie"}
	irregularPlural   = map[string]string{"leaf": "leaves", "loaf": "loaves", "potato": "potatoes", "tomato": "tomatoes", "mango": "mangoes"}
)

func singularWord(w string) string {
	lower := strings.ToLower(w)
	if s, ok := irregularSingular[lower]; ok {
		return s
	}
	switch {
	case strings.HasSuffix(lower, "ies") && len(lower) > 4:
		return w[:len(w)-3] + "y"
	case strings.HasSuffix(lower, "oes"), strings.HasSuffix(lower, "ches"),
		strings.HasSuffix(lower, "shes"), strings.HasSuffix(lower, "xes"):
		return w[:len(w)-2]
	case strings.HasSuffix(lower, "ss"), strings.HasSuffix(lower, "us"), strings.HasSuffix(lower, "is"):
		return w
	case strings.HasSuffix(lower, "s") && len(lower) > 2:
		return w[:len(w)-1]
	}
	return w
}

func pluralWord(w string) string {
	lower := strings.ToLower(w)
	if p, ok := irregularPlural[lower]; ok {
		return p
	}
	switch {
	case strings.HasSuffix(lower, "y") && len(lower) > 1 && !strings.ContainsRune("aeiou", rune(lower[len(lower)-2])):
		return w[:len(w)-1] + "ies"
	case strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"), strings.HasSuffix(lower, "x"), strings.HasSuffix(lower, "s"):
		return w + "es"
	}
	return w + "s"
}

// withLastWord applies f to the last word of a name, leaving the rest alone.
func withLastWord(name string, f func(string) string) string {
	i := strings.LastIndex(name, " ")
	return name[:i+1] + f(name[i+1:])
}

// nameKey is what two names must share to count as the same ingredient.
func nameKey(name string) string {
	key := withLastWord(keyFolding.Replace(strings.ToLower(name)), singularWord)
	if same, ok := sameThing[key]; ok {
		return same
	}
	return key
}

type amountKey struct {
	measure measure
	unit    string
	size    string
}

type groceryAmount struct {
	key amountKey
	// total is in the base unit for volume and weight, otherwise a plain count.
	total    float64
	unit     unit
	imperial bool
}

// GroceryItem is one line of the grocery list.
type GroceryItem struct {
	name string
	// shared means the name came from sameThing, so spellings can't replace it.
	shared  bool
	amounts []groceryAmount
}

func (g *GroceryItem) add(p parsed) {
	if p.amount <= 0 {
		return
	}
	key := amountKey{measure: p.unit.measure, size: p.size}
	total := p.amount
	if p.unit.measure == volume || p.unit.measure == weight {
		total *= p.unit.base
	} else {
		key.unit = p.unit.singular
	}
	for i := range g.amounts {
		if g.amounts[i].key == key {
			g.amounts[i].total += total
			g.amounts[i].imperial = g.amounts[i].imperial || !p.unit.metric
			return
		}
	}
	g.amounts = append(g.amounts, groceryAmount{key: key, total: total, unit: p.unit, imperial: !p.unit.metric})
	sort.SliceStable(g.amounts, func(i, j int) bool {
		a, b := g.amounts[i].key, g.amounts[j].key
		if a.measure != b.measure {
			return a.measure < b.measure
		}
		return a.size < b.size
	})
}

// Format renders the line for a recipe multiplier, such as "1 cup + 2 tbsp butter".
func (g GroceryItem) Format(multiplier float64) string {
	name := g.name
	var parts []string
	for _, a := range g.amounts {
		total := a.total * multiplier
		switch a.key.measure {
		case volume:
			parts = append(parts, formatVolume(total, a.imperial))
		case weight:
			parts = append(parts, formatWeight(total, a.imperial))
		case packaged:
			label := a.unit.singular
			if total > 1 {
				label = a.unit.plural
			}
			if a.key.size != "" {
				label = "(" + a.key.size + ") " + label
			}
			parts = append(parts, formatAmount(total)+" "+label)
		case counted:
			parts = append(parts, formatAmount(total))
			// Only a bare count reads as a number of these: "3 eggs".
			if len(g.amounts) == 1 {
				name = withLastWord(g.name, singularWord)
				if total > 1 {
					name = withLastWord(name, pluralWord)
				}
			}
		}
	}
	if len(parts) == 0 {
		return name
	}
	return strings.Join(parts, " + ") + " " + name
}

// formatAmount renders a number, rounding up so there's never too little on the list.
func formatAmount(v float64) string {
	frac := v - math.Floor(v)
	if math.Abs(frac-1.0/3.0) < 0.01 || math.Abs(frac-2.0/3.0) < 0.01 {
		return FormatFraction(v)
	}
	// Conversions leave tiny remainders that shouldn't count as needing more.
	if nearest := math.Round(v*8) / 8; math.Abs(v-nearest) < 0.005 {
		return FormatFraction(nearest)
	}
	return FormatFraction(math.Ceil(v*8) / 8)
}

// formatOunces keeps decimals like 8.8 oz, because they match the package to look for.
func formatOunces(oz float64) string {
	if nearest := math.Round(oz*8) / 8; math.Abs(oz-nearest) < 0.005 {
		return FormatFraction(nearest) + " oz"
	}
	return strconv.FormatFloat(math.Ceil(oz*10-1e-6)/10, 'f', 1, 64) + " oz"
}

// neat reports whether v is a whole multiple of step, within measuring error.
func neat(v, step float64) bool {
	n := v / step
	return math.Abs(n-math.Round(n)) < 0.01
}

func formatVolume(ml float64, imperial bool) string {
	if !imperial {
		if ml >= liter.base {
			return formatAmount(ml/liter.base) + " l"
		}
		return strconv.FormatFloat(math.Ceil(ml-1e-6), 'f', 0, 64) + " ml"
	}
	cups := ml / cup.base
	if cups >= 1 {
		label := cup.plural
		if neat(cups, 1) && math.Round(cups) == 1 {
			label = cup.singular
		}
		return formatAmount(cups) + " " + label
	}
	if cups >= 0.25-1e-3 {
		if neat(cups, 0.25) || neat(cups, 1.0/3.0) {
			return formatAmount(cups) + " " + cup.singular
		}
		// Too much to count in teaspoons, and exactness doesn't matter when shopping.
		return formatAmount(math.Ceil(ml/tablespoon.base-1e-6)) + " tbsp"
	}
	if tbsp := ml / tablespoon.base; tbsp >= 1-1e-3 && neat(tbsp, 0.5) {
		return formatAmount(tbsp) + " tbsp"
	}
	if tsp := ml / teaspoon.base; tsp > 9 {
		return formatAmount(math.Ceil(ml/tablespoon.base-1e-6)) + " tbsp"
	} else {
		return formatAmount(tsp) + " tsp"
	}
}

func formatWeight(grams float64, imperial bool) string {
	if !imperial {
		if grams >= kilogram.base {
			return formatAmount(grams/kilogram.base) + " kg"
		}
		return strconv.FormatFloat(math.Ceil(grams-1e-6), 'f', 0, 64) + " g"
	}
	if pounds := grams / pound.base; pounds >= 1-1e-3 {
		return formatAmount(pounds) + " lb"
	}
	return formatOunces(grams / ounce.base)
}

// CombineIngredients turns ingredient lines from any number of recipes into a sorted
// grocery list.
func CombineIngredients(lines []string) []GroceryItem {
	byKey := map[string]*GroceryItem{}
	var keys []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		for _, p := range parseIngredient(line) {
			key := nameKey(p.name)
			item, ok := byKey[key]
			if !ok {
				item = &GroceryItem{name: p.name}
				byKey[key] = item
				keys = append(keys, key)
			}
			if _, renamed := sameThing[withLastWord(keyFolding.Replace(strings.ToLower(p.name)), singularWord)]; renamed {
				item.name, item.shared = key, true
			} else if !item.shared && preferName(p.name, item.name) {
				item.name = p.name
			}
			item.add(p)
		}
	}
	sort.Strings(keys)
	items := make([]GroceryItem, 0, len(keys))
	for _, key := range keys {
		items = append(items, *byKey[key])
	}
	return items
}

// preferName picks the better spelling of the same name: plural reads best with
// amounts, and lowercase reads best mid-line, as in "1 tsp salt" over "1 tsp Salt".
func preferName(candidate, current string) bool {
	if len(candidate) != len(current) {
		return len(candidate) > len(current)
	}
	return candidate == strings.ToLower(candidate) && current != candidate
}
