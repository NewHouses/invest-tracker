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

type fakeRepo struct {
	assets  []domain.Asset
	totals  map[int64]float64
	saved   []domain.MonthlyResult
	listEr  error
	totalEr error
	saveEr  error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) TotalInvested(id int64) (float64, error) {
	if f.totalEr != nil {
		return 0, f.totalEr
	}
	return f.totals[id], nil
}

func (f *fakeRepo) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	if f.saveEr != nil {
		return 0, f.saveEr
	}
	f.saved = append(f.saved, m)
	return int64(len(f.saved)), nil
}

func runWith(assets []domain.Asset, totals map[int64]float64, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, totals: totals}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addresult.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard S&P 500"},
}

var sampleTotals = map[int64]float64{
	10: 1000,
	11: 2000,
}

const validInput = "1\n1100\n4\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTotals, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir resultado mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTotals, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"Investimentos:",
		"[1] Acción — AAPL",
		"[2] Índice — Vanguard S&P 500",
	}
	for _, e := range expects {
		if !strings.Contains(out, e) {
			t.Errorf("saída non contén %q:\n%s", e, out)
		}
	}
}

func TestRun_PrintsTotalInvested(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTotals, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "investido: 1000.00 USD") {
		t.Errorf("saída non contén o total previo:\n%s", out)
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTotals, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"Selecciona (1-2):",
		"Resultado (USD):",
		"Mes (1-12):",
		"Ano:",
	}
	for _, e := range expects {
		if !strings.Contains(out, e) {
			t.Errorf("saída non contén %q:\n%s", e, out)
		}
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleTotals, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "✓ Resultado gardado") {
		t.Errorf("saída non contén a confirmación:\n%s", out)
	}
	if !strings.Contains(out, "Ganhanzas/Perdas:") {
		t.Errorf("saída non contén Ganhanzas/Perdas:\n%s", out)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
}

func TestRun_HappyPath_SavesCorrectResult(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTotals, "2\n2200\n5\n2026\n")
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
	_, repo, err := runWith(sampleAssets, sampleTotals, "1\n1100,50\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].ResultUSD != 1100.50 {
		t.Errorf("ResultUSD = %v, esperabamos 1100.50", repo.saved)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleTotals, "99\n1\n1100\n4\n2026\n")
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
	_, repo, err := runWith(sampleAssets, sampleTotals, "99\n2\n2200\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].AssetID != 11 {
		t.Errorf("expected save sobre id=11, got %+v", repo.saved)
	}
}

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, nil, "")
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
	out, _, err := runWith(sampleAssets, sampleTotals, "1\n1100\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+100.00 USD") {
		t.Errorf("saída non mostra a ganhanza absoluta:\n%s", out)
	}
	if !strings.Contains(out, "+10.00%") {
		t.Errorf("saída non mostra a ganhanza %%:\n%s", out)
	}
}

func TestRun_ShowsLoss(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTotals, "1\n900\n4\n2026\n")
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
	out, _, err := runWith(sampleAssets, sampleTotals, "1\n1000\n4\n2026\n")
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

func TestRun_HandlesZeroInvested(t *testing.T) {
	totals := map[int64]float64{10: 0, 11: 0}
	out, _, err := runWith(sampleAssets, totals, "1\n100\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "n/a%") {
		t.Errorf("saída non mostra n/a%% para total=0:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTotals, "1\n1100\n4\n")
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
