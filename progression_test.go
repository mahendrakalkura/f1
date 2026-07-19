package main

import (
	"strings"
	"testing"
)

func TestRenderChart(t *testing.T) {
	rows := []chartRow{
		{label: "Max Verstappen", points: []float64{33, 51}},
		{label: "Lando Norris", points: []float64{18, 43}},
	}

	out := renderChart(rows, 40)
	if !strings.Contains(out, "Max Verstappen") || !strings.Contains(out, "Lando Norris") {
		t.Errorf("output missing driver names:\n%s", out)
	}
	if !strings.Contains(out, "51") || !strings.Contains(out, "43") {
		t.Errorf("output missing totals:\n%s", out)
	}
}

func TestRenderChartEmpty(t *testing.T) {
	if got := renderChart(nil, 40); got != "No completed rounds yet." {
		t.Errorf("got %q, want %q", got, "No completed rounds yet.")
	}

	rows := []chartRow{{label: "Max Verstappen"}}
	if got := renderChart(rows, 40); got != "No completed rounds yet." {
		t.Errorf("got %q, want %q", got, "No completed rounds yet.")
	}
}

func TestRenderChartNarrow(t *testing.T) {
	rows := []chartRow{{label: "Max Verstappen", points: []float64{33, 51}}}
	if out := renderChart(rows, 10); !strings.Contains(out, "Max Verstappen") {
		t.Errorf("narrow output missing label:\n%s", out)
	}
}

func TestSparkline(t *testing.T) {
	cases := []struct {
		name    string
		points  []float64
		width   int
		maximum float64
		want    string
	}{
		{"empty", nil, 5, 100, ""},
		{"zero width", []float64{1}, 0, 100, ""},
		{"zero maximum", []float64{0, 0}, 2, 0, "  "},
		{"full ramp", []float64{10, 20}, 2, 20, "+@"},
		{"downsampled", []float64{0, 10, 20, 30}, 2, 30, " *"},
		{"width capped at points", []float64{50}, 10, 50, "@"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sparkline(c.points, c.width, c.maximum); got != c.want {
				t.Errorf("sparkline(%v, %d, %v) = %q, want %q", c.points, c.width, c.maximum, got, c.want)
			}
		})
	}
}

func TestPosChar(t *testing.T) {
	cases := []struct {
		position int
		want     byte
	}{
		{1, '1'}, {9, '9'}, {10, 'A'}, {11, 'B'}, {15, 'F'}, {16, '?'},
	}
	for _, c := range cases {
		if got := posChar(c.position); got != c.want {
			t.Errorf("posChar(%d) = %c, want %c", c.position, got, c.want)
		}
	}
}

func TestPositionTrack(t *testing.T) {
	got := positionTrack([]float64{1, 2, 3, 10, 11}, 5)
	if got != "123AB" {
		t.Errorf("got %q, want %q", got, "123AB")
	}
	got = positionTrack([]float64{1, 3, 10}, 2)
	if got != "13" {
		t.Errorf("downsampled: got %q, want %q", got, "13")
	}
	if got := positionTrack(nil, 5); got != "" {
		t.Errorf("empty input: got %q, want empty", got)
	}
}

func TestRenderPositionChart(t *testing.T) {
	rows := []chartRow{
		{label: "Mercedes", points: []float64{1, 1, 2, 2, 3}},
		{label: "Ferrari", points: []float64{2, 3, 1, 1, 1}},
	}

	out := renderPositionChart(rows, 40)
	if !strings.Contains(out, "Mercedes") || !strings.Contains(out, "Ferrari") {
		t.Errorf("missing labels:\n%s", out)
	}
	if !strings.Contains(out, "P3") || !strings.Contains(out, "P1") {
		t.Errorf("missing final position:\n%s", out)
	}
}

func TestRenderPositionChartEmpty(t *testing.T) {
	if got := renderPositionChart(nil, 40); got != "No completed rounds yet." {
		t.Errorf("got %q, want 'No completed rounds yet.'", got)
	}
}

func TestVisibleColumns(t *testing.T) {
	cases := []struct {
		width      int
		labelWidth int
		want       int
	}{
		{80, 10, 7},
		{44, 10, 1},
		{20, 10, 1},
	}
	for _, c := range cases {
		if got := visibleColumns(c.width, c.labelWidth); got != c.want {
			t.Errorf("visibleColumns(%d, %d) = %d, want %d", c.width, c.labelWidth, got, c.want)
		}
	}
}
