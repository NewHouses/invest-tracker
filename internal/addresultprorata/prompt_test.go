package addresultprorata_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"invest-tracker/internal/addresultprorata"
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
	err := addresultprorata.Run(r, &buf, repo)
	return buf.String(), err
}

// 3 acciones + 1 indice. AAPL holding=1000, MSFT holding=500, GOOG holding=2500.
// Total acción = 4000.
func gainSetup() *fakeRepo {
	return &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
			{ID: 12, Type: domain.Indice, Name: "Vanguard"},
			{ID: 13, Type: domain.Accion, Name: "GOOG"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {EstimatedHolding: 1000},
			{11, 2026, 5}: {EstimatedHolding: 500},
			{12, 2026, 5}: {EstimatedHolding: 9999}, // Vanguard non debe entrar (non é Acción)
			{13, 2026, 5}: {EstimatedHolding: 2500},
		},
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n5\n2026\n400\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir resultados proporcionais por tipo") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PromptsTypeBeforeMonth(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n5\n2026\n400\n")
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

func TestRun_PrintsEligibleListAndTotal(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n5\n2026\n400\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Activos de tipo Acción en 05/2026:",
		"AAPL · no activo: 1000.00 USD",
		"MSFT · no activo: 500.00 USD",
		"GOOG · no activo: 2500.00 USD",
		"Suma de holdings: 4000.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
	// Vanguard (Indice) non debe aparecer
	if strings.Contains(out, "Vanguard") {
		t.Errorf("saída non debería incluír Vanguard:\n%s", out)
	}
}

func TestRun_DistributesGainProportionally(t *testing.T) {
	repo := gainSetup()
	// Total gain = 400 sobre holding total 4000 → 10% para cada un.
	// AAPL: 1000 → 1100 (+100)
	// MSFT: 500 → 550 (+50)
	// GOOG: 2500 → 2750 (+250)
	_, err := runWith(repo, "1\n5\n2026\n400\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved tamaño = %d, esperabamos 3", len(repo.saved))
	}
	want := map[int64]float64{
		10: 1100, // AAPL
		11: 550,  // MSFT
		13: 2750, // GOOG
	}
	for _, m := range repo.saved {
		exp, ok := want[m.AssetID]
		if !ok {
			t.Errorf("non esperabamos save sobre id=%d: %+v", m.AssetID, m)
			continue
		}
		if !almostEqual(m.ResultUSD, exp) {
			t.Errorf("asset id=%d: result %v, esperabamos %v", m.AssetID, m.ResultUSD, exp)
		}
		if m.Month != 5 || m.Year != 2026 {
			t.Errorf("data incorrecta: %+v", m)
		}
	}
}

