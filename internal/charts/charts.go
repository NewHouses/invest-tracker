// Package charts renders ASCII bar and line charts for the Ver gráficas
// submenu. Bar charts are drawn in-house; line charts wrap asciigraph.
package charts

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/guptarohit/asciigraph"

	"invest-tracker/internal/colors"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
)

// fgPalette ANSI 16-color foreground codes used to colorise bars / lines so
// that adjacent series are visually distinguishable.
var fgPalette = []string{
	"\x1b[31m", // red
	"\x1b[32m", // green
	"\x1b[33m", // yellow
	"\x1b[34m", // blue
	"\x1b[35m", // magenta
	"\x1b[36m", // cyan
}

var lineColors = []asciigraph.AnsiColor{
	asciigraph.Red,
	asciigraph.Green,
	asciigraph.Yellow,
	asciigraph.Blue,
	asciigraph.Magenta,
	asciigraph.Cyan,
}

// PaletteFG returns the i-th foreground color in the palette (cyclical).
func PaletteFG(i int) string { return fgPalette[i%len(fgPalette)] }

// PaletteLine returns the i-th asciigraph color for line charts (cyclical).
func PaletteLine(i int) asciigraph.AnsiColor { return lineColors[i%len(lineColors)] }

// BarItem is one row in a horizontal bar chart.
type BarItem struct {
	Label string
	Value float64
	Color string // ANSI escape; empty means no color
}

// RenderBars writes a horizontal bar chart sorted by Value descending. total
// is the basis for percentage calculation; if 0 the function returns silently.
func RenderBars(w io.Writer, items []BarItem, total float64) {
	if len(items) == 0 || total <= 0 {
		return
	}
	sorted := make([]BarItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	maxLabel := 0
	for _, it := range sorted {
		if l := utf8.RuneCountInString(it.Label); l > maxLabel {
			maxLabel = l
		}
	}

	const barWidth = 40
	for _, it := range sorted {
		pct := it.Value / total * 100
		n := int(it.Value/total*float64(barWidth) + 0.5)
		if n < 1 && it.Value > 0 {
			n = 1
		}
		bar := strings.Repeat("█", n)
		if it.Color != "" {
			bar = it.Color + bar + colors.Reset
		}
		fmt.Fprintf(w, "  %s  %5.1f%%  %s  %s\n",
			padRight(it.Label, maxLabel), pct, bar, money.USD(it.Value))
	}
}

// padRight pads s with spaces to the given visual rune count.
func padRight(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// Series is one line in a multi-series chart.
type Series struct {
	Label  string    `json:"label"`
	Values []float64 `json:"values"`
}

// RenderLine writes a line chart for the given series. months provides the
// chronological context for the caption (first → last). title is prepended to
// the caption shown below the chart. The plot uses one column per data point.
func RenderLine(w io.Writer, series []Series, months []domain.YearMonth, title string) {
	RenderLineWidth(w, series, months, title, 0)
}

// RenderLineWidth is like RenderLine but caps the plot width: when width > 0
// asciigraph resamples the series to that many columns. Use it for long series
// (e.g. a multi-year projection) that would otherwise overflow the terminal.
func RenderLineWidth(w io.Writer, series []Series, months []domain.YearMonth, title string, width int) {
	if len(series) == 0 || len(months) == 0 {
		return
	}
	data := make([][]float64, len(series))
	legends := make([]string, len(series))
	cs := make([]asciigraph.AnsiColor, len(series))
	for i, s := range series {
		data[i] = s.Values
		legends[i] = s.Label
		cs[i] = PaletteLine(i)
	}
	first := months[0]
	last := months[len(months)-1]
	caption := fmt.Sprintf("%s · %02d/%d → %02d/%d (%d meses)",
		title, first.Month, first.Year, last.Month, last.Year, len(months))

	plot := asciigraph.PlotMany(data,
		asciigraph.SeriesColors(cs...),
		asciigraph.SeriesLegends(legends...),
		asciigraph.Caption(caption),
		asciigraph.Height(15),
		asciigraph.Width(width),
	)
	fmt.Fprintln(w, plot)
}
