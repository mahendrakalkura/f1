package main

import "testing"

func TestConstructorStandingsURL(t *testing.T) {
	got := constructorStandingsURL("2024")
	want := "https://api.jolpi.ca/ergast/f1/2024/constructorStandings/?limit=100"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseFloat(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"25", 25},
		{"12.5", 12.5},
		{"", 0},
		{"abc", 0},
	}
	for _, c := range cases {
		if got := parseFloat(c.in); got != c.want {
			t.Errorf("parseFloat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"7", 7},
		{"", 0},
		{"abc", 0},
		{"1.5", 0},
	}
	for _, c := range cases {
		if got := parseInt(c.in); got != c.want {
			t.Errorf("parseInt(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRacesURL(t *testing.T) {
	got := racesURL("current")
	want := "https://api.jolpi.ca/ergast/f1/current/races/?limit=100"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRoundSprintURL(t *testing.T) {
	got := roundSprintURL("2024", 5)
	want := "https://api.jolpi.ca/ergast/f1/2024/5/sprint/?limit=100"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSeasonSlug(t *testing.T) {
	cases := []struct {
		year int
		want string
	}{
		{0, "current"},
		{-5, "current"},
		{2024, "2024"},
		{1950, "1950"},
	}
	for _, c := range cases {
		if got := seasonSlug(c.year); got != c.want {
			t.Errorf("seasonSlug(%d) = %q, want %q", c.year, got, c.want)
		}
	}
}
