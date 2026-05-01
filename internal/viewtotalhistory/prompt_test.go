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

// 03/2026:
//   AAPL: tx=1000, fondos=1000, result=1100
//   Vanguard: tx=2000, fondos=2000, result=2100
//   div=20 → agg: aporte=(3000−20)=2980, fondos=3000, base=3200, div=20,
//                 result=3220, G/P=+220, +7.33%
// 04/2026:
//   AAPL: tx=200, fondos=1300, result=1500
//   Vanguard: tx=0, fondos=2100, result=2150
//   div=50 → agg: aporte=(200−50)=150, fondos=3400, base=3650, div=50,
//                 result=3700, G/P=+300, +8.82%
//
// Lifetime: aporte=3200 (de 9999/12), últimos resultados=1500+2150=3650,
//           totalDiv=70, G/P Total = 3650+70-3200 = +520.
// Avg pct ≈ +8.08, Avg gain = 260.
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
	if !strings.Contains(out, "Reporte histórico completo") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
	if !strings.Contains(out, "2 activo(s) · 2 mes(es) con resultado") {
		t.Errorf("saída non sinala número de activos/meses:\n%s", out)
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

func TestRun_PrintsSummary_AllMetrics(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Aporte histórico total",
		"3200.00 USD",
		"Índice Medio",
		"+8.08%",
		"G/P Media",
		"+260.00 USD",
		"G/P Total",
		"+520.00 USD",
		"Dividendos totais",
		"70.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableColumns_InOrder(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	start := strings.Index(out, "Ano")
	if start < 0 {
		t.Fatalf("non se atopou a cabeceira da táboa:\n%s", out)
	}
	table := out[start:]
	wantCols := []string{
		"Ano", "Mes", "Aporte Mensual", "Fondos",
		"Índice", "G/P USD", "Dividendos", "Resultado",
	}
	prev := -1
	for _, col := range wantCols {
		idx := strings.Index(table, col)
		if idx < 0 {
			t.Errorf("non aparece a columna %q:\n%s", col, table)
			continue
		}
		if idx <= prev {
			t.Errorf("orde de columnas incorrecta para %q (idx=%d, prev=%d)", col, idx, prev)
		}
		prev = idx
	}
}

func TestRun_PrintsRow_03(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 03/2026: aporte=2980 (3000−20), fondos=3000, +7.33%, +220, div=20, result=3220
	for _, want := range []string{
		"2980.00", "3000.00", "+7.33%", "+220.00", "20.00", "3220.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRow_04(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 04/2026: aporte=150 (200−50), fondos=3400, +8.82%, +300, div=50, result=3700
	for _, want := range []string{
		"150.00", "3400.00", "+8.82%", "+300.00", "50.00", "3700.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HandlesMonthWithoutMetrics(t *testing.T) {
	// Mes con resultado pero EstimatedHolding=0 → métricas n/a, gain "—".
	repo := &fakeRepo{
		assets: twoAssets,
		summaries: map[sumKey]domain.MonthlySummary{
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
	if !strings.Contains(out, "n/a") {
		t.Errorf("saída non contén n/a para mes sen métricas:\n%s", out)
	}
}

func TestRun_PrintsRowsInOrder(t *testing.T) {
	out, err := runWith(gainSetup(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	start := strings.Index(out, "Ano")
	if start < 0 {
		t.Fatalf("Cabeceira da táboa non atopada")
	}
	table := out[start:]
	// fondos do 03 (3000.00) debe aparecer antes de fondos do 04 (3400.00)
	pos03 := strings.Index(table, "3000.00")
	pos04 := strings.Index(table, "3400.00")
	if pos03 < 0 || pos04 < 0 || pos03 >= pos04 {
		t.Errorf("filas non en orde cronolóxica: pos(3000)=%d pos(3400)=%d\n%s", pos03, pos04, table)
	}
}
