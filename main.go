package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	force := flag.Bool("force", false, "force refresh cached data")
	year := flag.Int("season", 0, "season year (default: current)")
	flag.Parse()

	store, err := newCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "f1: %v\n", err)
		os.Exit(1)
	}

	season := seasonSlug(*year)
	model, err := loadData(store, *force, season)
	if err != nil {
		fmt.Fprintf(os.Stderr, "f1: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(newAppModel(model), tea.WithAltScreen())
	_, err = program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "f1: %v\n", err)
		os.Exit(1)
	}
}
