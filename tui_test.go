package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func readyModel(t *testing.T) appModel {
	t.Helper()
	model := newAppModel(fixtureData())
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	app, ok := updated.(appModel)
	if !ok {
		t.Fatal("Update did not return an appModel")
	}
	return app
}

func press(t *testing.T, model appModel, msg tea.KeyMsg) appModel {
	t.Helper()
	updated, _ := model.Update(msg)
	app, ok := updated.(appModel)
	if !ok {
		t.Fatal("Update did not return an appModel")
	}
	return app
}

func TestProgLabelsWidth(t *testing.T) {
	model := fixtureData()

	got := progLabelsWidth(model.progression.drivers, modeDrivers)
	want := len("Max Verstappen") + len("Red Bull")
	if got != want {
		t.Errorf("drivers mode: got %d, want %d", got, want)
	}

	got = progLabelsWidth(model.progression.constructors, modeConstructors)
	want = len("Constructor") + len("Max Verstappen")
	if got != want {
		t.Errorf("constructors mode: got %d, want %d", got, want)
	}

	got = progLabelsWidth(nil, modeDrivers)
	want = len("Driver") + len("Team")
	if got != want {
		t.Errorf("empty series: got %d, want %d", got, want)
	}
}

func TestProgressionChartMode(t *testing.T) {
	model := readyModel(t)
	for range int(tabProgression) {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyTab})
	}
	if model.active != tabProgression {
		t.Fatalf("got tab %v, want tabProgression", model.active)
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if model.progMode != modeChart {
		t.Fatalf("got mode %v, want modeChart", model.progMode)
	}
	if got := model.maxProgOffset(); got != 0 {
		t.Errorf("chart mode maxProgOffset = %d, want 0", got)
	}

	view := model.View()
	if !strings.Contains(view, "Max Verstappen") {
		t.Errorf("chart view missing driver:\n%s", view)
	}
}

func TestProgressionOffsetBounds(t *testing.T) {
	model := readyModel(t)
	for range int(tabProgression) {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyTab})
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyLeft})
	if model.progOffset != 0 {
		t.Errorf("left at zero moved offset to %d", model.progOffset)
	}

	for range 10 {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyRight})
	}
	if model.progOffset != model.maxProgOffset() {
		t.Errorf("offset %d exceeds maximum %d", model.progOffset, model.maxProgOffset())
	}
}

func TestRaceListEnterEsc(t *testing.T) {
	model := readyModel(t)
	for range int(tabRaces) {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyTab})
	}
	if model.active != tabRaces {
		t.Fatalf("got tab %v, want tabRaces", model.active)
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.races.selected != 0 {
		t.Fatalf("enter did not select race, selected = %d", model.races.selected)
	}
	if body := model.body(); !strings.Contains(body, "Bahrain Grand Prix") || !strings.Contains(body, "2024-03-02") {
		t.Errorf("detail view missing heading or metadata:\n%s", body)
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyEscape})
	if model.races.selected != -1 {
		t.Errorf("esc did not deselect, selected = %d", model.races.selected)
	}
}

func TestRaceMeta(t *testing.T) {
	model := fixtureData()
	got := raceMeta(model.races[0])
	want := "2024-03-02 | Bahrain International Circuit | Sakhir, Bahrain"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if got := raceMeta(raceInfo{name: "Unknown"}); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestSprintListEnterEsc(t *testing.T) {
	model := readyModel(t)
	for range int(tabSprints) {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyTab})
	}
	if model.active != tabSprints {
		t.Fatalf("got tab %v, want tabSprints", model.active)
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.sprints.selected != 0 {
		t.Fatalf("enter did not select sprint, selected = %d", model.sprints.selected)
	}
	if body := model.body(); !strings.Contains(body, "Bahrain Grand Prix") {
		t.Errorf("sprint detail missing heading:\n%s", body)
	}
}

func TestTabCycling(t *testing.T) {
	model := readyModel(t)
	if model.active != tabDrivers {
		t.Fatalf("initial tab = %v, want tabDrivers", model.active)
	}

	model = press(t, model, tea.KeyMsg{Type: tea.KeyShiftTab})
	if model.active != tabProgression {
		t.Errorf("shift+tab wrapped to %v, want tabProgression", model.active)
	}

	for range tab(len(tabTitles)) {
		model = press(t, model, tea.KeyMsg{Type: tea.KeyTab})
	}
	if model.active != tabProgression {
		t.Errorf("full cycle landed on %v, want tabProgression", model.active)
	}
}

func TestTableHeightFloor(t *testing.T) {
	model := appModel{height: 4}
	if got := model.tableHeight(); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}
