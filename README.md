# f1

Terminal UI for the current Formula 1 season: driver standings, constructor standings, every race result, and a Wikipedia-style points progression grid. Data comes from the Jolpica F1 API (the Ergast successor) and is cached locally for 24 hours.

## Features

- Driver standings with team, points, wins, poles, and fastest laps; each driver's team is shown inline.
- Constructor standings with points, wins, poles, and fastest laps; each team's full driver names are listed one below the other in current-ranking order.
- Races list showing the winning driver and constructor, the pole sitter, and the fastest-lap driver for every round; press enter for the full finishing order.
- Progression tab: a Wikipedia-style matrix of colour-coded finishing positions per round (drivers) or cumulative points per round (constructors), with running points, wins, poles, and fastest-lap totals. Toggle drivers or constructors and scroll the round columns.
- Tables are rendered with the `lipgloss/table` library; all numeric columns are right aligned.
- Parallel, cached downloads. The first run fetches every endpoint concurrently through a bounded worker pool; later runs read from the 24-hour cache and start instantly.

Wins, poles, and fastest laps are aggregated across the season: wins from the standings, poles from qualifying (position one), and fastest laps from each race's fastest-lap rank.

## Requirements

- Go 1.26 or newer to build.
- `golangci-lint` for `make lint`.
- Network access on the first run (or after the cache expires).

## Usage

```sh
make build      # compile to ./main
make run        # run, using cached data when fresh
make run force  # run and re-download, ignoring the cache
make lint       # golangci-lint run
make clean      # remove the ./main binary
```

Running the binary directly:

```sh
./main          # uses cache when present, fetches only when empty or stale
./main --force  # force a fresh download
```

## Keys

```
+-------------+------------------------------------------+
| Key         | Action                                   |
+-------------+------------------------------------------+
| tab         | Next tab                                 |
| shift+tab   | Previous tab                             |
| up / down   | Move within the Races list, or scroll    |
| enter       | On Races, show that race's full result   |
| esc         | Back from a race result to the race list |
| d / c       | Progression: drivers / constructors      |
| left/right  | Progression: scroll round columns        |
| q / ctrl+c  | Quit                                     |
+-------------+------------------------------------------+
```

## Caching

Responses are stored under `$XDG_CACHE_HOME/f1` (typically `~/.cache/f1`), one file per request keyed by the URL hash. A cached file younger than 24 hours is served directly; otherwise it is re-fetched. Rate-limited (HTTP 429) and transient 5xx responses are retried with exponential backoff that honours any `Retry-After` header.

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
| progression.go | Progression legend and column-fit helper       |
| styles.go      | Lipgloss styles and result-category colours    |
| tables.go      | lipgloss/table builders for every screen       |
| tui.go         | Bubble Tea model, tabs, and viewports          |
+----------------+------------------------------------------------+
```
