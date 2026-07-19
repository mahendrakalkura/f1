package main

import "testing"

var (
	fixtureVerstappen = driver{DriverID: "max_verstappen", GivenName: "Max", FamilyName: "Verstappen"}
	fixtureNorris     = driver{DriverID: "lando_norris", GivenName: "Lando", FamilyName: "Norris"}
	fixtureRedBull    = constructor{ConstructorID: "red_bull", Name: "Red Bull"}
	fixtureMcLaren    = constructor{ConstructorID: "mclaren", Name: "McLaren"}
)

// fixtureData builds a two-round season: Verstappen wins round one (with the
// sprint and fastest lap), Norris wins round two; one pole each.
func fixtureData() *data {
	driverStandings := mrDataResponse{}
	driverStandings.MRData.StandingsTable.Round = "2"
	driverStandings.MRData.StandingsTable.Season = "2024"
	driverStandings.MRData.StandingsTable.StandingsLists = []standingsList{{
		DriverStandings: []driverStanding{
			{Constructors: []constructor{fixtureRedBull}, Driver: fixtureVerstappen, Points: "51", Position: "1", Wins: "1"},
			{Constructors: []constructor{fixtureMcLaren}, Driver: fixtureNorris, Points: "43", Position: "2", Wins: "1"},
		},
	}}

	constructorStandings := mrDataResponse{}
	constructorStandings.MRData.StandingsTable.StandingsLists = []standingsList{{
		ConstructorStandings: []constructorStanding{
			{Constructor: fixtureRedBull, Points: "51", Position: "1", Wins: "1"},
			{Constructor: fixtureMcLaren, Points: "43", Position: "2", Wins: "1"},
		},
	}}

	races := mrDataResponse{}
	races.MRData.RaceTable.Races = []race{
		{
			Circuit:  circuit{CircuitName: "Bahrain International Circuit", Location: location{Locality: "Sakhir", Country: "Bahrain"}},
			Date:     "2024-03-02",
			RaceName: "Bahrain Grand Prix",
			Round:    "1",
		},
		{
			Circuit:  circuit{CircuitName: "Shanghai International Circuit", Location: location{Locality: "Shanghai", Country: "China"}},
			Date:     "2024-04-21",
			RaceName: "Chinese Grand Prix",
			Round:    "2",
		},
	}

	results := []race{
		{
			Round: "1",
			Results: []result{
				{Constructor: fixtureRedBull, Driver: fixtureVerstappen, FastestLap: fastestLap{Rank: "1"}, Grid: "1", Points: "25", Position: "1", PositionText: "1", Status: "Finished"},
				{Constructor: fixtureMcLaren, Driver: fixtureNorris, Grid: "3", Points: "18", Position: "2", PositionText: "2", Status: "Finished"},
			},
		},
		{
			Round: "2",
			Results: []result{
				{Constructor: fixtureMcLaren, Driver: fixtureNorris, FastestLap: fastestLap{Rank: "1"}, Grid: "1", Points: "25", Position: "1", PositionText: "1", Status: "Finished"},
				{Constructor: fixtureRedBull, Driver: fixtureVerstappen, Grid: "2", Points: "18", Position: "2", PositionText: "2", Status: "Finished"},
			},
		},
	}

	qualifying := []race{
		{Round: "1", QualifyingResults: []qualifyingResult{{Constructor: fixtureRedBull, Driver: fixtureVerstappen, Position: "1"}}},
		{Round: "2", QualifyingResults: []qualifyingResult{{Constructor: fixtureMcLaren, Driver: fixtureNorris, Position: "1"}}},
	}

	sprints := []race{
		{
			Round: "1",
			SprintResults: []result{
				{Constructor: fixtureRedBull, Driver: fixtureVerstappen, FastestLap: fastestLap{Rank: "1"}, Grid: "4", Points: "8", Position: "1", PositionText: "1", Status: "Finished"},
				{Constructor: fixtureMcLaren, Driver: fixtureNorris, Grid: "2", Points: "7", Position: "2", PositionText: "2", Status: "Finished"},
			},
		},
	}

	roundOne := mrDataResponse{}
	roundOne.MRData.StandingsTable.StandingsLists = []standingsList{{
		ConstructorStandings: []constructorStanding{
			{Constructor: fixtureRedBull, Points: "33", Position: "1"},
			{Constructor: fixtureMcLaren, Points: "25", Position: "2"},
		},
	}}
	roundTwo := mrDataResponse{}
	roundTwo.MRData.StandingsTable.StandingsLists = []standingsList{{
		ConstructorStandings: []constructorStanding{
			{Constructor: fixtureRedBull, Points: "51", Position: "1"},
			{Constructor: fixtureMcLaren, Points: "43", Position: "2"},
		},
	}}

	return buildData(
		driverStandings,
		constructorStandings,
		races,
		results,
		qualifying,
		sprints,
		[]mrDataResponse{roundOne, roundTwo},
	)
}

