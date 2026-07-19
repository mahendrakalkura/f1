# AGENTS.md

## Project

Go terminal UI that displays the current Formula 1 season: driver standings, constructor standings, all race results, and a Wikipedia-style points progression matrix. Data is fetched from the Jolpica F1 API in parallel and cached on disk for 24 hours.

## Key files

- `api.go` - Jolpica JSON types and endpoint URL builders
- `cache.go` - 24-hour file cache plus HTTP download with backoff retry
- `fetch.go` - parallel fetch orchestration via errgroup (per-round results, qualifying, and constructor standings)
- `main.go` - entry point and `--force` flag
- `model.go` - domain model, driver/team cross references, season totals, progression series
- `progression.go` - progression legend and column-fit helper
- `styles.go` - Lipgloss styles and result-category colours
- `tables.go` - lipgloss/table builders for drivers, constructors, results, and progression
- `tui.go` - Bubble Tea model, tabs, viewports, and the Races table

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
- `make lint` - `golangci-lint run`

## Conventions

- Flat package layout: every file is `package main` in the repository root, no subdirectories.
- Alphabetical ordering of types, constants, variables, and functions.
- Errors wrapped with context; no silently ignored errors.
- No magic numbers; named constants instead.
