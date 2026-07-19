package main

import (
	"fmt"
	"sort"
	"strings"
)

type resultCategory int

const (
	categoryNone resultCategory = iota
	categoryWin
	categorySecond
	categoryThird
	categoryPoints
	categoryFinished
	categoryRetired
)

type data struct {
	constructors []constructorRow
	drivers      []driverRow
	progression  progression
	races        []raceInfo
	season       string
}

type constructorRow struct {
	drivers     []string
	fastestLaps int
	id          string
	name        string
	points      float64
	poles       int
	position    int
	wins        int
}

type driverRow struct {
	fastestLaps int
	id          string
	name        string
	points      float64
	poles       int
	position    int
	team        string
	wins        int
}

type raceInfo struct {
	fastestLap string
	name       string
	pole       string
	results    []raceResult
	round      int
	winner     string
	winnerTeam string
}

type raceResult struct {
	driver   string
	grid     int
	points   float64
	position int
	status   string
	team     string
}

type progression struct {
	constructors []seriesRow
	drivers      []seriesRow
	raceLabels   []string
	rounds       int
}

type seriesRow struct {
	cells       []matrixCell
	fastestLaps int
	label       string
	poles       int
	total       float64
	wins        int
}

type matrixCell struct {
	category resultCategory
	text     string
}

type seasonStats struct {
	fastestLapByDriver map[string]int
	fastestLapByRound  map[int]string
	fastestLapByTeam   map[string]int
	poleByDriver       map[string]int
	poleByRound        map[int]string
	poleByTeam         map[string]int
	winnerByRound      map[int]string
	winnerTeamByRound  map[int]string
}

// buildData turns the raw API responses into the display model, wiring the
// driver-to-team and team-to-driver cross references, the per-round series,
// and the season totals for wins, poles, and fastest laps.
func buildData(
	driverStandings mrDataResponse,
	constructorStandings mrDataResponse,
	races mrDataResponse,
	results []race,
	qualifying []race,
	roundConstructorStandings []mrDataResponse,
) *data {
	stats := aggregateStats(results, qualifying)
	teamDrivers := map[string][]string{}

	drivers := []driverRow{}
	if len(driverStandings.MRData.StandingsTable.StandingsLists) > 0 {
		for _, standing := range driverStandings.MRData.StandingsTable.StandingsLists[0].DriverStandings {
			team := ""
			teamID := ""
			if len(standing.Constructors) > 0 {
				last := standing.Constructors[len(standing.Constructors)-1]
				team = last.Name
				teamID = last.ConstructorID
			}

			name := driverName(standing.Driver)
			teamDrivers[teamID] = append(teamDrivers[teamID], name)

			row := driverRow{
				fastestLaps: stats.fastestLapByDriver[standing.Driver.DriverID],
				id:          standing.Driver.DriverID,
				name:        name,
				points:      parseFloat(standing.Points),
				poles:       stats.poleByDriver[standing.Driver.DriverID],
				position:    parseInt(standing.Position),
				team:        team,
				wins:        parseInt(standing.Wins),
			}
			drivers = append(drivers, row)
		}
	}

	constructors := []constructorRow{}
	if len(constructorStandings.MRData.StandingsTable.StandingsLists) > 0 {
		for _, standing := range constructorStandings.MRData.StandingsTable.StandingsLists[0].ConstructorStandings {
			row := constructorRow{
				drivers:     teamDrivers[standing.Constructor.ConstructorID],
				fastestLaps: stats.fastestLapByTeam[standing.Constructor.ConstructorID],
				id:          standing.Constructor.ConstructorID,
				name:        standing.Constructor.Name,
				points:      parseFloat(standing.Points),
				poles:       stats.poleByTeam[standing.Constructor.ConstructorID],
				position:    parseInt(standing.Position),
				wins:        parseInt(standing.Wins),
			}
			constructors = append(constructors, row)
		}
	}

	raceList, resultsByRound := buildRaces(races, results, stats)

	// The standings can report a round ahead of the results endpoint, so the
	// number of scored rounds is taken from the races that actually returned
	// results rather than from the standings round counter.
	completedRounds := 0
	for round := range resultsByRound {
		if round > completedRounds {
			completedRounds = round
		}
	}

	series := buildProgression(
		drivers,
		constructors,
		raceList,
		resultsByRound,
		roundConstructorStandings,
		completedRounds,
	)

	model := &data{
		constructors: constructors,
		drivers:      drivers,
		progression:  series,
		races:        raceList,
		season:       driverStandings.MRData.StandingsTable.Season,
	}
	return model
}

