package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab int

const (
	tabDrivers tab = iota
	tabConstructors
	tabRaces
	tabProgression
)

type progMode int

const (
	modeDrivers progMode = iota
	modeConstructors
)

const chromeHeight = 6

const raceRoundWidth = 5

var tabTitles = []string{"Drivers", "Constructors", "Races", "Progression"}

type appModel struct {
	active           tab
	constructorsView viewport.Model
	data             *data
	driversView      viewport.Model
	height           int
	progMode         progMode
	progOffset       int
	progView         viewport.Model
	racesTable       table.Model
	raceSelected     int
	ready            bool
	resultsView      viewport.Model
	width            int
}

func newAppModel(model *data) appModel {
	application := appModel{
		active:           tabDrivers,
		constructorsView: viewport.New(0, 0),
		data:             model,
		driversView:      viewport.New(0, 0),
		progMode:         modeDrivers,
		progView:         viewport.New(0, 0),
		racesTable:       newRacesTable(model),
		raceSelected:     -1,
		resultsView:      viewport.New(0, 0),
	}
	return application
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m appModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "tab":
		m.active = (m.active + 1) % tab(len(tabTitles))
		m.syncFocus()
		return m, nil

	case "shift+tab":
		m.active = (m.active + tab(len(tabTitles)) - 1) % tab(len(tabTitles))
		m.syncFocus()
		return m, nil
	}

	switch m.active {
	case tabDrivers:
		updated, cmd := m.driversView.Update(msg)
		m.driversView = updated
		return m, cmd

	case tabConstructors:
		updated, cmd := m.constructorsView.Update(msg)
		m.constructorsView = updated
		return m, cmd

	case tabRaces:
		return m.handleRacesKey(msg)

	case tabProgression:
		return m.handleProgressionKey(msg)
	}

	return m, nil
}

func (m appModel) handleRacesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.raceSelected >= 0 {
		if msg.String() == "esc" {
			m.raceSelected = -1
			return m, nil
		}

		updated, cmd := m.resultsView.Update(msg)
		m.resultsView = updated
		return m, cmd
	}

	if msg.String() == "enter" {
		index := m.racesTable.Cursor()
		if index >= 0 && index < len(m.data.races) {
			m.raceSelected = index
			m.resultsView.SetContent(renderResultsTable(m.data.races[index]))
			m.resultsView.GotoTop()
		}
		return m, nil
	}

	updated, cmd := m.racesTable.Update(msg)
	m.racesTable = updated
	return m, cmd
}

func (m appModel) handleProgressionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d":
		m.progMode = modeDrivers
		m.progOffset = 0
		m.refreshProgression()
		return m, nil

	case "c":
		m.progMode = modeConstructors
		m.progOffset = 0
		m.refreshProgression()
		return m, nil

	case "left":
		if m.progOffset > 0 {
			m.progOffset--
			m.refreshProgression()
		}
		return m, nil

	case "right":
		if m.progOffset < m.maxProgOffset() {
			m.progOffset++
			m.refreshProgression()
		}
		return m, nil
	}

	updated, cmd := m.progView.Update(msg)
	m.progView = updated
	return m, cmd
}

func (m appModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := fmt.Sprintf("F1 %s Season", m.data.season)
	sections := []string{
		titleStyle.Render(header),
		m.tabBar(),
		m.body(),
		helpStyle.Render(m.help()),
	}
	return strings.Join(sections, "\n")
}

