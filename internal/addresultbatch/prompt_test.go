package addresultbatch_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addresultbatch"
	"invest-tracker/internal/domain"
)

type sumKey struct {
	id    int64
	year  int
	month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[sumKey]domain.MonthlySummary
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
	return f.summaries[sumKey{id, year, month}], nil
}

func (f *fakeRepo) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	if f.saveEr != nil {
		return 0, f.saveEr
	}
	f.saved = append(f.saved, m)
	return int64(len(f.saved)), nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addresultbatch.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026},
}

// 3 meses con holding > 0
func threeMonthSetup() *fakeRepo {
	return &fakeRepo{
		assets: sampleAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {EstimatedHolding: 1000},
			{10, 2026, 5}: {EstimatedHolding: 1100},
			{10, 2026, 6}: {EstimatedHolding: 1200},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(threeMonthSetup(), "1\n4\n2026\n1100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir resultados mensuais en serie") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
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

func TestRun_PromptsAssetThenStartDate(t *testing.T) {
	out, err := runWith(threeMonthSetup(), "1\n4\n2026\n1100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Investimentos:",
		"[1] Acción — AAPL",
		"Introduce o mes inicial:",
		"Mes (1-12):",
		"Ano:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_RejectsStartDateBeforeAssetCreation(t *testing.T) {
	// Asset 1/2026; intento iniciar en 12/2025; re-prompt; logo 4/2026 OK.
	out, err := runWith(threeMonthSetup(), "1\n12\n2025\n4\n2026\n1100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non pode ser anterior á data do activo (01/2026)") {
		t.Errorf("saída non contén o erro de data anterior:\n%s", out)
	}
}

func TestRun_HappyPath_AdvancesMonths(t *testing.T) {
	// 3 meses (4, 5, 6) con resultados 1100, 1200, 1500. Logo "n" para parar.
	_, err := runWith(threeMonthSetup(),
		"1\n4\n2026\n1100\ns\n1200\ns\n1500\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRun_SavesEachMonthCorrectly(t *testing.T) {
	repo := threeMonthSetup()
	_, err := runWith(repo, "1\n4\n2026\n1100\ns\n1200\ns\n1500\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved tamaño = %d, esperabamos 3", len(repo.saved))
	}
	want := []domain.MonthlyResult{
		{AssetID: 10, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: 10, ResultUSD: 1200, Month: 5, Year: 2026},
		{AssetID: 10, ResultUSD: 1500, Month: 6, Year: 2026},
	}
	for i, w := range want {
		if repo.saved[i] != w {
			t.Errorf("saved[%d] = %+v, esperabamos %+v", i, repo.saved[i], w)
		}
	}
}

func TestRun_ShowsHoldingPerMonth(t *testing.T) {
	out, err := runWith(threeMonthSetup(),
		"1\n4\n2026\n1100\ns\n1200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"--- 04/2026 ---",
		"No activo: 1000.00 USD",
		"--- 05/2026 ---",
		"No activo: 1100.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_ShowsGainAfterEachSave(t *testing.T) {
	out, err := runWith(threeMonthSetup(), "1\n4\n2026\n1100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 04/2026: holding=1000, result=1100 → gain=+100, pct=+10%
	if !strings.Contains(out, "+100.00 USD") || !strings.Contains(out, "+10.00%") {
		t.Errorf("saída non mostra ganhanza:\n%s", out)
	}
}

func TestRun_SkipsOnEmptyInput(t *testing.T) {
	repo := threeMonthSetup()
	// 4/2026: baleiro (saltar) → continuar → 5/2026: 1200 → parar
	_, err := runWith(repo, "1\n4\n2026\n\ns\n1200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Errorf("saved = %d, esperabamos 1 (un saltado)", len(repo.saved))
	}
	if repo.saved[0].Month != 5 {
		t.Errorf("saved[0] mes = %d, esperabamos 5", repo.saved[0].Month)
	}
}

func TestRun_AutoSkipsMonthsWithoutHolding(t *testing.T) {
	// Mes 4 ten holding=0 (será saltado automaticamente). Mes 5 holding > 0.
	repo := &fakeRepo{
		assets: sampleAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {EstimatedHolding: 0},
			{10, 2026, 5}: {EstimatedHolding: 1100},
		},
	}
	_, err := runWith(repo, "1\n4\n2026\ns\n1200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved = %d, esperabamos 1 (mes 4 auto-saltado)", len(repo.saved))
	}
	if repo.saved[0].Month != 5 {
		t.Errorf("saved[0] mes = %d, esperabamos 5", repo.saved[0].Month)
	}
}

func TestRun_ContinueDefaultIsYes(t *testing.T) {
	repo := threeMonthSetup()
	// "" significa sí → continúa; ao final non.
	_, err := runWith(repo, "1\n4\n2026\n1100\n\n1200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved = %d, esperabamos 2 (baleiro = sí)", len(repo.saved))
	}
}

func TestRun_RolloverYear(t *testing.T) {
	// asset 10/2025; iniciar en 11/2025; saltarmos a 12/2025 e logo 01/2026.
	earlyAsset := []domain.Asset{
		{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 11, Year: 2025},
	}
	repo := &fakeRepo{
		assets: earlyAsset,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2025, 11}: {EstimatedHolding: 1000},
			{10, 2025, 12}: {EstimatedHolding: 1100},
			{10, 2026, 1}:  {EstimatedHolding: 1200},
		},
	}
	_, err := runWith(repo, "1\n11\n2025\n1100\ns\n1200\ns\n1300\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved = %d, esperabamos 3", len(repo.saved))
	}
	want := []struct{ y, m int }{{2025, 11}, {2025, 12}, {2026, 1}}
	for i, w := range want {
		if repo.saved[i].Year != w.y || repo.saved[i].Month != w.m {
			t.Errorf("saved[%d] = %02d/%d, esperabamos %02d/%d",
				i, repo.saved[i].Month, repo.saved[i].Year, w.m, w.y)
		}
	}
}

func TestRun_PrintsSummary(t *testing.T) {
	out, err := runWith(threeMonthSetup(),
		"1\n4\n2026\n1100\ns\n\ns\n1200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 1 gardado en 04, saltado en 05, 1 gardado en 06 → 2 gardados, 1 saltado
	if !strings.Contains(out, "2 resultado(s) e saltáronse 1 sobre Acción — AAPL") {
		t.Errorf("saída non contén o resumo:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, err := runWith(threeMonthSetup(), "1\n4\n2026\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
