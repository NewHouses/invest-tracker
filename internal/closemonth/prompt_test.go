package closemonth_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/closemonth"
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
	err := closemonth.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var threeAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
	{ID: 12, Type: domain.Fondo, Name: "Bond Fund"},
}

// summariesFor constrúe summaries onde EstimatedHolding == valor proporcionado
// (caso típico: sen prev result, holding == total invested up to).
func summariesFor(year, month int, holdings map[int64]float64) map[summaryKey]domain.MonthlySummary {
	out := make(map[summaryKey]domain.MonthlySummary)
	for id, h := range holdings {
		out[summaryKey{id, year, month}] = domain.MonthlySummary{
			TotalInvestedUpTo: h,
			EstimatedHolding:  h,
		}
	}
	return out
}

func TestRun_PrintsHeader(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Pechar mes (resultados)") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsMonthAndYear(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_NoEligibleAssets_PrintsHint(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{}) // nada > 0
	out, repo, err := runWith(threeAssets, sums, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos con capital investido para 04/2026") {
		t.Errorf("saída non contén a suxestión:\n%s", out)
	}
	if len(repo.saved) != 0 {
		t.Errorf("saved debería estar baleiro, got %v", repo.saved)
	}
}

func TestRun_PrintsAssetCount(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n2100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "2 activo(s) en 04/2026") {
		t.Errorf("saída non contén o contador:\n%s", out)
	}
}

func TestRun_PromptsForEachEligibleAsset(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n2100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"[1/2] Acción — AAPL (no activo: 1500.00 USD)",
		"[2/2] Índice — Vanguard (no activo: 2000.00 USD)",
		"Resultado (USD, baleiro = saltar):",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HappyPath_SavesOneResultPerAsset(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	_, repo, err := runWith(threeAssets, sums, "4\n2026\n1800\n2100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved tamaño = %d, esperabamos 2", len(repo.saved))
	}
	want := []domain.MonthlyResult{
		{AssetID: 10, ResultUSD: 1800, Month: 4, Year: 2026},
		{AssetID: 11, ResultUSD: 2100, Month: 4, Year: 2026},
	}
	for i, w := range want {
		if repo.saved[i] != w {
			t.Errorf("saved[%d] = %+v, queremos %+v", i, repo.saved[i], w)
		}
	}
}

func TestRun_SkipsEmptyInput(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	out, repo, err := runWith(threeAssets, sums, "4\n2026\n\n2100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].AssetID != 11 {
		t.Errorf("saved = %+v, esperabamos só id=11", repo.saved)
	}
	if !strings.Contains(out, "Saltado") {
		t.Errorf("saída non contén Saltado:\n%s", out)
	}
}

func TestRun_PrintsSummary(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n\n2100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Pechouse 04/2026: 1 resultado(s) gardado(s), 1 saltado(s)") {
		t.Errorf("saída non contén o resumo:\n%s", out)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500})
	_, repo, err := runWith(threeAssets, sums, "4\n2026\n1800,50\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].ResultUSD != 1800.50 {
		t.Errorf("saved = %+v, esperabamos 1800.50", repo.saved)
	}
}

func TestRun_RejectsInvalidResult(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500})
	out, repo, err := runWith(threeAssets, sums, "4\n2026\nabc\n0\n1800\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Resultado non válido") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].ResultUSD != 1800 {
		t.Errorf("saved = %+v, esperabamos 1800", repo.saved)
	}
}

func TestRun_ShowsGainAfterEachSave(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500})
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "+300.00 USD") || !strings.Contains(out, "+20.00%") {
		t.Errorf("saída non mostra gañanza:\n%s", out)
	}
}

func TestRun_ShowsPreviousResultHint(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1500,
			EstimatedHolding:  1500,
			Result:            1700,
			HasResult:         true,
		},
	}
	out, _, err := runWith(threeAssets, sums, "4\n2026\n1800\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Xa hai un resultado rexistrado este mes: 1700.00 USD") {
		t.Errorf("saída non contén o aviso de resultado previo:\n%s", out)
	}
}

func TestRun_UsesPrevResultPlusInvestedAsHolding(t *testing.T) {
	// Resultado anterior 1100 + investido este mes 200 = no activo 1300.
	// Resultado actual 1500 → gañanza 200, +15.38%.
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1200,
			InvestedInMonth:   200,
			EstimatedHolding:  1300,
			HasPrevResult:     true,
		},
	}
	out, repo, err := runWith(threeAssets, sums, "4\n2026\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
	for _, want := range []string{
		"no activo: 1300.00 USD",
		"+200.00 USD",
		"+15.38%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	sums := summariesFor(2026, 4, map[int64]float64{10: 1500, 11: 2000})
	_, repo, err := runWith(threeAssets, sums, "4\n2026\n1800\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada (falta resultado para 2º activo)")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.saved) != 1 {
		t.Errorf("saved tamaño = %d, esperabamos 1 (o primeiro gardado antes do EOF)", len(repo.saved))
	}
}
