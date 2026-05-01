package addresultproratabatch_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"invest-tracker/internal/addresultproratabatch"
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
	err := addresultproratabatch.Run(r, &buf, repo)
	return buf.String(), err
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

// 2 acciones (AAPL=1000, MSFT=500). Holdings constantes en 04, 05 e 06 para
// simplicidade do test (na realidade dependerían dos prev results, pero para
// o test limítase usar holdings fixos).
func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {EstimatedHolding: 1000},
			{11, 2026, 4}: {EstimatedHolding: 500},
			{10, 2026, 5}: {EstimatedHolding: 1100},
			{11, 2026, 5}: {EstimatedHolding: 550},
			{10, 2026, 6}: {EstimatedHolding: 1200},
			{11, 2026, 6}: {EstimatedHolding: 600},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n4\n2026\n150\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir resultados proporcionais por tipo en serie") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PromptsTypeBeforeMonth(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n4\n2026\n150\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	idxType := strings.Index(out, "Tipo de investimento:")
	idxMes := strings.Index(out, "Mes (1-12):")
	if !(idxType >= 0 && idxMes > idxType) {
		t.Errorf("orde esperada: tipo antes de mes; got tipo=%d mes=%d\n%s",
			idxType, idxMes, out)
	}
}

func TestRun_NoAssetsOfType_PrintsHint(t *testing.T) {
	repo := gainSetup()
	out, err := runWith(repo, "4\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Fondo.") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_DistributesAcrossMultipleMonths(t *testing.T) {
	repo := gainSetup()
	// 04: total holding 1500, gain 150 → +10% (AAPL: 1100, MSFT: 550). Continuar.
	// 05: total 1650, gain 165 → +10% (AAPL: 1210, MSFT: 605). Continuar.
	// 06: total 1800, gain -180 → -10% (AAPL: 1080, MSFT: 540). Parar.
	_, err := runWith(repo, "1\n4\n2026\n150\ns\n165\ns\n-180\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 6 {
		t.Fatalf("saved tamaño = %d, esperabamos 6", len(repo.saved))
	}
	// Verificacións simples: cada mes ten 2 entradas e os valores son correctos.
	want := []struct {
		assetID int64
		month   int
		amount  float64
	}{
		{10, 4, 1100},
		{11, 4, 550},
		{10, 5, 1210},
		{11, 5, 605},
		{10, 6, 1080},
		{11, 6, 540},
	}
	for i, w := range want {
		m := repo.saved[i]
		if m.AssetID != w.assetID || m.Month != w.month || !almostEqual(m.ResultUSD, w.amount) {
			t.Errorf("saved[%d] = %+v, esperabamos %+v", i, m, w)
		}
	}
}

func TestRun_AutoSkipsMonthsWithoutEligible(t *testing.T) {
	// 04 sen activos elixibles, 05 con activos. Continuar tras 04.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {EstimatedHolding: 0},
			{10, 2026, 5}: {EstimatedHolding: 1000},
		},
	}
	out, err := runWith(repo, "1\n4\n2026\ns\n100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Sen activos elixibles neste mes. Saltado.") {
		t.Errorf("saída non auto-salta o mes baleiro:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].Month != 5 {
		t.Errorf("esperabamos 1 save no mes 5, got %+v", repo.saved)
	}
}

func TestRun_SkipsMonthOnEmptyInput(t *testing.T) {
	repo := gainSetup()
	// 04: baleiro (saltado). Continuar. 05: 165 → garda. Parar.
	_, err := runWith(repo, "1\n4\n2026\n\ns\n165\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved = %d, esperabamos 2 (só 05 gardou)", len(repo.saved))
	}
	for _, m := range repo.saved {
		if m.Month != 5 {
			t.Errorf("saved[mes=%d] non debería existir", m.Month)
		}
	}
}

func TestRun_ContinueDefaultIsYes(t *testing.T) {
	repo := gainSetup()
	// Resposta baleira no continuar → segue.
	_, err := runWith(repo, "1\n4\n2026\n150\n\n165\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 4 {
		t.Errorf("saved = %d, esperabamos 4 (2 meses)", len(repo.saved))
	}
}

func TestRun_RolloverYear(t *testing.T) {
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2025, 11}: {EstimatedHolding: 1000},
			{10, 2025, 12}: {EstimatedHolding: 1100},
			{10, 2026, 1}:  {EstimatedHolding: 1200},
		},
	}
	_, err := runWith(repo, "1\n11\n2025\n100\ns\n100\ns\n100\nn\n")
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
	repo := gainSetup()
	// 04 garda 2, 05 saltado, 06 garda 2 → 4 gardados, 1 saltado.
	out, err := runWith(repo, "1\n4\n2026\n150\ns\n\ns\n180\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "4 resultado(s) e saltáronse 1 mes(es)") {
		t.Errorf("saída non contén o resumo:\n%s", out)
	}
}

func TestRun_ExcludesAssetsWithZeroHolding(t *testing.T) {
	// Mes 04: AAPL=1000, MSFT=0 (excluído), GOOG=500. Total=1500.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
			{ID: 12, Type: domain.Accion, Name: "GOOG"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 4}: {EstimatedHolding: 1000},
			{11, 2026, 4}: {EstimatedHolding: 0},
			{12, 2026, 4}: {EstimatedHolding: 500},
		},
	}
	out, err := runWith(repo, "1\n4\n2026\n300\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "MSFT · no activo: 0.00 USD (excluído)") {
		t.Errorf("saída non sinala MSFT como excluído:\n%s", out)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved = %d, esperabamos 2 (sen MSFT)", len(repo.saved))
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, err := runWith(gainSetup(), "1\n4\n2026\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_RejectsInvalidGainInput(t *testing.T) {
	repo := gainSetup()
	out, err := runWith(repo, "1\n4\n2026\nabc\n150\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Valor non válido") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved = %d, esperabamos 2 tras recuperación", len(repo.saved))
	}
}
