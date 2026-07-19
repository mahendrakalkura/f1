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
	tabSprints
	tabDrvProg
	tabConProg
)

type progMode int

const (
	modeDrivers progMode = iota
	modeConstructors
)

const chromeHeight = 6

const raceRoundWidth = 5

var tabTitles = []string{"Drivers", "Constructors", "Races", "Sprints", "Drivers Progression", "Constructors Progression"}

type appModel struct {
	active           tab
	constructorsView viewport.Model
	data             *data
	driversView      viewport.Model
	height           int
	progChart        bool
	progMode         progMode
	progOffset       int
	progView         viewport.Model
	races            raceList
	ready            bool
	sprints          raceList
	width            int
}

// raceList bundles the shared pieces of the Races and Sprints tabs: a table
// of rounds plus a viewport showing one round's full results.
type raceList struct {
	races    []raceInfo
	selected int
	table    table.Model
	view     viewport.Model
}

func newAppModel(model *data) appModel {
	application := appModel{
		active:           tabDrivers,
		constructorsView: viewport.New(0, 0),
		data:             model,
		driversView:      viewport.New(0, 0),
		progMode:         modeDrivers,
		progView:         viewport.New(0, 0),
		races:            newRaceList(model.races, newRacesTable(model.races)),
		sprints:          newRaceList(model.sprints, newSprintsTable(model.sprints)),
	}
	return application
}

func newRaceList(races []raceInfo, table table.Model) raceList {
	return raceList{
		races:    races,
		selected: -1,
		table:    table,
		view:     viewport.New(0, 0),
	}
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
		m.syncProgMode()
		m.syncFocus()
		return m, nil

	case "shift+tab":
		m.active = (m.active + tab(len(tabTitles)) - 1) % tab(len(tabTitles))
		m.syncProgMode()
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
		list, cmd := handleRaceListKey(m.races, msg)
		m.races = list
		return m, cmd

	case tabSprints:
		list, cmd := handleRaceListKey(m.sprints, msg)
		m.sprints = list
		return m, cmd

	case tabDrvProg, tabConProg:
		return m.handleProgressionKey(msg)
	}

	return m, nil
}

func handleRaceListKey(list raceList, msg tea.KeyMsg) (raceList, tea.Cmd) {
	if list.selected >= 0 {
		if msg.String() == "esc" {
			list.selected = -1
			return list, nil
		}

		updated, cmd := list.view.Update(msg)
		list.view = updated
		return list, cmd
	}

	if msg.String() == "enter" {
		index := list.table.Cursor()
		if index >= 0 && index < len(list.races) {
			list.selected = index
			list.view.SetContent(renderResultsTable(list.races[index]))
			list.view.GotoTop()
		}
		return list, nil
	}

	updated, cmd := list.table.Update(msg)
	list.table = updated
	return list, cmd
}

func (m appModel) handleProgressionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p":
		m.progChart = !m.progChart
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
		return m.races.body()

	case tabSprints:
		return m.sprints.body()

	case tabDrvProg, tabConProg:
		return fmt.Sprintf("%s\n%s", m.progressionHeader(), m.progView.View())
	}

	return ""
}

func (l raceList) body() string {
	if l.selected < 0 {
		return l.table.View()
	}

	race := l.races[l.selected]
	heading := titleStyle.Render(race.name)
	meta := helpStyle.Render(raceMeta(race))
	return fmt.Sprintf("%s\n%s\n%s", heading, meta, l.view.View())
}

func raceMeta(race raceInfo) string {
	parts := []string{}
	if race.date != "" {
		parts = append(parts, race.date)
	}
	if race.circuit != "" {
		parts = append(parts, race.circuit)
	}
	if race.location != "" {
		parts = append(parts, race.location)
	}
	return strings.Join(parts, " | ")
}

func (m appModel) progressionHeader() string {
	if m.progChart {
		if m.progMode == modeConstructors {
			return helpStyle.Render("Championship position per round (1-9, A=10 B=11 …)")
		}
		return helpStyle.Render("Cumulative points per round, scaled to the leader's total")
	}
	switch m.progMode {
	case modeConstructors:
		return helpStyle.Render("Cumulative points per round")
	}
	return matrixLegend()
}

