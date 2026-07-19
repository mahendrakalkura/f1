# AGENTS.md

## Project

Go terminal UI that displays a Formula 1 season: driver standings, constructor standings, all race and sprint results, and a Wikipedia-style points progression matrix plus a cumulative points chart. Data is fetched from the Jolpica F1 API in parallel and cached on disk for 24 hours. The season defaults to the current one and can be pinned with `--season <year>`.

## Key files

- `api.go` - Jolpica JSON types and endpoint URL builders
- `cache.go` - 24-hour file cache plus HTTP download with backoff retry and stale fallback
- `fetch.go` - parallel fetch orchestration via errgroup (per-round results, qualifying, sprints, and constructor standings)
- `main.go` - entry point, `--force` and `--season` flags
- `model.go` - domain model, driver/team cross references, season totals, progression series
- `progression.go` - progression legend, cumulative points sparkline chart, column-fit helper
- `styles.go` - Lipgloss styles and result-category colours
- `tables.go` - lipgloss/table builders for drivers, constructors, results, and progression
- `tui.go` - Bubble Tea model, tabs, viewports, and the shared race-list component behind the Races and Sprints tabs
- `*_test.go` - unit tests for every layer

## Dependencies

- Go 1.26 or newer
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/bubbles
- github.com/charmbracelet/lipgloss
- golang.org/x/sync/errgroup
- goimports (run by `make build`)
- golangci-lint (for `make lint`)

## Commands

- `make build` - run goimports and `go mod tidy`, then compile to `./main`
- `make run` - run with cache; `make run force` re-downloads
- `make test` - `go test ./...`
- `make lint` - `golangci-lint run`

## Conventions

- Flat package layout: every file is `package main` in the repository root, no subdirectories.
- Alphabetical ordering of types, constants, variables, and functions.
- Errors wrapped with context; no silently ignored errors.
- No magic numbers; named constants instead.
- Tests live alongside the code in `*_test.go` and run with `make test`; CI runs build, tests, and lint on every push and pull request.
