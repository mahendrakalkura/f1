# f1

Terminal UI for the Formula 1 season: driver standings, constructor standings, every race and sprint result, and a Wikipedia-style points progression grid. Data comes from the Jolpica F1 API (the Ergast successor) and is cached locally for 24 hours. The current season is shown by default; any season back to 1950 can be selected with `--season`.

## Features

- Driver standings with team, points, wins, poles, and fastest laps; each driver's team is shown inline.
- Constructor standings with points, wins, poles, and fastest laps; each team's full driver names are listed one below the other in current-ranking order.
- Races list showing the winning driver and constructor, the pole sitter, and the fastest-lap driver for every round; press enter for the full finishing order plus the date, circuit, and location.
- Sprints list with the winning driver and constructor and the fastest-lap driver for every sprint round; press enter for the full sprint result.
- Progression tab: a Wikipedia-style matrix of colour-coded finishing positions per round (drivers) or cumulative points per round (constructors), with running points, wins, poles, and fastest-lap totals. Drivers mode shows each driver's team; constructors mode lists each team's drivers. Press p for a per-driver cumulative points chart. Scroll the round columns with left and right.
- Tables are rendered with the `lipgloss/table` library; all numeric columns are right aligned.
- Parallel, cached downloads. The first run fetches every endpoint concurrently through a bounded worker pool; later runs read from the 24-hour cache and start instantly. When a download fails but a stale cache file exists, the stale file is served with a warning instead of aborting.

Wins, poles, and fastest laps are aggregated across the season: wins from the standings, poles from qualifying (position one), and fastest laps from each race's fastest-lap rank.

## Requirements

- Go 1.26 or newer to build.
- `goimports` on `PATH` (`make build` runs it before compiling).
- `golangci-lint` for `make lint`.
- Network access on the first run (or after the cache expires, though stale data is served as a fallback).

## Usage

```sh
make build      # goimports, go mod tidy, compile to ./main
make run        # build then run, using cached data when fresh
make run force  # build then run and re-download, ignoring the cache
make test       # go test ./...
make lint       # golangci-lint run
```

Running the binary directly:

```sh
./main                  # uses cache when present, fetches only when empty or stale
./main --force          # force a fresh download
./main --season 2024    # show a past season instead of the current one
```

## Keys

```
+------------------+------------------------------------------------+
| Key              | Action                                         |
+------------------+------------------------------------------------+
| tab              | Next tab                                       |
| shift+tab        | Previous tab                                   |
| up / down        | Move within the Races list, or scroll          |
| enter            | On Races or Sprints, show that round's result  |
| esc              | Back from a result to the list                 |
| p                | Progression: points chart toggle               |
| left/right       | Progression: scroll round columns              |
| q / ctrl+c       | Quit                                           |
+------------------+------------------------------------------------+
```

## Caching

Responses are stored under `./.cache` in the project directory (git-ignored), one file per request keyed by the URL hash. A cached file younger than 24 hours is served directly; otherwise it is re-fetched. When re-fetching fails, the stale file is served with a warning. Files older than 30 days are deleted in the background on every run. Rate-limited (HTTP 429) and transient 5xx responses are retried with exponential backoff that honours any `Retry-After` header.

## Files

```
+----------------+------------------------------------------------+
| File           | Purpose                                        |
+----------------+------------------------------------------------+
| api.go         | Jolpica JSON types and endpoint URL builders   |
| cache.go       | 24-hour file cache and HTTP download + retry   |
| fetch.go       | Parallel fetch orchestration (errgroup)        |
| main.go        | Entry point and flag parsing                   |
| model.go       | Domain model, cross references, season totals  |
| progression.go | Legend, points chart, and column-fit helper    |
| styles.go      | Lipgloss styles and result-category colours    |
| tables.go      | lipgloss/table builders for every screen       |
| tui.go         | Bubble Tea model, tabs, and viewports          |
| *_test.go      | Unit tests for every layer                     |
+----------------+------------------------------------------------+
```
