package main

import (
	"encoding/json"
	"fmt"

	"golang.org/x/sync/errgroup"
)

const fetchConcurrency = 4

type fetcher struct {
	cache  *cache
	force  bool
	season string
}

// loadData fetches every endpoint needed for the TUI. The standings, races,
// and per-round results, qualifying, sprints, and constructor standings are
// downloaded in parallel through a bounded worker pool, served from cache
// when fresh. Results and qualifying are fetched per round rather than from
// the aggregate endpoints, which lag a round behind and split rounds across
// pages.
func loadData(store *cache, force bool, season string) (*data, error) {
	f := &fetcher{cache: store, force: force, season: season}

	driverStandings := mrDataResponse{}
	constructorStandings := mrDataResponse{}
	races := mrDataResponse{}

	head := errgroup.Group{}
	head.SetLimit(fetchConcurrency)
	head.Go(func() error { return f.get(driverStandingsURL(f.season), &driverStandings) })
	head.Go(func() error { return f.get(constructorStandingsURL(f.season), &constructorStandings) })
	head.Go(func() error { return f.get(racesURL(f.season), &races) })

	err := head.Wait()
	if err != nil {
		return nil, err
	}

	completedRounds := parseInt(driverStandings.MRData.StandingsTable.Round)

	roundResults := make([]mrDataResponse, completedRounds)
	roundQualifying := make([]mrDataResponse, completedRounds)
	roundSprints := make([]mrDataResponse, completedRounds)
	roundConstructorStandings := make([]mrDataResponse, completedRounds)

	rest := errgroup.Group{}
	rest.SetLimit(fetchConcurrency)

	for round := 1; round <= completedRounds; round++ {
		index := round - 1
		rest.Go(func() error { return f.get(roundResultsURL(f.season, round), &roundResults[index]) })
		rest.Go(func() error { return f.get(roundQualifyingURL(f.season, round), &roundQualifying[index]) })
		rest.Go(func() error { return f.get(roundSprintURL(f.season, round), &roundSprints[index]) })
		rest.Go(func() error {
			return f.get(roundConstructorStandingsURL(f.season, round), &roundConstructorStandings[index])
		})
	}

	err = rest.Wait()
	if err != nil {
		return nil, err
	}

	results := []race{}
	for _, page := range roundResults {
		results = append(results, page.MRData.RaceTable.Races...)
	}

	qualifying := []race{}
	for _, page := range roundQualifying {
		qualifying = append(qualifying, page.MRData.RaceTable.Races...)
	}

	sprints := []race{}
	for _, page := range roundSprints {
		sprints = append(sprints, page.MRData.RaceTable.Races...)
	}

	model := buildData(
		driverStandings,
		constructorStandings,
		races,
		results,
		qualifying,
		sprints,
		roundConstructorStandings,
	)
	return model, nil
}

func (f *fetcher) get(url string, out *mrDataResponse) error {
	body, err := f.cache.get(url, f.force)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}
