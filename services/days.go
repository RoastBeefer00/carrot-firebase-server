package services

import (
	"slices"
	"time"
)

// DateLayout is how a recipe's Day is stored.
const DateLayout = "2006-01-02"

// Today is the current date in the named time zone, such as "America/Chicago",
// or in UTC when the zone is empty or unknown.
func Today(zone string) time.Time {
	loc, err := time.LoadLocation(zone)
	if zone == "" || err != nil {
		loc = time.UTC
	}
	y, m, d := time.Now().In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// WeekStart returns the Monday of the week containing day.
func WeekStart(day time.Time) time.Time {
	offset := (int(day.Weekday()) + 6) % 7 // Monday is 0
	return day.AddDate(0, 0, -offset)
}

// WeekDays returns the seven dates of the week starting on monday.
func WeekDays(monday time.Time) []time.Time {
	days := make([]time.Time, 7)
	for i := range days {
		days[i] = monday.AddDate(0, 0, i)
	}
	return days
}

// ParseDay reads a stored Day. Empty and malformed values aren't dates.
func ParseDay(day string) (time.Time, bool) {
	t, err := time.Parse(DateLayout, day)
	return t, err == nil
}

// ValidDay reports whether day is a date, or empty for no day.
func ValidDay(day string) bool {
	_, ok := ParseDay(day)
	return day == "" || ok
}

// legacyWeekdays are the values Day held before it stored dates.
var legacyWeekdays = map[string]int{"Mon": 0, "Tue": 1, "Wed": 2, "Thu": 3, "Fri": 4, "Sat": 5, "Sun": 6}

// MigrateLegacyDays turns days saved as weekday names ("Mon") into that day's
// date in the week starting on monday, and reports whether anything changed.
func (s *State) MigrateLegacyDays(monday time.Time) bool {
	changed := false
	for i, r := range s.Recipes {
		if offset, ok := legacyWeekdays[r.Day]; ok {
			s.Recipes[i].Day = monday.AddDate(0, 0, offset).Format(DateLayout)
			changed = true
		} else if r.Day != "" && !ValidDay(r.Day) {
			s.Recipes[i].Day = ""
			changed = true
		}
	}
	return changed
}

// ByDay returns the recipes in date order, with unplanned recipes last. Recipes
// on the same day keep the order they were added in.
func ByDay(recipes []Recipe) []Recipe {
	sorted := slices.Clone(recipes)
	slices.SortStableFunc(sorted, func(a, b Recipe) int {
		switch {
		case a.Day == b.Day:
			return 0
		case a.Day == "":
			return 1
		case b.Day == "":
			return -1
		case a.Day < b.Day: // dates in DateLayout sort as strings
			return -1
		default:
			return 1
		}
	})
	return sorted
}
