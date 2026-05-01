package viewtypereport_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtypereport"
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
	err := viewtypereport.Run(r, &buf, repo)
	return buf.String(), err
}

// 3 acciones + 1 indice
var mixedAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Accion, Name: "MSFT"},
	{ID: 12, Type: domain.Indice, Name: "Vanguard"},
	{ID: 13, Type: domain.Accion, Name: "GOOG"},
}

// Input: type=1 (Accion), month=4, year=2026
const validInput = "1\n4\n2026\n"

func gainSummariesAccion() map[summaryKey]domain.MonthlySummary {
	return map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1000, InvestedInMonth: 0,
			EstimatedHolding: 1000, Result: 1100, HasResult: true,
		},
		{11, 2026, 4}: {
			TotalInvestedUpTo: 500, InvestedInMonth: 200,
			EstimatedHolding: 500, Result: 600, HasResult: true,
		},
		{13, 2026, 4}: {
			TotalInvestedUpTo: 800, InvestedInMonth: 0,
			EstimatedHolding: 800, Result: 900, HasResult: true,
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Informe mensual por tipo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsForType(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Tipo de investimento:",
		"[1] Acción",
		"[2] Índice",
		"[3] Copy-trading",
		"[4] Fondo",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAssetListBeforeMonthYear(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Activos de tipo Acción:",
		"- AAPL",
		"- MSFT",
		"- GOOG",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "- Vanguard") {
		t.Errorf("saída non debería listar Vanguard (é Índice):\n%s", out)
	}
	idxList := strings.Index(out, "Activos de tipo Acción:")
	idxMes := strings.Index(out, "Mes (1-12):")
	if !(idxList < idxMes) {
		t.Errorf("a lista debería aparecer antes de pedir o mes; got list=%d mes=%d", idxList, idxMes)
	}
}

func TestRun_PromptsMonthAndYear(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_NoAssetsOfType_PrintsHint(t *testing.T) {
	// Selecciona Fondo (4), pero non hai
	out, err := runWith(mixedAssets, gainSummariesAccion(), "4\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Fondo.") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_NoActiveInMonth_PrintsHint(t *testing.T) {
	// Acción existe pero ningún ten EstimatedHolding > 0 en 04/2026
	out, err := runWith(mixedAssets, map[summaryKey]domain.MonthlySummary{}, "1\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos de tipo Acción con capital investido en 04/2026.") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_PrintsTable_AggregatedFields(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Tipo: Acción · 04/2026",
		"Activos incluídos: 3",
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

func TestRun_AggregatesMetricsCorrectly(t *testing.T) {
	out, err := runWith(mixedAssets, gainSummariesAccion(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// totalInvested = 1000 + 500 + 800 = 2300
	// investedInMonth = 0 + 200 + 0 = 200
	// holding = 1000 + 500 + 800 = 2300
	// resultSum = 1100 + 600 + 900 = 2600
	// gain = 2600 - 2300 = 300
	// pct = 300/2300 * 100 ≈ 13.04
	for _, want := range []string{
		"2300.00 USD",
		"200.00 USD",
		"2600.00 USD",
		"+300.00 USD",
		"+13.04%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_ShowsAggregateLoss(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1000, InvestedInMonth: 0,
			EstimatedHolding: 1000, Result: 800, HasResult: true,
		},
		{11, 2026, 4}: {
			TotalInvestedUpTo: 1000, InvestedInMonth: 0,
			EstimatedHolding: 1000, Result: 900, HasResult: true,
		},
	}
	out, err := runWith(mixedAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-300.00 USD") {
		t.Errorf("saída non mostra perda absoluta:\n%s", out)
	}
	if !strings.Contains(out, "-15.00%") {
		t.Errorf("saída non mostra perda %%:\n%s", out)
	}
}

func TestRun_PartialResults_ShowsPartialLabel(t *testing.T) {
	// Só 2 dos 3 activos teñen resultado
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1000, EstimatedHolding: 1000,
			Result: 1100, HasResult: true,
		},
		{11, 2026, 4}: {
			TotalInvestedUpTo: 500, EstimatedHolding: 500,
			Result: 600, HasResult: true,
		},
		{13, 2026, 4}: {
			TotalInvestedUpTo: 800, EstimatedHolding: 800,
			Result: 0, HasResult: false,
		},
	}
	out, err := runWith(mixedAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Resultado (parc.)") {
		t.Errorf("saída non marca parcialidade:\n%s", out)
	}
	if !strings.Contains(out, "(2/3 activos)") {
		t.Errorf("saída non mostra cobertura 2/3:\n%s", out)
	}
}

func TestRun_NoneHaveResults_ShowsDashes(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {TotalInvestedUpTo: 1000, EstimatedHolding: 1000},
		{11, 2026, 4}: {TotalInvestedUpTo: 500, EstimatedHolding: 500},
	}
	out, err := runWith(mixedAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Count(out, "—") < 3 {
		t.Errorf("saída non contén guións longos para resultado/G&P/índice:\n%s", out)
	}
}

func TestRun_FiltersByTypeOnly(t *testing.T) {
	// Verifica que só inclúe os Acción, non Vanguard (Indice)
	sums := gainSummariesAccion()
	// Engadir tamén un summary para Vanguard que NON debe aparecer
	sums[summaryKey{12, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 99999, EstimatedHolding: 99999,
		Result: 99999, HasResult: true,
	}
	out, err := runWith(mixedAssets, sums, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(out, "99999") {
		t.Errorf("saída non debería incluír Vanguard:\n%s", out)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, err := runWith(mixedAssets, gainSummariesAccion(), "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