func (m appModel) help() string {
	switch m.active {
	case tabRaces:
		return raceListHelp(m.races.selected)

	case tabSprints:
		return raceListHelp(m.sprints.selected)

	case tabDrvProg, tabConProg:
		return "p points chart | left/right scroll | up/down scroll | tab switch | q quit"
	}

	return "up/down scroll | tab/shift+tab switch | q quit"
}

func raceListHelp(selected int) string {
	if selected >= 0 {
		return "esc back | up/down scroll | tab switch | q quit"
	}
	return "enter results | up/down move | tab switch | q quit"
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

	m.resizeRaceList(&m.races)
	m.resizeRaceList(&m.sprints)
}

func (m *appModel) resizeRaceList(list *raceList) {
	list.view.Width = m.width
	list.view.Height = m.tableHeight() - 2
	if list.selected >= 0 {
		list.view.SetContent(renderResultsTable(list.races[list.selected]))
	}
	list.table.SetHeight(m.tableHeight())
}

func (m *appModel) refreshProgression() {
	if m.progChart {
		if m.progMode == modeConstructors {
			m.progView.SetContent(renderPositionChart(m.data.progression.chartConstructorPositions, m.width))
			return
		}
		m.progView.SetContent(renderChart(m.data.progression.chart, m.width))
		return
	}

	series := m.progressionSeries()
	rounds := m.data.progression.rounds - m.progOffset
	visible := visibleColumns(m.width, progLabelsWidth(series, m.progMode))
	if visible > rounds {
		visible = rounds
	}

	m.progView.SetContent(renderProgressionTable(series, m.data.progression.raceLabels, m.progMode, m.progOffset, visible))
}

func (m appModel) maxProgOffset() int {
	if m.progChart {
		return 0
	}

	visible := visibleColumns(m.width, progLabelsWidth(m.progressionSeries(), m.progMode))
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
	m.races.table.Blur()
	m.sprints.table.Blur()
	if m.active == tabRaces {
		m.races.table.Focus()
	}
	if m.active == tabSprints {
		m.sprints.table.Focus()
	}
}

// syncProgMode aligns the progression subject with the active tab and
// refreshes the viewport so the correct matrix or chart renders immediately.
func (m *appModel) syncProgMode() {
	m.progChart = false
	m.progOffset = 0
	if m.active == tabConProg {
		m.progMode = modeConstructors
	} else {
		m.progMode = modeDrivers
	}
	m.refreshProgression()
}

// progLabelsWidth returns the combined width of the name column and the
// context column (team for drivers, stacked driver names for constructors).
func progLabelsWidth(series []seriesRow, mode progMode) int {
	nameWidth := len("Driver")
	detailWidth := len("Team")
	if mode == modeConstructors {
		nameWidth = len("Constructor")
		detailWidth = len("Drivers")
	}

	for _, row := range series {
		nameWidth = max(nameWidth, len(row.label))
		if mode == modeConstructors {
			for _, driver := range row.drivers {
				detailWidth = max(detailWidth, len(driver))
			}
			continue
		}
		detailWidth = max(detailWidth, len(row.team))
	}
	return nameWidth + detailWidth
}

func newRacesTable(races []raceInfo) table.Model {
	headers := []string{"Round", "Grand Prix", "Winner", "Constructor", "Pole", "Fastest Lap"}
	rows := make([]table.Row, 0, len(races))
	for _, race := range races {
		rows = append(rows, table.Row{
			fmt.Sprintf("%*d", raceRoundWidth, race.round),
			race.name,
			race.winner,
			race.winnerTeam,
			race.pole,
			race.fastestLap,
		})
	}
	return racesTableModel(headers, rows)
}

func newSprintsTable(races []raceInfo) table.Model {
	headers := []string{"Round", "Grand Prix", "Winner", "Constructor", "Fastest Lap"}
	rows := make([]table.Row, 0, len(races))
	for _, race := range races {
		rows = append(rows, table.Row{
			fmt.Sprintf("%*d", raceRoundWidth, race.round),
			race.name,
			race.winner,
			race.winnerTeam,
			race.fastestLap,
		})
	}
	return racesTableModel(headers, rows)
}

// racesTableModel sizes each column to its widest value so nothing is
// truncated with an ellipsis; bubbles tables use fixed column widths.
func racesTableModel(headers []string, rows []table.Row) table.Model {
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
