package charts_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"invest-tracker/internal/charts"
	"invest-tracker/internal/domain"
)

func TestRenderBars_SortedAndPercents(t *testing.T) {
	var buf bytes.Buffer
	charts.RenderBars(&buf, []charts.BarItem{
		{Label: "MSFT", Value: 500},
		{Label: "AAPL", Value: 1500},
		{Label: "GOOG", Value: 1000},
	}, 3000)
	out := buf.String()

	// Orde: AAPL (50%) → GOOG (33.3%) → MSFT (16.7%)
	idxAAPL := strings.Index(out, "AAPL")
	idxGOOG := strings.Index(out, "GOOG")
	idxMSFT := strings.Index(out, "MSFT")
	if !(idxAAPL >= 0 && idxGOOG > idxAAPL && idxMSFT > idxGOOG) {
		t.Errorf("orde inesperada: AAPL=%d GOOG=%d MSFT=%d\n%s",
			idxAAPL, idxGOOG, idxMSFT, out)
	}

	// Porcentaxes esperadas
	for _, want := range []string{"50.0%", "33.3%", "16.7%"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}

	// Cantidades USD
	for _, want := range []string{"$1500.00", "$1000.00", "$500.00"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRenderBars_EmptyOrZeroTotal_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	charts.RenderBars(&buf, nil, 100)
	if buf.Len() != 0 {
		t.Errorf("agardábase saída baleira para nil items, got %q", buf.String())
	}
	charts.RenderBars(&buf, []charts.BarItem{{Label: "X", Value: 1}}, 0)
	if buf.Len() != 0 {
		t.Errorf("agardábase saída baleira con total=0, got %q", buf.String())
	}
}

func TestRenderBars_AppliesColors(t *testing.T) {
	var buf bytes.Buffer
	charts.RenderBars(&buf, []charts.BarItem{
		{Label: "X", Value: 100, Color: "\x1b[31m"},
	}, 100)
	if !strings.Contains(buf.String(), "\x1b[31m") {
		t.Errorf("color non aplicado:\n%s", buf.String())
	}
}

func TestRenderLine_IncludesLegendsAndCaption(t *testing.T) {
	var buf bytes.Buffer
	months := []domain.YearMonth{
		{Year: 2026, Month: 4},
		{Year: 2026, Month: 5},
		{Year: 2026, Month: 6},
	}
	series := []charts.Series{
		{Label: "Resultado", Values: []float64{100, 110, 120}},
		{Label: "Aporte", Values: []float64{100, 100, 100}},
	}
	charts.RenderLine(&buf, series, months, "Test Title")
	out := buf.String()

	// Lendas presentes
	for _, want := range []string{"Resultado", "Aporte"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén lenda %q:\n%s", want, out)
		}
	}
	// Caption coas datas e título
	for _, want := range []string{"Test Title", "04/2026", "06/2026", "3 meses"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q no caption:\n%s", want, out)
		}
	}
}

func TestRenderLine_EmptySeries_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	charts.RenderLine(&buf, nil, []domain.YearMonth{{Year: 2026, Month: 4}}, "X")
	if buf.Len() != 0 {
		t.Errorf("agardábase saída baleira para series=nil, got %q", buf.String())
	}
}

// Con moitos puntos (p.e. os 240 meses dunha proxección) o gráfico debe
// remostrearse ao ancho pedido en lugar de usar unha columna por punto.
func TestRenderLineWidth_LimitsPlotWidth(t *testing.T) {
	const n = 240
	const width = 100

	months := make([]domain.YearMonth, n)
	values := make([]float64, n)
	for i := range months {
		months[i] = domain.YearMonth{Year: 2026, Month: 1}.AddMonths(i)
		values[i] = float64(i)
	}
	series := []charts.Series{{Label: "Capital total", Values: values}}

	var narrow, auto bytes.Buffer
	charts.RenderLineWidth(&narrow, series, months, "Proxección", width)
	charts.RenderLine(&auto, series, months, "Proxección")

	if maxLineLen(narrow.String()) >= maxLineLen(auto.String()) {
		t.Errorf("a gráfica limitada (%d) debe ser máis estreita ca a automática (%d)",
			maxLineLen(narrow.String()), maxLineLen(auto.String()))
	}
	for _, want := range []string{"Capital total", "Proxección", "01/2026", "12/2045", "240 meses"} {
		if !strings.Contains(narrow.String(), want) {
			t.Errorf("saída non contén %q", want)
		}
	}
}

func TestRenderLineWidth_EmptyMonths_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	charts.RenderLineWidth(&buf, []charts.Series{{Label: "X", Values: []float64{1}}}, nil, "X", 50)
	if buf.Len() != 0 {
		t.Errorf("agardábase saída baleira sen meses, got %q", buf.String())
	}
}

func maxLineLen(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		if n := utf8.RuneCountInString(line); n > max {
			max = n
		}
	}
	return max
}

func TestPaletteFG_Cyclical(t *testing.T) {
	c0 := charts.PaletteFG(0)
	c1 := charts.PaletteFG(1)
	cN := charts.PaletteFG(100)
	if c0 == "" || c1 == "" {
		t.Errorf("cores baleiras: c0=%q c1=%q", c0, c1)
	}
	if c0 == c1 {
		t.Errorf("c0 e c1 deben ser diferentes")
	}
	// O índice 100 debe coincidir con algún c[100 % len(palette)].
	if cN == "" {
		t.Errorf("PaletteFG(100) baleiro")
	}
}
