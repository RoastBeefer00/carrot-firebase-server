package views

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
)

// ShellData is what the header, bottom navigation, and toast need on every page.
type ShellData struct {
	User         services.User
	Active       string // "plan", "recipes", "groceries", or "" for other pages
	GroceryCount int
	IsAdmin      bool
	// Toast is a short message about what the last action did, if anything.
	Toast string
}

// Week is the week a page is showing, Monday through Sunday.
type Week struct {
	Start time.Time // Monday
	Today time.Time
}

// Days returns the week's seven dates.
func (w Week) Days() []time.Time {
	return services.WeekDays(w.Start)
}

// Param is how the week appears in a URL or request: empty for the current
// week, so plain links always open on it.
func (w Week) Param() string {
	if w.Start.Equal(services.WeekStart(w.Today)) {
		return ""
	}
	return w.Start.Format(services.DateLayout)
}

func (w Week) shifted(weeks int) Week {
	return Week{Start: w.Start.AddDate(0, 0, 7*weeks), Today: w.Today}
}

// Label names the week relative to today, such as "Next week".
func (w Week) Label() string {
	switch int(w.Start.Sub(services.WeekStart(w.Today)).Hours() / 24 / 7) {
	case 0:
		return "This week"
	case 1:
		return "Next week"
	case -1:
		return "Last week"
	default:
		return "Week of " + w.Start.Format("Jan 2")
	}
}

// Range gives the week's dates, such as "Nov 9 – 15" or "Nov 30 – Dec 6".
func (w Week) Range() string {
	end := w.Start.AddDate(0, 0, 6)
	if end.Month() == w.Start.Month() {
		return fmt.Sprintf("%s – %d", w.Start.Format("Jan 2"), end.Day())
	}
	return fmt.Sprintf("%s – %s", w.Start.Format("Jan 2"), end.Format("Jan 2"))
}

// Contains reports whether a stored Day falls in this week.
func (w Week) Contains(day string) bool {
	t, ok := services.ParseDay(day)
	return ok && !t.Before(w.Start) && t.Before(w.Start.AddDate(0, 0, 7))
}

func (w Week) isToday(day time.Time) bool {
	return day.Equal(w.Today)
}

// PlanData is what the plan page shows.
type PlanData struct {
	Recipes  []services.Recipe
	ShowWife bool
	Week     Week
}

// elsewhere returns the planned recipes that aren't in the week shown.
func (d PlanData) elsewhere() []services.Recipe {
	var out []services.Recipe
	for _, r := range services.ByDay(d.Recipes) {
		if r.Day != "" && !d.Week.Contains(r.Day) {
			out = append(out, r)
		}
	}
	return out
}

// RecipesData is what the recipes page shows.
type RecipesData struct {
	List      string // "week" or "fav"
	Meals     []services.Recipe
	Favorites []services.Recipe
	Selected  *services.Recipe
	Week      Week
	// Detail shows the selected recipe instead of the list on a phone.
	Detail bool
}

// Shown returns the recipes in the list that's currently shown.
func (d RecipesData) Shown() []services.Recipe {
	if d.List == "fav" {
		return d.Favorites
	}
	return d.Meals
}

func (d RecipesData) selectedID() string {
	if d.Selected == nil || !d.Detail {
		return ""
	}
	return d.Selected.Id
}

func (d RecipesData) elsewhere() []services.Recipe {
	return PlanData{Recipes: d.Meals, Week: d.Week}.elsewhere()
}

// PlanURL is the address of the plan page for a week param.
func PlanURL(week string) string {
	if week == "" {
		return "/"
	}
	return "/?week=" + url.QueryEscape(week)
}

// RecipesURL is the address of the recipes page for a list, selected recipe,
// and week param.
func RecipesURL(list, id, week string) string {
	q := url.Values{}
	if id != "" {
		q.Set("id", id)
	}
	if list == "fav" {
		q.Set("list", "fav")
	}
	if week != "" {
		q.Set("week", week)
	}
	if len(q) == 0 {
		return "/recipes"
	}
	return "/recipes?" + q.Encode()
}

func recipesOn(recipes []services.Recipe, day string) []services.Recipe {
	var out []services.Recipe
	for _, r := range recipes {
		if r.Day == day {
			out = append(out, r)
		}
	}
	return out
}

func iso(day time.Time) string {
	return day.Format(services.DateLayout)
}

func planSummary(d PlanData) string {
	if len(d.Recipes) == 0 {
		return d.Week.Range()
	}
	planned := 0
	for _, r := range d.Recipes {
		if d.Week.Contains(r.Day) {
			planned++
		}
	}
	return fmt.Sprintf("%s · %d planned", d.Week.Range(), planned)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// DayLong names a stored Day in full, such as "Monday, Nov 9", or "" for none.
func DayLong(day string) string {
	t, ok := services.ParseDay(day)
	if !ok {
		return ""
	}
	return t.Format("Monday, Jan 2")
}

// dayShort names a stored Day briefly, such as "Mon, Nov 9".
func dayShort(day string) string {
	t, ok := services.ParseDay(day)
	if !ok {
		return ""
	}
	return t.Format("Mon, Jan 2")
}

func dayLabel(r services.Recipe) string {
	if r.Day == "" {
		return "Not planned yet"
	}
	return DayLong(r.Day)
}

func vals(pairs ...string) string {
	m := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return toJSON(m)
}

// dayChip names a stored Day for a day button, such as "Mon 9".
func dayChip(day string) string {
	t, ok := services.ParseDay(day)
	if !ok {
		return ""
	}
	return t.Format("Mon 2")
}

// dayChoices are the week's dates followed by "" for not planned.
func dayChoices(w Week) []string {
	choices := make([]string, 0, 8)
	for _, day := range w.Days() {
		choices = append(choices, iso(day))
	}
	return append(choices, "")
}
