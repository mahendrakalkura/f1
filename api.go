package main

import (
	"fmt"
	"strconv"
)

const baseURL = "https://api.jolpi.ca/ergast/f1"

const pageLimit = 100

type mrDataResponse struct {
	MRData struct {
		RaceTable      raceTable      `json:"RaceTable"`
		StandingsTable standingsTable `json:"StandingsTable"`
	} `json:"MRData"`
}

type standingsTable struct {
	Round          string          `json:"round"`
	Season         string          `json:"season"`
	StandingsLists []standingsList `json:"StandingsLists"`
}

type standingsList struct {
	ConstructorStandings []constructorStanding `json:"ConstructorStandings"`
	DriverStandings      []driverStanding      `json:"DriverStandings"`
}

type constructorStanding struct {
	Constructor constructor `json:"Constructor"`
	Points      string      `json:"points"`
	Position    string      `json:"position"`
	Wins        string      `json:"wins"`
}

type driverStanding struct {
	Constructors []constructor `json:"Constructors"`
	Driver       driver        `json:"Driver"`
	Points       string        `json:"points"`
	Position     string        `json:"position"`
	Wins         string        `json:"wins"`
}

type raceTable struct {
	Races []race `json:"Races"`
}

type race struct {
	Circuit           circuit            `json:"Circuit"`
	Date              string             `json:"date"`
	QualifyingResults []qualifyingResult `json:"QualifyingResults"`
	RaceName          string             `json:"raceName"`
	Results           []result           `json:"Results"`
	Round             string             `json:"round"`
	SprintResults     []result           `json:"SprintResults"`
}

type result struct {
	Constructor  constructor `json:"Constructor"`
	Driver       driver      `json:"Driver"`
	FastestLap   fastestLap  `json:"FastestLap"`
	Grid         string      `json:"grid"`
	Points       string      `json:"points"`
	Position     string      `json:"position"`
	PositionText string      `json:"positionText"`
	Status       string      `json:"status"`
}

type qualifyingResult struct {
	Constructor constructor `json:"Constructor"`
	Driver      driver      `json:"Driver"`
	Position    string      `json:"position"`
}

type fastestLap struct {
	Rank string `json:"rank"`
}

type circuit struct {
	CircuitName string   `json:"circuitName"`
	Location    location `json:"Location"`
}

type location struct {
	Country  string `json:"country"`
	Locality string `json:"locality"`
}

type constructor struct {
	ConstructorID string `json:"constructorId"`
	Name          string `json:"name"`
}

type driver struct {
	DriverID   string `json:"driverId"`
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

func constructorStandingsURL(season string) string {
	return fmt.Sprintf("%s/%s/constructorStandings/?limit=%d", baseURL, season, pageLimit)
}

func driverStandingsURL(season string) string {
	return fmt.Sprintf("%s/%s/driverStandings/?limit=%d", baseURL, season, pageLimit)
}

func racesURL(season string) string {
	return fmt.Sprintf("%s/%s/races/?limit=%d", baseURL, season, pageLimit)
}

func roundConstructorStandingsURL(season string, round int) string {
	return fmt.Sprintf("%s/%s/%d/constructorStandings/?limit=%d", baseURL, season, round, pageLimit)
}

func roundQualifyingURL(season string, round int) string {
	return fmt.Sprintf("%s/%s/%d/qualifying/?limit=%d", baseURL, season, round, pageLimit)
}

func roundResultsURL(season string, round int) string {
	return fmt.Sprintf("%s/%s/%d/results/?limit=%d", baseURL, season, round, pageLimit)
}

func roundSprintURL(season string, round int) string {
	return fmt.Sprintf("%s/%s/%d/sprint/?limit=%d", baseURL, season, round, pageLimit)
}

func parseFloat(text string) float64 {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0
	}
	return value
}

func parseInt(text string) int {
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}
	return value
}

// seasonSlug maps the --season flag to an API path segment; zero or negative
// means the API's rolling "current" season alias.
func seasonSlug(year int) string {
	if year <= 0 {
		return "current"
	}
	return strconv.Itoa(year)
}
