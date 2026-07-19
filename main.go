package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	force := flag.Bool("force", false, "force refresh cached data")
	flag.Parse()

	store, err := newCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "f1: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "f1: loading season data...")
	model, err := loadData(store, *force)
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