func TestRun_DistributesLossProportionally(t *testing.T) {
	repo := gainSetup()
	// Perda total = -200. Distribución -5%.
	// AAPL: 1000 → 950
	// MSFT: 500 → 475
	// GOOG: 2500 → 2375
	_, err := runWith(repo, "1\n5\n2026\n-200\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := map[int64]float64{10: 950, 11: 475, 13: 2375}
	for _, m := range repo.saved {
		if !almostEqual(m.ResultUSD, want[m.AssetID]) {
			t.Errorf("id=%d result %v, esperabamos %v", m.AssetID, m.ResultUSD, want[m.AssetID])
		}
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	repo := gainSetup()
	// 200,50 = 200.50 → +5.0125% → AAPL 1050.125, MSFT 525.0625, GOOG 2625.3125
	_, err := runWith(repo, "1\n5\n2026\n200,50\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, m := range repo.saved {
		if m.AssetID == 10 && !almostEqual(m.ResultUSD, 1050.125) {
			t.Errorf("AAPL result %v, esperabamos 1050.125", m.ResultUSD)
		}
	}
}

func TestRun_PrintsDistributionSummary(t *testing.T) {
	out, err := runWith(gainSetup(), "1\n5\n2026\n400\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Distribución proporcional:",
		"AAPL: 1000.00 → 1100.00 USD",
		"+10.00%",
		"3 resultado(s) e saltáronse 0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_ExcludesAssetsWithZeroHolding(t *testing.T) {
	// 3 acciones: AAPL=1000, MSFT=0 (excluído), GOOG=500.
	// Total elixible = 1500. Ganhanza 300 → +20% para os elixibles.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
			{ID: 11, Type: domain.Accion, Name: "MSFT"},
			{ID: 12, Type: domain.Accion, Name: "GOOG"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {EstimatedHolding: 1000},
			{11, 2026, 5}: {EstimatedHolding: 0}, // excluído
			{12, 2026, 5}: {EstimatedHolding: 500},
		},
	}
	out, err := runWith(repo, "1\n5\n2026\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// MSFT debe aparecer como excluído na lista pero non gardarse.
	if !strings.Contains(out, "MSFT · no activo: 0.00 USD (excluído: non conta para a repartición)") {
		t.Errorf("saída non sinala MSFT como excluído:\n%s", out)
	}

	// Suma de holdings é 1500 (NON 1500+0).
	if !strings.Contains(out, "Suma de holdings: 1500.00 USD") {
		t.Errorf("suma de holdings non é 1500:\n%s", out)
	}

	// Só AAPL e GOOG reciben resultados; MSFT non.
	if len(repo.saved) != 2 {
		t.Fatalf("saved tamaño = %d, esperabamos 2 (sen MSFT)", len(repo.saved))
	}
	for _, m := range repo.saved {
		if m.AssetID == 11 {
			t.Errorf("MSFT (id=11) non debería gardarse: %+v", m)
		}
	}

	// Distribución +20% para os elixibles.
	want := map[int64]float64{10: 1200, 12: 600}
	for _, m := range repo.saved {
		if !almostEqual(m.ResultUSD, want[m.AssetID]) {
			t.Errorf("id=%d result %v, esperabamos %v", m.AssetID, m.ResultUSD, want[m.AssetID])
		}
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

func TestRun_NoActiveInMonth_PrintsHint(t *testing.T) {
	// Acción existen pero ninguén ten holding > 0 en 04/2026.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
		},
		summaries: map[sumKey]domain.MonthlySummary{}, // sen entradas → 0
	}
	out, err := runWith(repo, "1\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Acción con capital en 04/2026.") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_SkipsAssetWithNonPositiveResult(t *testing.T) {
	// Holding total 1000. Perda 1000 → 0 → marcado como saltado.
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL"},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {EstimatedHolding: 1000},
		},
	}
	out, err := runWith(repo, "1\n5\n2026\n-1000\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("saved debería estar baleiro (resultado=0), got %v", repo.saved)
	}
	if !strings.Contains(out, "Saltado") {
		t.Errorf("saída non sinala Saltado:\n%s", out)
	}
	if !strings.Contains(out, "saltáronse 1") {
		t.Errorf("contador de saltados incorrecto:\n%s", out)
	}
}

func TestRun_ZeroGain_AppliesIdentity(t *testing.T) {
	// Ganhanza 0 → cada activo recibe o seu propio holding como resultado.
	repo := gainSetup()
	_, err := runWith(repo, "1\n5\n2026\n0\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved tamaño = %d, esperabamos 3", len(repo.saved))
	}
	want := map[int64]float64{10: 1000, 11: 500, 13: 2500}
	for _, m := range repo.saved {
		if !almostEqual(m.ResultUSD, want[m.AssetID]) {
			t.Errorf("id=%d result %v, esperabamos %v", m.AssetID, m.ResultUSD, want[m.AssetID])
		}
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, err := runWith(gainSetup(), "1\n5\n2026\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_RejectsInvalidGainInput(t *testing.T) {
	repo := gainSetup()
	out, err := runWith(repo, "1\n5\n2026\nabc\n400\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Valor non válido") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.saved) != 3 {
		t.Errorf("saved = %d, esperabamos 3 tras recuperación", len(repo.saved))
	}
}
