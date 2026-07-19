package main

import "github.com/charmbracelet/lipgloss"

var (
	activeTabStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Padding(0, 2)

	chartStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("81"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

func categoryStyle(category resultCategory) lipgloss.Style {
	base := lipgloss.NewStyle().Align(lipgloss.Right)

	switch category {
	case categoryWin:
		return base.Background(lipgloss.Color("220")).Foreground(lipgloss.Color("16"))
	case categorySecond:
		return base.Background(lipgloss.Color("250")).Foreground(lipgloss.Color("16"))
	case categoryThird:
		return base.Background(lipgloss.Color("173")).Foreground(lipgloss.Color("16"))
	case categoryPoints:
		return base.Background(lipgloss.Color("34")).Foreground(lipgloss.Color("16"))
	case categoryFinished:
		return base.Background(lipgloss.Color("39")).Foreground(lipgloss.Color("16"))
	case categoryRetired:
		return base.Background(lipgloss.Color("97")).Foreground(lipgloss.Color("230"))
	default:
		return base.Foreground(lipgloss.Color("245"))
	}
}