// aggregateStats walks every race and qualifying session once to tally season
// totals and per-round headline names (winner, pole, fastest lap).
func aggregateStats(results []race, qualifying []race) seasonStats {
	stats := seasonStats{
		fastestLapByDriver: map[string]int{},
		fastestLapByRound:  map[int]string{},
		fastestLapByTeam:   map[string]int{},
		poleByDriver:       map[string]int{},
		poleByRound:        map[int]string{},
		poleByTeam:         map[string]int{},
		winnerByRound:      map[int]string{},
		winnerTeamByRound:  map[int]string{},
	}

	for _, item := range results {
		round := parseInt(item.Round)
		for _, entry := range item.Results {
			if entry.Position == "1" {
				stats.winnerByRound[round] = driverName(entry.Driver)
				stats.winnerTeamByRound[round] = entry.Constructor.Name
			}
			if entry.FastestLap.Rank == "1" {
				stats.fastestLapByDriver[entry.Driver.DriverID]++
				stats.fastestLapByTeam[entry.Constructor.ConstructorID]++
				stats.fastestLapByRound[round] = driverName(entry.Driver)
			}
		}
	}

	for _, item := range qualifying {
		round := parseInt(item.Round)
		for _, entry := range item.QualifyingResults {
			if entry.Position != "1" {
				continue
			}
			stats.poleByDriver[entry.Driver.DriverID]++
			stats.poleByTeam[entry.Constructor.ConstructorID]++
			stats.poleByRound[round] = driverName(entry.Driver)
		}
	}

	return stats
}

// buildRaces merges schedule metadata with finishing results and headline
// names, keyed by round, returning the ordered list and a lookup by driver.
func buildRaces(races mrDataResponse, results []race, stats seasonStats) ([]raceInfo, map[int]map[string]result) {
	infoByRound := map[int]*raceInfo{}
	order := []int{}

	for _, item := range races.MRData.RaceTable.Races {
		round := parseInt(item.Round)
		info := &raceInfo{
			name:  item.RaceName,
			round: round,
		}
		infoByRound[round] = info
		order = append(order, round)
	}

	resultsByRound := map[int]map[string]result{}
	for _, item := range results {
		round := parseInt(item.Round)

		info, ok := infoByRound[round]
		if !ok {
			info = &raceInfo{name: item.RaceName, round: round}
			infoByRound[round] = info
			order = append(order, round)
		}

		byDriver := map[string]result{}
		for _, entry := range item.Results {
			byDriver[entry.Driver.DriverID] = entry

			row := raceResult{
				driver:   driverName(entry.Driver),
				grid:     parseInt(entry.Grid),
				points:   parseFloat(entry.Points),
				position: parseInt(entry.Position),
				status:   entry.Status,
				team:     entry.Constructor.Name,
			}
			info.results = append(info.results, row)
		}
		resultsByRound[round] = byDriver
	}

	for round, info := range infoByRound {
		info.fastestLap = stats.fastestLapByRound[round]
		info.pole = stats.poleByRound[round]
		info.winner = stats.winnerByRound[round]
		info.winnerTeam = stats.winnerTeamByRound[round]
	}

	sort.Ints(order)
	list := []raceInfo{}
	for _, round := range order {
		list = append(list, *infoByRound[round])
	}
	return list, resultsByRound
}

