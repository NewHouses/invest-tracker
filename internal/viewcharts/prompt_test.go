package viewcharts_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewcharts"
)

type sumKey struct {
	id          int64
	year, month int
}

type divKey struct {
	year, month int
}

type fakeRepo struct {
	assets         []domain.Asset
	summaries      map[sumKey]domain.MonthlySummary
	monthsForAsset map[int64][]domain.YearMonth
	allMonths      []domain.YearMonth
	dividends      map[divKey]float64
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) { return f.assets, nil }
func (f *fakeRepo) MonthlySummary(id int64, y, m int) (domain.MonthlySummary, error) {
	return f.summaries[sumKey{id, y, m}], nil
}
func (f *fakeRepo) MonthsWithResults() ([]domain.YearMonth, error) { return f.allMonths, nil }
func (f *fakeRepo) MonthsWithResultsForAsset(id int64) ([]domain.YearMonth, error) {
	return f.monthsForAsset[id], nil
}
func (f *fakeRepo) SumDividends(y, m int) (float64, error) {
	return f.dividends[divKey{y, m}], nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewcharts.Run(r, &buf, repo)
	return buf.String(), err
}

// Setup base: 2 acciones (AAPL, MSFT) e 1 índice (Vanguard). Resultados en 04 e 05.
// AAPL: lifetime invested 1500 (1000 inicial + 500 tx).
// MSFT: lifetime invested 600.
// Vanguard: lifetime invested 2000.
// Total lifetime: 4100.
func setup() *fakeRepo {
	return &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
			{ID: 12, Type: domain.Indice, Name: "Vanguard"},
		},
		monthsForAsset: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
			11: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
			12: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
		},
		allMonths: []domain.YearMonth{{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}:  {InvestedInMonth: 1000, EstimatedHolding: 1000, Result: 1100, HasResult: true, TotalInvestedUpTo: 1000},
			{11, 2026, 4}:  {InvestedInMonth: 600, EstimatedHolding: 600, Result: 650, HasResult: true, TotalInvestedUpTo: 600},
			{12, 2026, 4}:  {InvestedInMonth: 2000, EstimatedHolding: 2000, Result: 2100, HasResult: true, TotalInvestedUpTo: 2000},
			{10, 2026, 5}:  {InvestedInMonth: 500, EstimatedHolding: 1600, Result: 1700, HasResult: true, TotalInvestedUpTo: 1500},
			{11, 2026, 5}:  {InvestedInMonth: 0, EstimatedHolding: 650, Result: 700, HasResult: true, TotalInvestedUpTo: 600},
			{12, 2026, 5}:  {InvestedInMonth: 0, EstimatedHolding: 2100, Result: 2200, HasResult: true, TotalInvestedUpTo: 2000},
			{10, 9999, 12}: {TotalInvestedUpTo: 1500},
			{11, 9999, 12}: {TotalInvestedUpTo: 600},
			{12, 9999, 12}: {TotalInvestedUpTo: 2000},
		},
		dividends: map[divKey]float64{
			{2026, 4}: 10,
			{2026, 5}: 20,
		},
	}
}

func TestRun_PrintsHeaderAndSubmenu(t *testing.T) {
	out, err := runWith(setup(), "0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Ver gráficas",
		"[1] Distribución de aportes",
		"[2] Evolución dun activo",
		"[3] Evolución dun tipo (agregada)",
		"[4] Evolución dos activos dun tipo",
		"[5] Evolución dos tipos",
		"[6] Evolución do resultado total",
		"[0] Voltar",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_RejectsInvalidChoice(t *testing.T) {
	out, err := runWith(setup(), "abc\n7\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
}

func TestRun_VoltarReturns(t *testing.T) {
	_, err := runWith(setup(), "0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_EOFReturnsError(t *testing.T) {
	_, err := runWith(setup(), "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

// ---- Chart 1: Distribución ----

func TestChart1_RendersBars(t *testing.T) {
	// Selecciona 1, despois 0 para saír.
	out, err := runWith(setup(), "1\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Distribución de aportes",
		// Total = 1500 + 600 + 2000 = 4100
		"$4100.00",
		"AAPL", "MSFT", "Vanguard",
		// Porcentaxes esperadas: Vanguard 48.8%, AAPL 36.6%, MSFT 14.6%
		"48.8%", "36.6%", "14.6%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestChart1_NoAssets_PrintsHint(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "1\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

// ---- Chart 2: Evolución dun activo ----

func TestChart2_PromptsAssetAndRenders(t *testing.T) {
	// Submenú=2, selecciona activo 1 (AAPL), despois Voltar.
	out, err := runWith(setup(), "2\n1\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Evolución dun activo",
		"Selecciona (1-3):",
		"Resultado",
		"Aporte acumulado",
		"Acción — AAPL",
		"04/2026", "05/2026",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

// ---- Chart 3: Evolución dun tipo (agregada) ----

func TestChart3_PromptsTypeAndRenders(t *testing.T) {
	// Submenú=3, type=1 (Acción), despois Voltar.
	out, err := runWith(setup(), "3\n1\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Evolución dun tipo (agregada)",
		"Tipo de investimento:",
		"Resultado agregado",
		"Aporte acumulado",
		"Tipo: Acción",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestChart3_NoAssetsOfType_PrintsHint(t *testing.T) {
	repo := setup()
	repo.assets = []domain.Asset{
		{ID: 12, Type: domain.Indice, Name: "Vanguard"},
	}
	out, err := runWith(repo, "3\n1\n0\n") // tipo Acción, sen activos
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Acción") {
		t.Errorf("saída non contén suxestión sen activos:\n%s", out)
	}
}

// ---- Chart 4: Evolución dos activos dun tipo ----

func TestChart4_RendersOneSeriesPerAsset(t *testing.T) {
	// Submenú=4, type=1 (Acción), Voltar.
	out, err := runWith(setup(), "4\n1\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Evolución dos activos dun tipo",
		"Activos de tipo Acción",
		// Lendas: nomes dos activos do tipo Acción
		"AAPL", "MSFT",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

// ---- Chart 5: Evolución dos tipos ----

func TestChart5_RendersOneSeriesPerType(t *testing.T) {
	out, err := runWith(setup(), "5\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Evolución dos tipos",
		"Resultado por tipo",
		"Acción", "Índice",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

// ---- Chart 6: Evolución do resultado total ----

func TestChart6_RendersTotal(t *testing.T) {
	out, err := runWith(setup(), "6\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Evolución do resultado total",
		"Resultado total da carteira",
		"Resultado + dividendos acum.",
		"04/2026", "05/2026",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestChart6_NoMonths_PrintsHint(t *testing.T) {
	repo := setup()
	repo.allMonths = nil
	out, err := runWith(repo, "6\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados") {
		t.Errorf("saída non contén suxestión sen meses:\n%s", out)
	}
}
