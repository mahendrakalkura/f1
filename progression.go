package main

import (
	"fmt"
	"strings"
)

const roundColumnWidth = 6

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
