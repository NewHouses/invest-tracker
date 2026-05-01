package viewreport_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewreport"
)

type summaryKey struct {
	id    int64
	year  int
	month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[summaryKey]domain.MonthlySummary
	listEr    error
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
	return f.summaries[summaryKey{id, year, month}], nil
}

func runWith(assets []domain.Asset, summaries map[summaryKey]domain.MonthlySummary, input string) (string, error) {
	repo := &fakeRepo{assets: assets, summaries: summaries}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewreport.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

func gainSummary() map[summaryKey]domain.MonthlySummary {
	return map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {
			TotalInvestedUpTo: 1500, InvestedInMonth: 500,
			EstimatedHolding: 1500, Result: 1800, HasResult: true,
		},
	}
}

const validInput = "1\n5\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Informe mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Selecciona (1-2):", "Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén a suxestión:\n%s", out)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), "99\n1\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, err := runWith(sampleAssets, gainSummary(), "1\n5\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_PrintsTable_AllMetrics(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Investido ata o mes",
		"Investido este mes",
		"No activo",
		"Resultado",
		"Gañanzas/Perdas",
		"Índice",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsHeader_AssetAndPeriod(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Acción — AAPL · 05/2026") {
		t.Errorf("saída non contén a cabeceira do informe:\n%s", out)
	}
}

func TestRun_ShowsGain(t *testing.T) {
	out, err := runWith(sampleAssets, gainSummary(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+300.00 USD") {
		t.Errorf("saída non mostra a gañanza absoluta:\n%s", out)
	}
	if !strings.Contains(out, "+20.00%") {
		t.Errorf("saída non mostra a gañanza %%:\n%s", out)
	}
}

func TestRun_ShowsLoss(t *testing.T) {
	summaries := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {
			TotalInvestedUpTo: 1500, InvestedInMonth: 500,
			EstimatedHolding: 1500, Result: 1200, HasResult: true,
		},
	}
	out, err := runWith(sampleAssets, summaries, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-300.00 USD") {
		t.Errorf("saída non mostra a perda absoluta:\n%s", out)
	}
	if !strings.Contains(out, "-20.00%") {
		t.Errorf("saída non mostra a perda %%:\n%s", out)
	}
}

func TestRun_ShowsBreakeven(t *testing.T) {
	summaries := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {
			TotalInvestedUpTo: 1000, InvestedInMonth: 0,
			EstimatedHolding: 1000, Result: 1000, HasResult: true,
		},
	}
	out, err := runWith(sampleAssets, summaries, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+0.00 USD") {
		t.Errorf("saída non mostra +0.00 USD:\n%s", out)
	}
	if !strings.Contains(out, "+0.00%") {
		t.Errorf("saída non mostra +0.00%%:\n%s", out)
	}
}

func TestRun_NoResult_ShowsDashes(t *testing.T) {
	summaries := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {
			TotalInvestedUpTo: 1500, InvestedInMonth: 500,
			EstimatedHolding: 1500, Result: 0, HasResult: false,
		},
	}
	out, err := runWith(sampleAssets, summaries, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Count(out, "—") < 3 {
		t.Errorf("saída non contén pelo menos 3 guións longos para os campos sen resultado:\n%s", out)
	}
}

func TestRun_ZeroInvested_ShowsNA(t *testing.T) {
	summaries := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {
			TotalInvestedUpTo: 0, InvestedInMonth: 0,
			EstimatedHolding: 0, Result: 100, HasResult: true,
		},
	}
	out, err := runWith(sampleAssets, summaries, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "n/a") {
		t.Errorf("saída non contén n/a para o índice:\n%s", out)
	}
}

// Test de regresión: cando hai prev_result, EstimatedHolding != TotalInvestedUpTo,
// e a gañanza calcúlase contra o holding (non contra o cost basis acumulado).
// Isto garante coherencia con viewassetgeneral.
func TestRun_UsesEstimatedHoldingForGain(t *testing.T) {
	// Holding=1300 (prev result 1100 + invested este mes 200).
	// Result=1500. Gain = 1500 - 1300 = 200, pct = 200/1300 ≈ 15.38%.
	// Se usase TotalInvestedUpTo (1200), gain sería 300 e pct 25%.
	summaries := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1200,
			InvestedInMonth:   200,
			EstimatedHolding:  1300,
			HasPrevResult:     true,
			Result:            1500,
			HasResult:         true,
		},
	}
	out, err := runWith(sampleAssets, summaries, "1\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "No activo  1300.00 USD") {
		// Pode haber espazos por tabwriter; só verificamos que aparece a cifra.
		if !strings.Contains(out, "1300.00 USD") {
			t.Errorf("saída non contén 'No activo: 1300.00 USD':\n%s", out)
		}
	}
	if !strings.Contains(out, "+200.00 USD") {
		t.Errorf("saída non contén gañanza +200.00 (gain contra holding):\n%s", out)
	}
	if !strings.Contains(out, "+15.38%") {
		t.Errorf("saída non contén pct +15.38%% (contra holding):\n%s", out)
	}
}
