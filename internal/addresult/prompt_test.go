package addresult_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addresult"
	"invest-tracker/internal/domain"
)

type summaryKey struct {
	id    int64
	year  int
	month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[summaryKey]domain.MonthlySummary
	saved     []domain.MonthlyResult
	listEr    error
	sumEr     error
	saveEr    error
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

func (f *fakeRepo) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	if f.saveEr != nil {
		return 0, f.saveEr
	}
	f.saved = append(f.saved, m)
	return int64(len(f.saved)), nil
}

func runWith(assets []domain.Asset, summaries map[summaryKey]domain.MonthlySummary, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, summaries: summaries}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addresult.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard S&P 500"},
}

// summariesForHolding constrúe summaries onde EstimatedHolding == holding
// (sen previo, equivalente ao caso onde non hai resultado anterior).
func summariesForHolding(year, month int, holdings map[int64]float64) map[summaryKey]domain.MonthlySummary {
	out := make(map[summaryKey]domain.MonthlySummary)
	for id, h := range holdings {
		out[summaryKey{id, year, month}] = domain.MonthlySummary{
			TotalInvestedUpTo: h,
			EstimatedHolding:  h,
		}
	}
	return out
}

// Input layout: month, year, asset, result
const validInput = "4\n2026\n1\n1100\n"

func TestRun_PrintsHeader(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000, 11: 2000})
	out, _, err := runWith(sampleAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir resultado mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsMonthAndYearFirst(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000, 11: 2000})
	out, _, err := runWith(sampleAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
	idxMes := strings.Index(out, "Mes (1-12):")
	idxAno := strings.Index(out, "Ano:")
	idxSel := strings.Index(out, "Selecciona (1-")
	if !(idxMes < idxAno && idxAno < idxSel) {
		t.Errorf("orde esperada: Mes < Ano < Selecciona; got mes=%d ano=%d sel=%d", idxMes, idxAno, idxSel)
	}
}

func TestRun_PrintsAssetListWithHolding(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000, 11: 2000})
	out, _, err := runWith(sampleAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Investimentos:",
		"[1] Acción — AAPL (no activo: 1000.00 USD)",
		"[2] Índice — Vanguard S&P 500 (no activo: 2000.00 USD)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_FiltersAssetsWithZeroHolding(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000}) // só id=10 con capital
	out, _, err := runWith(sampleAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "[1] Acción — AAPL") {
		t.Errorf("saída non contén AAPL:\n%s", out)
	}
	if strings.Contains(out, "Vanguard") {
		t.Errorf("saída non debería conter Vanguard (sen capital en 04/2026):\n%s", out)
	}
	if !strings.Contains(out, "Selecciona (1-1):") {
		t.Errorf("debería listar só 1 activo elixible:\n%s", out)
	}
}

func TestRun_NoEligibleAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(sampleAssets, summariesForHolding(2026, 4, nil), "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos con capital investido en 04/2026") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000, 11: 2000})
	out, _, err := runWith(sampleAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Mes (1-12):",
		"Ano:",
		"Selecciona (1-2):",
		"Resultado (USD):",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	out, repo, err := runWith(sampleAssets, sums, "4\n2026\n1\n1100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "✓ Resultado gardado") {
		t.Errorf("saída non contén a confirmación:\n%s", out)
	}
	if !strings.Contains(out, "No activo: 1000.00 USD") {
		t.Errorf("saída non contén o detalle 'No activo':\n%s", out)
	}
	if !strings.Contains(out, "Gañanzas/Perdas:") {
		t.Errorf("saída non contén Gañanzas/Perdas:\n%s", out)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
}

func TestRun_HappyPath_SavesCorrectResult(t *testing.T) {
	sums := summariesForHolding(2026, 5, map[int64]float64{10: 1000, 11: 2000})
	_, repo, err := runWith(sampleAssets, sums, "5\n2026\n2\n2200\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
	got := repo.saved[0]
	want := domain.MonthlyResult{
		AssetID:   11,
		ResultUSD: 2200,
		Month:     5,
		Year:      2026,
	}
	if got != want {
		t.Errorf("guardado = %+v, queremos %+v", got, want)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	_, repo, err := runWith(sampleAssets, sums, "4\n2026\n1\n1100,50\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].ResultUSD != 1100.50 {
		t.Errorf("ResultUSD = %v, esperabamos 1100.50", repo.saved)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	out, repo, err := runWith(sampleAssets, sums, "4\n2026\n99\n1\n1100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
	if len(repo.saved) != 1 {
		t.Errorf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
}

func TestRun_RecoversFromInvalidSelection(t *testing.T) {
	sums := summariesForHolding(2026, 5, map[int64]float64{10: 1000, 11: 2000})
	_, repo, err := runWith(sampleAssets, sums, "5\n2026\n99\n2\n2200\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].AssetID != 11 {
		t.Errorf("expected save sobre id=11, got %+v", repo.saved)
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, nil, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén a suxestión:\n%s", out)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}

func TestRun_ShowsGain(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	out, _, err := runWith(sampleAssets, sums, "4\n2026\n1\n1100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+100.00 USD") {
		t.Errorf("saída non mostra a gañanza absoluta:\n%s", out)
	}
	if !strings.Contains(out, "+10.00%") {
		t.Errorf("saída non mostra a gañanza %%:\n%s", out)
	}
}

func TestRun_ShowsLoss(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	out, _, err := runWith(sampleAssets, sums, "4\n2026\n1\n900\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-100.00 USD") {
		t.Errorf("saída non mostra a perda absoluta:\n%s", out)
	}
	if !strings.Contains(out, "-10.00%") {
		t.Errorf("saída non mostra a perda %%:\n%s", out)
	}
}

func TestRun_ShowsBreakeven(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	out, _, err := runWith(sampleAssets, sums, "4\n2026\n1\n1000\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+0.00 USD") {
		t.Errorf("saída non mostra 0.00 USD:\n%s", out)
	}
	if !strings.Contains(out, "+0.00%") {
		t.Errorf("saída non mostra 0.00%%:\n%s", out)
	}
}

func TestRun_UsesPrevResultPlusInvested(t *testing.T) {
	// Resultado anterior 1100 + investido este mes 200 = no activo 1300.
	// Gañanza con resultado actual 1500: 1500-1300 = 200 (+15.38%).
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1200,
			InvestedInMonth:   200,
			EstimatedHolding:  1300,
			HasPrevResult:     true,
		},
	}
	out, _, err := runWith(sampleAssets, sums, "4\n2026\n1\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "no activo: 1300.00 USD") {
		t.Errorf("saída non mostra holding (prev result + invested):\n%s", out)
	}
	if !strings.Contains(out, "+200.00 USD") {
		t.Errorf("saída non mostra gañanza calculada contra holding:\n%s", out)
	}
	if !strings.Contains(out, "+15.38%") {
		t.Errorf("saída non mostra %% calculada contra holding:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	sums := summariesForHolding(2026, 4, map[int64]float64{10: 1000})
	_, repo, err := runWith(sampleAssets, sums, "4\n2026\n1\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}
