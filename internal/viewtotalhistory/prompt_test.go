package viewtotalhistory_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalhistory"
)

type sumKey struct {
	id    int64
	year  int
	month int
}

type divKey struct {
	year, month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[sumKey]domain.MonthlySummary
	dividends map[divKey]float64
	months    []domain.YearMonth
	listEr    error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) MonthlySummary(id int64, year, month int) (domain.MonthlySummary, error) {
	return f.summaries[sumKey{id, year, month}], nil
}

func (f *fakeRepo) SumDividends(year, month int) (float64, error) {
	return f.dividends[divKey{year, month}], nil
}

func (f *fakeRepo) MonthsWithResults() ([]domain.YearMonth, error) {
	return f.months, nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewtotalhistory.Run(r, &buf, repo)
	return buf.String(), err
}

var twoAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

// Setup: 2 activos, 2 meses con resultados (03/2026 e 04/2026), dividendos en ambos.
//
// 03/2026:
//   AAPL: invested=1000 holding=1000 result=1100
//   Vanguard: invested=2000 holding=2000 result=2100
//   Total: invested=3000, result=3200, holding_no_div=3000 (resultPrev=0)
//   Dividends: 20
//   gain_no_div = 3200 - 3000 = 200
//   pct_no_div = 200/3000 ≈ 6.67%
//   holding_with_div = 3000 + 0 = 3000 (no prev div)
//   gain_with_div = (3200 + 20) - 3000 = 220
//   pct_with_div = 220/3000 ≈ 7.33%
//
// 04/2026:
//   AAPL: invested ata=1200 (this month=200) holding=1300 result=1500
//   Vanguard: invested ata=2000 (this month=0) holding=2100 result=2150
//   Total: invested=3200, this month=200, result=3650, resultPrev=3200
//   Dividends: 50, prev=20
//   holding_no_div = 3200 + 200 = 3400
//   gain_no_div = 3650 - 3400 = 250
//   pct_no_div = 250/3400 ≈ 7.35%
//   holding_with_div = 3400 + 20 = 3420
//   gain_with_div = (3650 + 50) - 3420 = 280
//   pct_with_div = 280/3420 ≈ 8.19%
//
// Lifetime invested (consultado a 9999/12) = 3200 (final cumulative)
// Lifetime gain = 3650 - 3200 = 450
// Lifetime pct = 450/3200 ≈ 14.06%
//
// Avg pct (sen div) = (6.67 + 7.35)/2 ≈ 7.01
// Avg gain (sen div) = (200 + 250)/2 = 225
// Avg pct (con div) = (7.33 + 8.19)/2 ≈ 7.76
// Avg gain (con div) = (220 + 280)/2 = 250

func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: twoAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 3}: {
				TotalInvestedUpTo: 1000, InvestedInMonth: 1000,
				EstimatedHolding: 1000, Result: 1100, HasResult: true,
			},
			{11, 2026, 3}: {
				TotalInvestedUpTo: 2000, InvestedInMonth: 2000,
				EstimatedHolding: 2000, Result: 2100, HasResult: true,
			},
			{10, 2026, 4}: {
				TotalInvestedUpTo: 1200, InvestedInMonth: 200,
				EstimatedHolding: 1300, Result: 1500, HasResult: true, HasPrevResult: true,
			},
			{11, 2026, 4}: {
				TotalInvestedUpTo: 2000, InvestedInMonth: 0,
				EstimatedHolding: 2100, Result: 2150, HasResult: true, HasPrevResult: true,
			},
			// Lifetime probe (9999, 12)
			{10, 9999, 12}: {TotalInvestedUpTo: 1200},
			{11, 9999, 12}: {TotalInvestedUpTo: 2000},
		},
		dividends: map[divKey]float64{
			{2026, 3}: 20,
			{2026, 4}: 50,
		},
		months: []domain.YearMonth{
			{Year: 2026, Month: 3},
			{Year: 2026, Month: 4},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Resultado xeral") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
	if !strings.Contains(out, "2 mes(es) con resultado") {
		t.Errorf("saída non sinala número de meses:\n%s", out)
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

func TestRun_NoMonths_PrintsHint(t *testing.T) {
	repo := &fakeRepo{assets: twoAssets}
	out, err := runWith(repo, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_PrintsAllRequiredHeaderLabels(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total investido",
		"Total Gañanzas/Perdas",
		"Total Índice",
		"Índice medio mensual (sen div)",
		"Gañanza media mensual (sen div)",
		"Índice medio mensual (con div)",
		"Gañanza media mensual (con div)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableHeaders(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Investimentos por mes:",
		"Ano",
		"Mes",
		"Investido total",
		"Este mes",
		"+Div prev",
		"No activo s/d",
		"No activo c/d",
		"Resultados e gañanzas por mes:",
		"Result s/d",
		"Result c/d",
		"G/P s/d",
		"G/P c/d",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsLifetimeTotals(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// lifetimeInvested = 3200, gain = 3650-3200 = 450, pct = 450/3200 ≈ 14.06%
	for _, want := range []string{
		"3200.00 USD",
		"+450.00 USD",
		"+14.06%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAverages(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"+7.01",      // avg pct sen div
		"+225.00 USD", // avg gain sen div
		"+7.76",      // avg pct con div
		"+250.00 USD", // avg gain con div
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRowFor03(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 03/2026: invested=3000 (este mes=3000) +DivPrev=0 holdingNoDiv=3000 holdingWithDiv=3000
	// dividends=20 result=3200 resultcd=3220 gain=200 +220 pct=6.67% +7.33%
	for _, want := range []string{
		"3000.00",
		"+200.00",
		"+220.00",
		"+6.67%",
		"+7.33%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRowFor04(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 04/2026: holdingNoDiv=3400 holdingWithDiv=3420 result=3650 result_cd=3700 gain=250 +280 pct=7.35% +8.19%
	for _, want := range []string{
		"3400.00",
		"3420.00",
		"3650.00",
		"3700.00",
		"+250.00",
		"+280.00",
		"+7.35%",
		"+8.19%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HandlesMonthWithoutMetrics(t *testing.T) {
	// Mes inicial sen prev result e sen invested = holdingNoDiv = 0 → sen métricas
	repo := &fakeRepo{
		assets: twoAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			// Sen TotalInvestedUpTo nin holding nin result; só un resultado "fantasma"
			{10, 2026, 5}: {Result: 100, HasResult: true},
			{10, 9999, 12}: {TotalInvestedUpTo: 0},
			{11, 9999, 12}: {TotalInvestedUpTo: 0},
		},
		months: []domain.YearMonth{{Year: 2026, Month: 5}},
	}
	out, err := runWith(repo, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "—") {
		t.Errorf("saída non contén guión para mes sen métricas:\n%s", out)
	}
}

func TestRun_PrintsRowsInOrder(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 3000.00 (03) debe aparecer antes de 3400.00 (holding 04) na táboa A
	idxA := strings.Index(out, "Investimentos por mes:")
	if idxA < 0 {
		t.Fatalf("Cabeceira A non atopada")
	}
	tableA := out[idxA:]
	pos03 := strings.Index(tableA, "3000.00")
	pos04 := strings.Index(tableA, "3400.00")
	if pos03 < 0 || pos04 < 0 || pos03 >= pos04 {
		t.Errorf("filas non en orde cronolóxica: pos(3000)=%d pos(3400)=%d\n%s", pos03, pos04, tableA)
	}
}
