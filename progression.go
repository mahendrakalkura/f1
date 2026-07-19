package main

import (
	"fmt"
	"math"
	"strings"
)

const roundColumnWidth = 6

// chartRamp maps a value's share of the maximum to an ASCII intensity
// character, from empty (zero) to full (the leader's total).
const chartRamp = " .:-=+*#%@"

func matrixLegend() string {
	entries := []struct {
		category resultCategory
		label    string
	}{
		{categoryWin, "Win"},
		{categorySecond, "2nd"},
		{categoryThird, "3rd"},
		{categoryPoints, "Points"},
		{categoryFinished, "No points"},
		{categoryRetired, "Ret/DNF"},
	}

	parts := []string{}
	for _, entry := range entries {
		swatch := categoryStyle(entry.category).Render("  ")
		parts = append(parts, fmt.Sprintf("%s %s", swatch, entry.label))
	}
	return strings.Join(parts, "  ")
}

// renderChart draws one cumulative-points sparkline per driver, scaled so the
// season leader's total fills the ramp.
func renderChart(rows []chartRow, width int) string {
	if len(rows) == 0 || len(rows[0].points) == 0 {
		return "No completed rounds yet."
	}

	labelWidth := 0
	maximum := 0.0
	for _, row := range rows {
		if len(row.label) > labelWidth {
			labelWidth = len(row.label)
		}
		if total := row.points[len(row.points)-1]; total > maximum {
			maximum = total
		}
	}

	totalWidth := len(formatPoints(maximum))
	chartWidth := max(width-labelWidth-totalWidth-2, 1)

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		total := formatPoints(row.points[len(row.points)-1])
		line := fmt.Sprintf(
			"%-*s %s %*s",
			labelWidth,
			row.label,
			chartStyle.Render(sparkline(row.points, chartWidth, maximum)),
			totalWidth,
			total,
		)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// sparkline renders one character per sampled round, with the character's
// intensity proportional to the cumulative value against the season maximum.
func sparkline(points []float64, width int, maximum float64) string {
	if width < 1 || len(points) == 0 {
		return ""
	}
	if width > len(points) {
		width = len(points)
	}

	var line strings.Builder
	for column := 0; column < width; column++ {
		index := column * len(points) / width
		level := 0
		if maximum > 0 {
			level = int(math.Round(points[index] / maximum * float64(len(chartRamp)-1)))
		}
		line.WriteByte(chartRamp[level])
	}
	return line.String()
}

// posChar maps a championship position (1-based) to a single display
// character. Positions 1–9 use digits; 10–15 use letters A–F.
func posChar(position int) byte {
	if position < 10 {
		return byte('0' + position)
	}
	if position < 16 {
		return byte('A' + position - 10)
	}
	return '?'
}

// renderPositionChart draws each constructor's championship position after
// every round as a compact digit track, downsampled to fit the terminal width.
func renderPositionChart(rows []chartRow, width int) string {
	if len(rows) == 0 || len(rows[0].points) == 0 {
		return "No completed rounds yet."
	}

	labelWidth := 0
	maximum := 0.0
	for _, row := range rows {
		labelWidth = max(labelWidth, len(row.label))
		if total := row.points[len(row.points)-1]; total > maximum {
			maximum = total
		}
	}

	totalWidth := len(formatPoints(maximum))
	trackWidth := max(width-labelWidth-totalWidth-2, 1)

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		last := int(math.Round(row.points[len(row.points)-1]))
		total := fmt.Sprintf("P%d", last)
		line := fmt.Sprintf(
			"%-*s %s %*s",
			labelWidth,
			row.label,
			chartStyle.Render(positionTrack(row.points, trackWidth)),
			totalWidth,
			total,
		)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// positionTrack renders a downsampled sequence of position characters, one
// per sampled round.
func positionTrack(positions []float64, width int) string {
	if width < 1 || len(positions) == 0 {
		return ""
	}
	if width > len(positions) {
		width = len(positions)
	}

	var line strings.Builder
	for column := 0; column < width; column++ {
		index := column * len(positions) / width
		line.WriteByte(posChar(int(math.Round(positions[index]))))
	}
	return line.String()
}

// visibleColumns estimates how many round columns fit the terminal width once
// the label column and the four right-hand summary columns are reserved.
func visibleColumns(width, labelWidth int) int {
	reserved := labelWidth + 4 + 4*roundColumnWidth
	usable := width - reserved
	if usable < roundColumnWidth {
		return 1
	}
	return usable / roundColumnWidth
}