func (m appModel) tabBar() string {
	rendered := make([]string, len(tabTitles))
	for index, title := range tabTitles {
		if tab(index) == m.active {
			rendered[index] = activeTabStyle.Render(title)
			continue
		}
		rendered[index] = inactiveTabStyle.Render(title)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m appModel) body() string {
	switch m.active {
	case tabDrivers:
		return m.driversView.View()

	case tabConstructors:
		return m.constructorsView.View()

	case tabRaces:
		if m.raceSelected >= 0 {
			heading := titleStyle.Render(m.data.races[m.raceSelected].name)
			return fmt.Sprintf("%s\n%s", heading, m.resultsView.View())
		}
		return m.racesTable.View()

	case tabProgression:
		return fmt.Sprintf("%s\n%s", matrixLegend(), m.progView.View())
	}

	return ""
}

func (m appModel) help() string {
	switch m.active {
	case tabRaces:
		if m.raceSelected >= 0 {
			return "esc back | up/down scroll | tab switch | q quit"
		}
		return "enter results | up/down move | tab switch | q quit"

	case tabProgression:
		return "d drivers | c constructors | left/right scroll | up/down scroll | tab switch | q quit"
	}

	return "up/down scroll | tab/shift+tab switch | q quit"
}

func (m *appModel) resize() {
	height := m.tableHeight()

	m.driversView.Width = m.width
	m.driversView.Height = height
	m.driversView.SetContent(renderDriversTable(m.data))

	m.constructorsView.Width = m.width
	m.constructorsView.Height = height
	m.constructorsView.SetContent(renderConstructorsTable(m.data))

	m.progView.Width = m.width
	m.progView.Height = height - 1
	m.refreshProgression()

	m.resultsView.Width = m.width
	m.resultsView.Height = height - 1
	if m.raceSelected >= 0 {
		m.resultsView.SetContent(renderResultsTable(m.data.races[m.raceSelected]))
	}

	m.racesTable.SetHeight(height)
}

func (m *appModel) refreshProgression() {
	series := m.progressionSeries()
	rounds := m.data.progression.rounds - m.progOffset
	visible := visibleColumns(m.width, progLabelWidth(series))
	if visible > rounds {
		visible = rounds
	}

	title := "Driver"
	if m.progMode == modeConstructors {
		title = "Constructor"
	}

	m.progView.SetContent(renderProgressionTable(series, m.data.progression.raceLabels, title, m.progOffset, visible))
}

func (m appModel) maxProgOffset() int {
	visible := visibleColumns(m.width, progLabelWidth(m.progressionSeries()))
	maximum := m.data.progression.rounds - visible
	if maximum < 0 {
		return 0
	}
	return maximum
}

func (m appModel) progressionSeries() []seriesRow {
	if m.progMode == modeConstructors {
		return m.data.progression.constructors
	}
	return m.data.progression.drivers
}

func (m appModel) tableHeight() int {
	height := m.height - chromeHeight
	if height < 3 {
		return 3
	}
	return height
}

func (m *appModel) syncFocus() {
	m.racesTable.Blur()
	if m.active == tabRaces {
		m.racesTable.Focus()
	}
}

func progLabelWidth(series []seriesRow) int {
	width := len("Driver")
	for _, row := range series {
		if len(row.label) > width {
			width = len(row.label)
		}
	}
	return width
}

func newRacesTable(model *data) table.Model {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		Bold(true).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("252"))
	styles.Selected = styles.Selected.
		Background(lipgloss.Color("63")).
		Bold(true).
		Foreground(lipgloss.Color("230"))

	headers := []string{"Round", "Grand Prix", "Winner", "Constructor", "Pole", "Fastest Lap"}

	rows := make([]table.Row, 0, len(model.races))
	for _, race := range model.races {
		rows = append(rows, table.Row{
			fmt.Sprintf("%*d", raceRoundWidth, race.round),
			race.name,
			race.winner,
			race.winnerTeam,
			race.pole,
			race.fastestLap,
		})
	}

	// Size each column to its widest value so nothing is truncated with an
	// ellipsis; bubbles tables use fixed column widths.
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = lipgloss.Width(header)
	}
	for _, row := range rows {
		for index, cell := range row {
			width := lipgloss.Width(cell)
			if width > widths[index] {
				widths[index] = width
			}
		}
	}

	columns := make([]table.Column, len(headers))
	for index, header := range headers {
		columns[index] = table.Column{Title: header, Width: widths[index]}
	}

	built := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithStyles(styles),
		table.WithFocused(true),
	)
	return built
}
