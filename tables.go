package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	ltable "github.com/charmbracelet/lipgloss/table"
)

// numericColumns lists, per table, the column indices that hold numbers and
// must be right aligned.
func alignFunc(numeric map[int]bool) ltable.StyleFunc {
	return func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if row == ltable.HeaderRow {
			style = style.Bold(true).Foreground(lipgloss.Color("252"))
		}
		if numeric[col] {
			style = style.Align(lipgloss.Right)
		}
		return style
	}
}

func baseTable() *ltable.Table {
	return ltable.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		BorderRow(false)
}

func renderDriversTable(model *data) string {
	rows := make([][]string, 0, len(model.drivers))
	for _, row := range model.drivers {
		rows = append(rows, []string{
			fmt.Sprintf("%d", row.position),
			row.name,
			row.team,
			formatPoints(row.points),
			fmt.Sprintf("%d", row.wins),
			fmt.Sprintf("%d", row.poles),
			fmt.Sprintf("%d", row.fastestLaps),
		})
	}

	numeric := map[int]bool{0: true, 3: true, 4: true, 5: true, 6: true}
	table := baseTable().
		Headers("Position", "Driver", "Team", "Points", "Wins", "Poles", "FL").
		Rows(rows...).
		StyleFunc(alignFunc(numeric))
	return table.Render()
}

func renderConstructorsTable(model *data) string {
	rows := make([][]string, 0, len(model.constructors))
	for _, row := range model.constructors {
		rows = append(rows, []string{
			fmt.Sprintf("%d", row.position),
			row.name,
			strings.Join(row.drivers, "\n"),
			formatPoints(row.points),
			fmt.Sprintf("%d", row.wins),
			fmt.Sprintf("%d", row.poles),
			fmt.Sprintf("%d", row.fastestLaps),
		})
	}

	numeric := map[int]bool{0: true, 3: true, 4: true, 5: true, 6: true}
	table := baseTable().
		Headers("Position", "Constructor", "Drivers", "Points", "Wins", "Poles", "FL").
		Rows(rows...).
		StyleFunc(alignFunc(numeric))
	return table.Render()
}

func renderResultsTable(race raceInfo) string {
	rows := make([][]string, 0, len(race.results))
	for _, result := range race.results {
		rows = append(rows, []string{
			fmt.Sprintf("%d", result.position),
			result.driver,
			result.team,
			fmt.Sprintf("%d", result.grid),
			formatPoints(result.points),
			result.status,
		})
	}

	numeric := map[int]bool{0: true, 3: true, 4: true}
	table := baseTable().
		Headers("Position", "Driver", "Team", "Grid", "Points", "Status").
		Rows(rows...).
		StyleFunc(alignFunc(numeric))
	return table.Render()
}

// renderProgressionTable draws the Wikipedia-style grid for the visible round
// window using colored cells for finishing positions and right-aligned totals.
func renderProgressionTable(series []seriesRow, labels []string, title string, colOffset, roundCols int) string {
	if len(series) == 0 || len(labels) == 0 {
		return "No completed rounds yet."
	}

	headers := []string{title}
	for column := colOffset; column < colOffset+roundCols; column++ {
		headers = append(headers, labels[column])
	}
	headers = append(headers, "Pts", "W", "Pol", "FL")

	rows := make([][]string, 0, len(series))
	for _, row := range series {
		cells := []string{row.label}
		for column := colOffset; column < colOffset+roundCols; column++ {
			cells = append(cells, row.cells[column].text)
		}
		cells = append(cells,
			formatPoints(row.total),
			fmt.Sprintf("%d", row.wins),
			fmt.Sprintf("%d", row.poles),
			fmt.Sprintf("%d", row.fastestLaps),
		)
		rows = append(rows, cells)
	}

	table := baseTable().
		BorderColumn(true).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(progressionStyle(series, colOffset, roundCols))
	return table.Render()
}

func progressionStyle(series []seriesRow, colOffset, roundCols int) ltable.StyleFunc {
	return func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)

		if row == ltable.HeaderRow {
			return style.Bold(true).Align(lipgloss.Right).Foreground(lipgloss.Color("252"))
		}

		if col == 0 {
			return style.Foreground(lipgloss.Color("252"))
		}

		if col >= 1 && col <= roundCols {
			cell := series[row].cells[colOffset+col-1]
			return categoryStyle(cell.category).Padding(0, 1)
		}

		return style.Align(lipgloss.Right)
	}
}