func TestBuildDataChart(t *testing.T) {
	model := fixtureData()
	if len(model.progression.chart) != 2 {
		t.Fatalf("got %d chart rows, want 2", len(model.progression.chart))
	}

	leader := model.progression.chart[0]
	if leader.label != "Max Verstappen" {
		t.Errorf("got leader %q, want %q", leader.label, "Max Verstappen")
	}
	// Race one win plus sprint win: 25 + 8 = 33; second place in round two: +18.
	want := []float64{33, 51}
	for index, value := range want {
		if leader.points[index] != value {
			t.Errorf("leader.points[%d] = %v, want %v", index, leader.points[index], value)
		}
	}
}

func TestBuildDataConstructors(t *testing.T) {
	model := fixtureData()
	if len(model.constructors) != 2 {
		t.Fatalf("got %d constructors, want 2", len(model.constructors))
	}

	leader := model.constructors[0]
	if leader.name != "Red Bull" || leader.points != 51 || leader.poles != 1 || leader.fastestLaps != 1 {
		t.Errorf("unexpected leader row: %+v", leader)
	}
	if len(leader.drivers) != 1 || leader.drivers[0] != "Max Verstappen" {
		t.Errorf("unexpected drivers: %v", leader.drivers)
	}
}

func TestBuildDataDrivers(t *testing.T) {
	model := fixtureData()
	if model.season != "2024" {
		t.Errorf("got season %q, want %q", model.season, "2024")
	}
	if len(model.drivers) != 2 {
		t.Fatalf("got %d drivers, want 2", len(model.drivers))
	}

	leader := model.drivers[0]
	if leader.name != "Max Verstappen" || leader.team != "Red Bull" {
		t.Errorf("unexpected leader: %+v", leader)
	}
	if leader.points != 51 || leader.wins != 1 || leader.poles != 1 || leader.fastestLaps != 1 {
		t.Errorf("unexpected leader totals: %+v", leader)
	}
}

func TestBuildDataProgression(t *testing.T) {
	model := fixtureData()
	if model.progression.rounds != 2 {
		t.Fatalf("got %d rounds, want 2", model.progression.rounds)
	}
	if model.progression.raceLabels[0] != "BAH" || model.progression.raceLabels[1] != "CHI" {
		t.Errorf("unexpected labels: %v", model.progression.raceLabels)
	}

	leader := model.progression.drivers[0]
	if leader.cells[0].category != categoryWin || leader.cells[0].text != "1" {
		t.Errorf("unexpected round one cell: %+v", leader.cells[0])
	}
	if leader.cells[1].category != categorySecond || leader.cells[1].text != "2" {
		t.Errorf("unexpected round two cell: %+v", leader.cells[1])
	}
	if leader.team != "Red Bull" {
		t.Errorf("driver series missing team: %+v", leader)
	}

	constructors := model.progression.constructors[0]
	if constructors.cells[0].text != "33" || constructors.cells[1].text != "51" {
		t.Errorf("unexpected constructor cells: %+v", constructors.cells)
	}
	if len(constructors.drivers) != 1 || constructors.drivers[0] != "Max Verstappen" {
		t.Errorf("constructor series missing drivers: %+v", constructors)
	}
}

func TestBuildDataRaces(t *testing.T) {
	model := fixtureData()
	if len(model.races) != 2 {
		t.Fatalf("got %d races, want 2", len(model.races))
	}

	first := model.races[0]
	if first.winner != "Max Verstappen" || first.winnerTeam != "Red Bull" {
		t.Errorf("unexpected winner: %+v", first)
	}
	if first.pole != "Max Verstappen" || first.fastestLap != "Max Verstappen" {
		t.Errorf("unexpected headlines: %+v", first)
	}
	if first.circuit != "Bahrain International Circuit" || first.date != "2024-03-02" || first.location != "Sakhir, Bahrain" {
		t.Errorf("unexpected metadata: %+v", first)
	}
	if len(first.results) != 2 || first.results[0].driver != "Max Verstappen" {
		t.Errorf("unexpected results: %+v", first.results)
	}

	second := model.races[1]
	if second.winner != "Lando Norris" || second.pole != "Lando Norris" {
		t.Errorf("unexpected round two headlines: %+v", second)
	}
}

