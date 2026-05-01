package viewassetgeneral_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassetgeneral"
)

type sumKey struct {
	id    int64
	year  int
	month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[sumKey]domain.MonthlySummary
	months    map[int64][]domain.YearMonth
	listEr    error
	monthsEr  error
	sumEr     error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) MonthlySummary(id int64, year, month int) (domain.MonthlySummary, error) {
	if f.sumEr != nil {
		return domain.MonthlySummary{}, f.sumEr
	}
	return f.summaries[sumKey{id, year, month}], nil
}

func (f *fakeRepo) MonthsWithResultsForAsset(id int64) ([]domain.YearMonth, error) {
	if f.monthsEr != nil {
		return nil, f.monthsEr
	}
	return f.months[id], nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewassetgeneral.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

// 03/2026 (mes inicial): invested=1000, holding=1000 (no prev), result=1100,
//   gain=+100, pct=+10.00%
// 04/2026: invested=200 (este mes), totalUpTo=1200, holding=1100+200=1300
//   (prev result 1100), result=1500, gain=+200, pct=+15.38%
//
// Lifetime: invested=1200, gain=1500-1200=300, pct=300/1200=25%
// Avg: pct (10+15.38)/2 ≈ 12.69%, gain (100+200)/2 = 150
func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: sampleAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 3}: {
				TotalInvestedUpTo: 1000, InvestedInMonth: 1000,
				EstimatedHolding: 1000, Result: 1100, HasResult: true,
			},
			{10, 2026, 4}: {
				TotalInvestedUpTo: 1200, InvestedInMonth: 200,
				EstimatedHolding: 1300, Result: 1500, HasResult: true, HasPrevResult: true,
			},
			// Lifetime probe
			{10, 9999, 12}: {TotalInvestedUpTo: 1200},
		},
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 3}, {Year: 2026, Month: 4}},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Resultado xeral dun activo") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAssetNameInTitle(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Acción — AAPL") {
		t.Errorf("saída non contén nome do activo:\n%s", out)
	}
	if !strings.Contains(out, "2 mes(es) con resultado") {
		t.Errorf("saída non contén contador de meses:\n%s", out)
	}
}

func TestRun_NoAssets_PrintsHint(t *testing.T) {
	repo := &fakeRepo{}
	out, err := runWith(repo, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_NoMonthsForAsset_PrintsHint(t *testing.T) {
	repo := &fakeRepo{
		assets: sampleAssets,
		months: map[int64][]domain.YearMonth{},
	}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados para este activo") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_PrintsHeaderLabels(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total investido",
		"Total Gañanzas/Perdas",
		"Total Índice",
		"Índice medio mensual",
		"Gañanza media mensual",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableHeaders(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Investimentos por mes:",
		"Investido total",
		"Este mes",
		"No activo",
		"Resultados e gañanzas por mes:",
		"Resultado",
		"G/P USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsLifetimeTotals(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// lifetime: invested=1200, gain=300, pct=25.00%
	for _, want := range []string{
		"1200.00 USD",
		"+300.00 USD",
		"+25.00%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAverages(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// avg pct = (10.00 + 15.38)/2 ≈ 12.69
	// avg gain = (100 + 200)/2 = 150
	for _, want := range []string{
		"+12.69",
		"+150.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRowFor03(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 03/2026: total=1000 este mes=1000 holding=1000 result=1100 gain=+100 pct=+10.00%
	for _, want := range []string{
		"1000.00",
		"1100.00",
		"+100.00",
		"+10.00%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRowFor04(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 04/2026: total=1200 este mes=200 holding=1300 result=1500 gain=+200 pct=+15.38%
	for _, want := range []string{
		"1200.00",
		"200.00",
		"1300.00",
		"1500.00",
		"+200.00",
		"+15.38%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_RecoversFromInvalidSelection(t *testing.T) {
	out, err := runWith(gainSetup(), "99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
	if !strings.Contains(out, "Acción — AAPL") {
		t.Errorf("non chegou ao informe tras recuperarse:\n%s", out)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, err := runWith(gainSetup(), "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
