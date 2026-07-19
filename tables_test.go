package main

import (
	"strings"
	"testing"
)

func TestRenderConstructorsTable(t *testing.T) {
	out := renderConstructorsTable(fixtureData())
	for _, want := range []string{"Constructor", "Red Bull", "McLaren", "51", "43"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderDriversTable(t *testing.T) {
	out := renderDriversTable(fixtureData())
	for _, want := range []string{"Driver", "Max Verstappen", "Lando Norris", "Red Bull", "51"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderProgressionTable(t *testing.T) {
	model := fixtureData()
	out := renderProgressionTable(model.progression.drivers, model.progression.raceLabels, "Driver", 0, 2)
	for _, want := range []string{"BAH", "CHI", "Max Verstappen", "Pts"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderProgressionTableEmpty(t *testing.T) {
	if got := renderProgressionTable(nil, nil, "Driver", 0, 0); got != "No completed rounds yet." {
		t.Errorf("got %q, want %q", got, "No completed rounds yet.")
	}
}

func TestRenderResultsTable(t *testing.T) {
	model := fixtureData()
	out := renderResultsTable(model.races[0])
	for _, want := range []string{"Position", "Max Verstappen", "Lando Norris", "Finished"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