func TestBuildDataSprints(t *testing.T) {
	model := fixtureData()
	if len(model.sprints) != 1 {
		t.Fatalf("got %d sprints, want 1", len(model.sprints))
	}

	sprint := model.sprints[0]
	if sprint.name != "Bahrain Grand Prix" || sprint.circuit != "Bahrain International Circuit" {
		t.Errorf("unexpected sprint metadata: %+v", sprint)
	}
	if sprint.winner != "Max Verstappen" || sprint.winnerTeam != "Red Bull" || sprint.fastestLap != "Max Verstappen" {
		t.Errorf("unexpected sprint headlines: %+v", sprint)
	}
	if len(sprint.results) != 2 || sprint.results[1].driver != "Lando Norris" {
		t.Errorf("unexpected sprint results: %+v", sprint.results)
	}
}

func TestCellText(t *testing.T) {
	cases := []struct {
		positionText string
		classified   bool
		want         string
	}{
		{"4", true, "4"},
		{"D", false, "DSQ"},
		{"E", false, "EX"},
		{"F", false, "DNQ"},
		{"N", false, "NC"},
		{"W", false, "DNS"},
		{"R", false, "Ret"},
	}
	for _, c := range cases {
		if got := cellText(c.positionText, c.classified); got != c.want {
			t.Errorf("cellText(%q, %v) = %q, want %q", c.positionText, c.classified, got, c.want)
		}
	}
}

func TestDriverCell(t *testing.T) {
	cases := []struct {
		name     string
		entry    result
		category resultCategory
		text     string
	}{
		{"absent", result{}, categoryNone, ""},
		{"win", result{Points: "25", Position: "1", PositionText: "1"}, categoryWin, "1"},
		{"second", result{Points: "18", Position: "2", PositionText: "2"}, categorySecond, "2"},
		{"third", result{Points: "15", Position: "3", PositionText: "3"}, categoryThird, "3"},
		{"points", result{Points: "2", Position: "9", PositionText: "9"}, categoryPoints, "9"},
		{"no points", result{Points: "0", Position: "14", PositionText: "14"}, categoryFinished, "14"},
		{"retired", result{PositionText: "R"}, categoryRetired, "Ret"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cell := driverCell(c.entry)
			if cell.category != c.category || cell.text != c.text {
				t.Errorf("got %+v, want category %v text %q", cell, c.category, c.text)
			}
		})
	}
}

func TestDriverName(t *testing.T) {
	if got := driverName(fixtureVerstappen); got != "Max Verstappen" {
		t.Errorf("got %q, want %q", got, "Max Verstappen")
	}
}

func TestFormatPoints(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{25, "25"},
		{12.5, "12.5"},
	}
	for _, c := range cases {
		if got := formatPoints(c.in); got != c.want {
			t.Errorf("formatPoints(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsClassified(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"1", true},
		{"12", true},
		{"", false},
		{"R", false},
		{"1R", false},
	}
	for _, c := range cases {
		if got := isClassified(c.in); got != c.want {
			t.Errorf("isClassified(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRaceAbbrev(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Australian Grand Prix", "AUS"},
		{"Austrian Grand Prix", "AUT"},
		{"Bahrain Grand Prix", "BAH"},
		{"Abu Dhabi Grand Prix", "ABU"},
		{"Monaco Grand Prix", "MON"},
		{"", "?"},
	}
	for _, c := range cases {
		if got := raceAbbrev(c.in); got != c.want {
			t.Errorf("raceAbbrev(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRoundPointsByEntity(t *testing.T) {
	withList := mrDataResponse{}
	withList.MRData.StandingsTable.StandingsLists = []standingsList{{
		ConstructorStandings: []constructorStanding{{Constructor: fixtureRedBull, Points: "33"}},
	}}
	empty := mrDataResponse{}

	points := roundPointsByEntity([]mrDataResponse{withList, empty}, 3, constructorPointsFromList)
	if points[0]["red_bull"] != 33 {
		t.Errorf("round one points = %v, want 33", points[0]["red_bull"])
	}
	if len(points[1]) != 0 {
		t.Errorf("round with empty lists = %v, want empty", points[1])
	}
	if len(points[2]) != 0 {
		t.Errorf("round beyond responses = %v, want empty", points[2])
	}
}
