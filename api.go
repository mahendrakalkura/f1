package main

import (
	"fmt"
	"strconv"
)

const baseURL = "https://api.jolpi.ca/ergast/f1/current"

const pageLimit = 100

type mrDataResponse struct {
	MRData struct {
		Limit          string         `json:"limit"`
		Offset         string         `json:"offset"`
		RaceTable      raceTable      `json:"RaceTable"`
		StandingsTable standingsTable `json:"StandingsTable"`
		Total          string         `json:"total"`
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
	Round                string                `json:"round"`
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
	Races  []race `json:"Races"`
	Round  string `json:"round"`
	Season string `json:"season"`
}

type race struct {
	Circuit           circuit            `json:"Circuit"`
	Date              string             `json:"date"`
	QualifyingResults []qualifyingResult `json:"QualifyingResults"`
	RaceName          string             `json:"raceName"`
	Results           []result           `json:"Results"`
	Round             string             `json:"round"`
	Time              string             `json:"time"`
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
	CircuitID   string   `json:"circuitId"`
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
	Nationality   string `json:"nationality"`
}

type driver struct {
	Code        string `json:"code"`
	DriverID    string `json:"driverId"`
	FamilyName  string `json:"familyName"`
	GivenName   string `json:"givenName"`
	Nationality string `json:"nationality"`
}

func constructorStandingsURL() string {
	return fmt.Sprintf("%s/constructorStandings/?limit=%d", baseURL, pageLimit)
}

func driverStandingsURL() string {
	return fmt.Sprintf("%s/driverStandings/?limit=%d", baseURL, pageLimit)
}

func racesURL() string {
	return fmt.Sprintf("%s/races/?limit=%d", baseURL, pageLimit)
}

func roundConstructorStandingsURL(round int) string {
	return fmt.Sprintf("%s/%d/constructorStandings/?limit=%d", baseURL, round, pageLimit)
}

func roundDriverStandingsURL(round int) string {
	return fmt.Sprintf("%s/%d/driverStandings/?limit=%d", baseURL, round, pageLimit)
}

func roundQualifyingURL(round int) string {
	return fmt.Sprintf("%s/%d/qualifying/?limit=%d", baseURL, round, pageLimit)
}

func roundResultsURL(round int) string {
	return fmt.Sprintf("%s/%d/results/?limit=%d", baseURL, round, pageLimit)
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