// buildProgression assembles the Wikipedia-style matrix rows for drivers
// (colored finishing positions) and constructors (cumulative points), each
// carrying its season totals for the summary columns.
func buildProgression(
	drivers []driverRow,
	constructors []constructorRow,
	races []raceInfo,
	resultsByRound map[int]map[string]result,
	roundConstructorStandings []mrDataResponse,
	completedRounds int,
) progression {
	labels := make([]string, completedRounds)
	for _, info := range races {
		if info.round >= 1 && info.round <= completedRounds {
			labels[info.round-1] = raceAbbrev(info.name)
		}
	}

	constructorPoints := roundPointsByEntity(roundConstructorStandings, completedRounds, constructorPointsFromList)

	driverSeries := []seriesRow{}
	for _, row := range drivers {
		cells := make([]matrixCell, completedRounds)
		for round := 1; round <= completedRounds; round++ {
			cells[round-1] = driverCell(resultsByRound[round][row.id])
		}

		series := seriesRow{
			cells:       cells,
			fastestLaps: row.fastestLaps,
			label:       row.name,
			poles:       row.poles,
			total:       row.points,
			wins:        row.wins,
		}
		driverSeries = append(driverSeries, series)
	}

	constructorSeries := []seriesRow{}
	for _, row := range constructors {
		cells := make([]matrixCell, completedRounds)
		for round := 1; round <= completedRounds; round++ {
			value := constructorPoints[round-1][row.id]
			cells[round-1] = matrixCell{category: categoryNone, text: formatPoints(value)}
		}

		series := seriesRow{
			cells:       cells,
			fastestLaps: row.fastestLaps,
			label:       row.name,
			poles:       row.poles,
			total:       row.points,
			wins:        row.wins,
		}
		constructorSeries = append(constructorSeries, series)
	}

	result := progression{
		constructors: constructorSeries,
		drivers:      driverSeries,
		raceLabels:   labels,
		rounds:       completedRounds,
	}
	return result
}

func roundPointsByEntity(
	responses []mrDataResponse,
	completedRounds int,
	extract func(standingsList) map[string]float64,
) []map[string]float64 {
	perRound := make([]map[string]float64, completedRounds)
	for round := 1; round <= completedRounds; round++ {
		perRound[round-1] = map[string]float64{}
		if round-1 >= len(responses) {
			continue
		}
		lists := responses[round-1].MRData.StandingsTable.StandingsLists
		if len(lists) == 0 {
			continue
		}
		perRound[round-1] = extract(lists[0])
	}
	return perRound
}

func constructorPointsFromList(list standingsList) map[string]float64 {
	points := map[string]float64{}
	for _, standing := range list.ConstructorStandings {
		points[standing.Constructor.ConstructorID] = parseFloat(standing.Points)
	}
	return points
}

func driverCell(entry result) matrixCell {
	if entry.Position == "" && entry.PositionText == "" {
		return matrixCell{category: categoryNone, text: ""}
	}

	category := categoryFinished
	position := parseInt(entry.Position)
	points := parseFloat(entry.Points)
	classified := isClassified(entry.PositionText)

	switch {
	case !classified:
		category = categoryRetired
	case position == 1:
		category = categoryWin
	case position == 2:
		category = categorySecond
	case position == 3:
		category = categoryThird
	case points > 0:
		category = categoryPoints
	}

	cell := matrixCell{
		category: category,
		text:     cellText(entry.PositionText, classified),
	}
	return cell
}

func cellText(positionText string, classified bool) string {
	if classified {
		return positionText
	}

	switch positionText {
	case "D":
		return "DSQ"
	case "E":
		return "EX"
	case "F":
		return "DNQ"
	case "N":
		return "NC"
	case "W":
		return "DNS"
	default:
		return "Ret"
	}
}

func isClassified(positionText string) bool {
	if positionText == "" {
		return false
	}
	for _, char := range positionText {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func driverName(entry driver) string {
	return fmt.Sprintf("%s %s", entry.GivenName, entry.FamilyName)
}

// raceCodes disambiguates race names whose three-letter prefixes collide
// (for example Australian and Austrian both abbreviate to AUS).
var raceCodes = map[string]string{
	"Australian": "AUS",
	"Austrian":   "AUT",
}

func raceAbbrev(name string) string {
	trimmed := strings.TrimSpace(strings.TrimSuffix(name, "Grand Prix"))
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "?"
	}

	code, ok := raceCodes[fields[0]]
	if ok {
		return code
	}

	word := fields[0]
	if len(word) > 3 {
		word = word[:3]
	}
	return strings.ToUpper(word)
}

func formatPoints(points float64) string {
	if points == float64(int(points)) {
		return fmt.Sprintf("%d", int(points))
	}
	return fmt.Sprintf("%.1f", points)
}
