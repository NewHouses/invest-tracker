package viewassethistory_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassethistory"
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
	err := viewassethistory.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

// Escenario base:
//   04/2026: aporte 500, no activo 1500, resultado 1800 → G/P +300, +20%
//   05/2026: aporte 200, no activo 2000 (1800 prev + 200), resultado 1900 → G/P -100, -5%
// Total Aportado lifetime = 1700.
func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: sampleAssets,
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}, {Year: 2026, Month: 5}},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {
				TotalInvestedUpTo: 1500, InvestedInMonth: 500,
				EstimatedHolding: 1500, Result: 1800, HasResult: true,
			},
			{10, 2026, 5}: {
				TotalInvestedUpTo: 1700, InvestedInMonth: 200,
				EstimatedHolding: 2000, HasPrevResult: true,
				Result: 1900, HasResult: true,
			},
			// Lifetime call → Year=9999, Month=12.
			{10, 9999, 12}: {TotalInvestedUpTo: 1700},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Reporte histórico dun activo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
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

func TestRun_PromptsSelection(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona (1-2):") {
		t.Errorf("saída non contén o prompt:\n%s", out)
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén a suxestión:\n%s", out)
	}
}

func TestRun_NoResults_PrintsHint(t *testing.T) {
	repo := &fakeRepo{assets: sampleAssets, months: map[int64][]domain.YearMonth{}}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados para este activo") {
		t.Errorf("saída non contén a suxestión sobre resultados:\n%s", out)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, err := runWith(gainSetup(), "99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
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

func TestRun_PrintsSummary_AllMetrics(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total Aportado",
		"1700.00 USD",
		"Índice Medio Mensual",
		"+7.50%",
		"Gañanzas/Perdas Medias Mensuais",
		"+100.00 USD",
		"Total Gañanzas/Perdas",
		"+200.00 USD",
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
	// Limítase a busca á sección da táboa (a partir de "Ano") para evitar
	// colisións con "Índice" da táboa resumo de arriba.
	start := strings.Index(out, "Ano")
	if start < 0 {
		t.Fatalf("non se atopou a cabeceira da táboa:\n%s", out)
	}
	table := out[start:]
	wantCols := []string{
		"Ano",
		"Mes",
		"Aporte Mensual",
		"No activo",
		"Índice",
		"G/P USD",
		"Resultado",
	}
	prev := -1
	for _, col := range wantCols {
		idx := strings.Index(table, col)
		if idx < 0 {
			t.Errorf("saída non contén a columna %q:\n%s", col, table)
			continue
		}
		if idx <= prev {
			t.Errorf("orde de columnas incorrecta para %q (idx=%d, prev=%d)", col, idx, prev)
		}
		prev = idx
	}
}

func TestRun_PrintsRowValues(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		// 04/2026: aporte 500, no activo 1500, +20%, +300, 1800
		"500.00", "1500.00", "+20.00%", "+300.00", "1800.00",
		// 05/2026: aporte 200, no activo 2000, -5%, -100, 1900
		"200.00", "2000.00", "-5.00%", "-100.00", "1900.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_LossOnly(t *testing.T) {
	repo := &fakeRepo{
		assets: sampleAssets,
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {
				TotalInvestedUpTo: 1000, InvestedInMonth: 1000,
				EstimatedHolding: 1000, Result: 800, HasResult: true,
			},
			{10, 9999, 12}: {TotalInvestedUpTo: 1000},
		},
	}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-200.00") {
		t.Errorf("saída non mostra perda absoluta:\n%s", out)
	}
	if !strings.Contains(out, "-20.00%") {
		t.Errorf("saída non mostra perda %%:\n%s", out)
	}
}

func TestRun_ZeroHoldingRow_ShowsNA(t *testing.T) {
	repo := &fakeRepo{
		assets: sampleAssets,
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {
				TotalInvestedUpTo: 0, InvestedInMonth: 0,
				EstimatedHolding: 0, Result: 100, HasResult: true,
			},
			{10, 9999, 12}: {TotalInvestedUpTo: 0},
		},
	}
	out, err := runWith(repo, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "n/a") {
		t.Errorf("saída non contén n/a para o índice cando holding=0:\n%s", out)
	}
}
