package services

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestWeekStart(t *testing.T) {
	for day, want := range map[string]string{
		"2026-11-07": "2026-11-02", // Saturday
		"2026-11-08": "2026-11-02", // Sunday ends the week
		"2026-11-09": "2026-11-09", // Monday starts one
		"2026-12-31": "2026-12-28",
	} {
		if got := WeekStart(date(day)).Format(DateLayout); got != want {
			t.Errorf("WeekStart(%s) = %s, want %s", day, got, want)
		}
	}
}

func TestWeekDays(t *testing.T) {
	days := WeekDays(date("2026-11-09"))
	if len(days) != 7 || days[0].Format(DateLayout) != "2026-11-09" || days[6].Format(DateLayout) != "2026-11-15" {
		t.Errorf("WeekDays(2026-11-09) = %v, want Nov 9 through Nov 15", days)
	}
}

func TestToday(t *testing.T) {
	if got, want := Today("not/a-zone"), Today("UTC"); !got.Equal(want) {
		t.Errorf("Today with an unknown zone = %v, want UTC's %v", got, want)
	}
	if got := Today("America/Chicago"); got.Hour() != 0 || got.Location() != time.UTC {
		t.Errorf("Today should be a bare date, got %v", got)
	}
}

func TestValidDay(t *testing.T) {
	for day, want := range map[string]bool{"": true, "2026-11-09": true, "Mon": false, "2026-13-01": false, "soon": false} {
		if got := ValidDay(day); got != want {
			t.Errorf("ValidDay(%q) = %v, want %v", day, got, want)
		}
	}
}

func TestMigrateLegacyDays(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", Day: "Mon"}, {Id: "b", Day: "Sun"}, {Id: "c", Day: "2026-11-20"}, {Id: "d"}, {Id: "e", Day: "junk"}}}

	if !s.MigrateLegacyDays(date("2026-11-02")) {
		t.Fatal("MigrateLegacyDays reported no change")
	}
	want := []string{"2026-11-02", "2026-11-08", "2026-11-20", "", ""}
	for i, r := range s.Recipes {
		if r.Day != want[i] {
			t.Errorf("recipe %s Day = %q, want %q", r.Id, r.Day, want[i])
		}
	}
	if s.MigrateLegacyDays(date("2026-11-02")) {
		t.Error("migrating again should change nothing")
	}
}

func TestSetDay(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a"}, {Id: "b", Day: "2026-11-10"}}}

	if r, ok := s.SetDay("a", "2026-11-09"); !ok || r.Day != "2026-11-09" {
		t.Fatalf("SetDay(a, Nov 9) = %+v, %v", r, ok)
	}
	if r, _ := s.SetDay("b", ""); r.Day != "" {
		t.Errorf("SetDay(b, \"\") left Day = %q, want unplanned", r.Day)
	}
	if _, ok := s.SetDay("missing", "2026-11-09"); ok {
		t.Error("SetDay(missing) = true, want false")
	}
}

func TestSwapDays(t *testing.T) {
	s := State{Recipes: []Recipe{{Id: "a", Day: "2026-11-09"}, {Id: "b"}}}

	a, b, ok := s.SwapDays("a", "b")
	if !ok || a.Day != "" || b.Day != "2026-11-09" {
		t.Fatalf("SwapDays = %+v, %+v, %v, want a unplanned and b on Nov 9", a, b, ok)
	}
	if _, _, ok := s.SwapDays("a", "a"); ok {
		t.Error("swapping a recipe with itself should fail")
	}
	if _, _, ok := s.SwapDays("a", "missing"); ok {
		t.Error("swapping with a missing recipe should fail")
	}
}

func TestByDay(t *testing.T) {
	got := ByDay([]Recipe{
		{Id: "none"}, {Id: "next", Day: "2026-11-16"}, {Id: "mon1", Day: "2026-11-09"}, {Id: "mon2", Day: "2026-11-09"},
	})

	want := []string{"mon1", "mon2", "next", "none"}
	for i, r := range got {
		if r.Id != want[i] {
			t.Fatalf("ByDay order = %v, want %v", ids(got), want)
		}
	}
}
