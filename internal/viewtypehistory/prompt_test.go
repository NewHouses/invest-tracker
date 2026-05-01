package viewtypehistory_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtypehistory"
)

type sumKey struct {
	id          int64
	year, month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[sumKey]domain.MonthlySummary
	months    map[int64][]domain.YearMonth
	listEr    error
	sumEr     error
	monthEr   error
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
	if f.monthEr != nil {
		return nil, f.monthEr
	}
	return f.months[id], nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewtypehistory.Run(r, &buf, repo)
	return buf.String(), err
}

// 2 acciones (AAPL, MSFT) e 1 índice (Vanguard, debe excluírse).
//
// 04/2026 (Acción agg):
//   AAPL: aporte 1000, holding 1000, result 1100 (+10%)
//   MSFT: aporte 500,  holding 500,  result 550  (+10%)
//   → aporte=1500, holding=1500, result=1650, G/P=+150, +10%
// 05/2026 (Acción agg):
//   AAPL: aporte 0,   holding 1100, result 1320 (+20%)
//   MSFT: aporte 0,   holding 550,  result 605  (+10%)
//   → aporte=0, holding=1650, result=1925, G/P=+275, ≈+16.67%
//
// Lifetime invested = 1000 + 500 = 1500.
// Lifetime result = 1320 + 605 = 1925; lifetime G/P = +425.
// Avg pct = (10 + 16.6667)/2 ≈ +13.33%; Avg gain = (150 + 275)/2 = +212.50.
func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
			{ID: 12, Type: domain.Indice, Name: "Vanguard"}, // tipo distinto: ignorado
		},
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
			11: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {InvestedInMonth: 1000, EstimatedHolding: 1000, Result: 1100, HasResult: true, TotalInvestedUpTo: 1000},
			{11, 2026, 4}: {InvestedInMonth: 500, EstimatedHolding: 500, Result: 550, HasResult: true, TotalInvestedUpTo: 500},
			{10, 2026, 5}: {InvestedInMonth: 0, EstimatedHolding: 1100, HasPrevResult: true, Result: 1320, HasResult: true, TotalInvestedUpTo: 1000},
			{11, 2026, 5}: {InvestedInMonth: 0, EstimatedHolding: 550, HasPrevResult: true, Result: 605, HasResult: true, TotalInvestedUpTo: 500},
			{10, 9999, 12}: {TotalInvestedUpTo: 1000},
			{11, 9999, 12}: {TotalInvestedUpTo: 500},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Reporte histórico dun tipo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsType(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Tipo de investimento:") {
		t.Errorf("saída non solicita o tipo:\n%s", out)
	}
}

func TestRun_NoAssetsOfType_PrintsHint(t *testing.T) {
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 12, Type: domain.Indice, Name: "Vanguard"},
		},
	}
	out, err := runWith(repo, "1\n") // tipo 1 = Acción, sen activos
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Acción") {
		t.Errorf("saída non contén suxestión sen activos do tipo:\n%s", out)
	}
}

func TestRun_NoResults_PrintsHint(t *testing.T) {
	repo := &fakeRepo{
		assets: []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL"}},
		months: map[int64][]domain.YearMonth{},
	}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados para activos de tipo Acción") {
		t.Errorf("saída non contén suxestión sen resultados:\n%s", out)
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

func TestRun_PrintsTypeAndAssetCount(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Tipo: Acción · 2 activo(s) · 2 mes(es)") {
		t.Errorf("saída non contén a cabeceira de tipo/conta:\n%s", out)
	}
}

func TestRun_PrintsSummary_AllMetrics(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total Aportado",
		"1500.00 USD",
		"Índice Medio Mensual",
		"+13.33%",
		"Gañanzas/Perdas Medias Mensuais",
		"+212.50 USD",
		"Total Gañanzas/Perdas",
		"+425.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableColumns_InOrder(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	start := strings.Index(out, "Ano")
	if start < 0 {
		t.Fatalf("non se atopou a cabeceira da táboa:\n%s", out)
	}
	table := out[start:]
	wantCols := []string{
		"Ano", "Mes", "Aporte Mensual", "No activo",
		"Índice", "G/P USD", "Resultado",
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

func TestRun_PrintsAggregatedRowValues(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		// Mes 04: aporte 1500, holding 1500, +10%, +150, result 1650
		"1500.00", "+10.00%", "+150.00", "1650.00",
		// Mes 05: aporte 0, holding 1650, +16.67%, +275, result 1925
		"0.00", "1650.00", "+16.67%", "+275.00", "1925.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_OnlyAssetsWithResultThisMonth(t *testing.T) {
	// AAPL ten resultado en 04 e 05; MSFT só en 05.
	// 04/2026 debe agregar só AAPL.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
		},
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
			11: {{Year: 2026, Month: 5}},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {InvestedInMonth: 1000, EstimatedHolding: 1000, Result: 1200, HasResult: true, TotalInvestedUpTo: 1000},
			// MSFT en 04: sen resultado, debería excluírse aínda que estivese cargado.
			{11, 2026, 4}: {InvestedInMonth: 500, EstimatedHolding: 500, HasResult: false},
			{10, 2026, 5}: {InvestedInMonth: 0, EstimatedHolding: 1200, HasPrevResult: true, Result: 1320, HasResult: true, TotalInvestedUpTo: 1000},
			{11, 2026, 5}: {InvestedInMonth: 500, EstimatedHolding: 500, Result: 550, HasResult: true, TotalInvestedUpTo: 500},
			{10, 9999, 12}: {TotalInvestedUpTo: 1000},
			{11, 9999, 12}: {TotalInvestedUpTo: 500},
		},
	}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Mes 04: só AAPL → aporte 1000, holding 1000, result 1200, +20%, +200
	// Mes 05: AAPL + MSFT → aporte 500, holding 1700, result 1870, ≈+10%, +170
	for _, want := range []string{
		"1000.00", "+20.00%", "+200.00", "1200.00",
		"500.00", "1700.00", "+10.00%", "+170.00", "1870.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}
