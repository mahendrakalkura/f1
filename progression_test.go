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
