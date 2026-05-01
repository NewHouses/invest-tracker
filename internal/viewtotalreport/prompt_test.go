package viewtotalreport_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalreport"
)

type sumKey struct {
	id    int64
	year  int
	month int
}

type divKey struct {
	year, month int
}

type fakeRepo struct {
	assets    []domain.Asset
	summaries map[sumKey]domain.MonthlySummary
	dividends map[divKey]float64
	months    map[divKey][]domain.YearMonth // por (year, month) target → meses con resultados
	listEr    error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) MonthlySummary(id int64, year, month int) (domain.MonthlySummary, error) {
	return f.summaries[sumKey{id, year, month}], nil
}

func (f *fakeRepo) SumDividends(year, month int) (float64, error) {
	return f.dividends[divKey{year, month}], nil
}

func (f *fakeRepo) MonthsWithResultsUpTo(year, month int) ([]domain.YearMonth, error) {
	return f.months[divKey{year, month}], nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewtotalreport.Run(r, &buf, repo)
	return buf.String(), err
}

var twoAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

// Setup: 2 activos. Mes target 04/2026. Mes anterior 03/2026 ten resultados.
//
// 03/2026: AAPL=1100 (compra=1000, +100 gañanza); Vanguard=2100 (compra=2000, +100).
//          Sumas previas: invested=3000, result=3200, dividends=20.
// 04/2026: AAPL invested este mes=200 holding=1100+200=1300, result=1500.
//          Vanguard invested este mes=0 holding=2100+0=2100, result=2150.
//          Sumas: invested ata=3200, este mes=200, result=3650, dividends=50.
func gainSetup() *fakeRepo {
	summaries := map[sumKey]domain.MonthlySummary{
		// 03/2026
		{10, 2026, 3}: {
			TotalInvestedUpTo: 1000, InvestedInMonth: 1000,
			EstimatedHolding: 1000, Result: 1100, HasResult: true,
		},
		{11, 2026, 3}: {
			TotalInvestedUpTo: 2000, InvestedInMonth: 2000,
			EstimatedHolding: 2000, Result: 2100, HasResult: true,
		},
		// 04/2026
		{10, 2026, 4}: {
			TotalInvestedUpTo: 1200, InvestedInMonth: 200,
			EstimatedHolding: 1300, Result: 1500, HasResult: true, HasPrevResult: true,
		},
		{11, 2026, 4}: {
			TotalInvestedUpTo: 2000, InvestedInMonth: 0,
			EstimatedHolding: 2100, Result: 2150, HasResult: true, HasPrevResult: true,
		},
	}
	return &fakeRepo{
		assets:    twoAssets,
		summaries: summaries,
		dividends: map[divKey]float64{
			{2026, 3}: 20,
			{2026, 4}: 50,
		},
		months: map[divKey][]domain.YearMonth{
			{2026, 4}: {{Year: 2026, Month: 3}, {Year: 2026, Month: 4}},
		},
	}
}

const validInput = "4\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Informe mensual total") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetListBeforeMonthYear(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Activos:",
		"- Acción — AAPL",
		"- Índice — Vanguard",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "Activos:") > strings.Index(out, "Mes (1-12):") {
		t.Errorf("a lista de activos debería mostrarse antes do prompt mes/ano:\n%s", out)
	}
}

func TestRun_PromptsMonthAndYear(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	repo := &fakeRepo{}
	out, err := runWith(repo, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_NoActiveInMonth_PrintsHint(t *testing.T) {
	// Pool con activos pero ninguén ten holding > 0 en 04/2026
	repo := &fakeRepo{assets: twoAssets, summaries: map[sumKey]domain.MonthlySummary{}}
	out, err := runWith(repo, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non hai activos con capital investido en 04/2026") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_PrintsAllRequiredFieldLabels(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total investido ata o mes",
		"Investido este mes",
		"Investimento + dividendos prev. mes",
		"No activo (sen div)",
		"No activo (con div)",
		"Dividendos este mes",
		"Resultado (sen div)",
		"Resultado total (con div)",
		"Gañanzas/Perdas",
		"Gañanzas/Perdas (con div)",
		"Índice",
		"Índice (con div)",
		"Promedios mensuais",
		"Índice medio mensual (sen div)",
		"Gañanza media mensual (sen div)",
		"Índice medio mensual (con div)",
		"Gañanza media mensual (con div)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_AggregatesCurrentMonthCorrectly(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 04/2026 esperado:
	// totalInvested = 1200 + 2000 = 3200
	// investedInMonth = 200 + 0 = 200
	// resultSum = 1500 + 2150 = 3650
	// resultSumPrev = 1100 + 2100 = 3200
	// dividends = 50
	// dividendsPrev = 20
	// HoldingNoDiv = 3200 + 200 = 3400
	// HoldingWithDiv = 3400 + 20 = 3420
	// ResultNoDiv = 3650
	// ResultWithDiv = 3650 + 50 = 3700
	// GainNoDiv = 3650 - 3400 = 250
	// PctNoDiv = 250 / 3400 ≈ 7.35%
	// GainWithDiv = 3700 - 3420 = 280
	// PctWithDiv = 280 / 3420 ≈ 8.19%
	// Investimento + dividendos prev. mes = 200 + 20 = 220
	for _, want := range []string{
		"3200.00 USD", // total invested ata o mes
		"200.00 USD",  // investido este mes
		"220.00 USD",  // investimento + dividendos prev. mes
		"3400.00 USD", // no activo sen div
		"3420.00 USD", // no activo con div
		"50.00 USD",   // dividendos este mes
		"3650.00 USD", // resultado sen div
		"3700.00 USD", // resultado con div
		"+250.00 USD", // gain sen div
		"+280.00 USD", // gain con div
		"+7.35%",      // pct sen div
		"+8.19%",      // pct con div
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAverages(t *testing.T) {
	out, err := runWith(gainSetup(), validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 03/2026: HoldingNoDiv = 0 (resultSumPrev=0) + invested=3000 = 3000
	//          ResultNoDiv = 3200, GainNoDiv = 200, PctNoDiv = 6.67%
	//          HoldingWithDiv = 3000 + 0 = 3000 (no prev div)
	//          ResultWithDiv = 3200 + 20 = 3220, GainWithDiv = 220, PctWithDiv ≈ 7.33%
	// 04/2026: como en TestRun_AggregatesCurrentMonthCorrectly
	// avg pct sen div = (6.67 + 7.35) / 2 ≈ 7.01
	// avg gain sen div = (200 + 250) / 2 = 225
	// avg pct con div = (7.33 + 8.19) / 2 ≈ 7.76
	// avg gain con div = (220 + 280) / 2 = 250
	if !strings.Contains(out, "2 mes(es) con resultado") {
		t.Errorf("saída non indica 2 meses no contador:\n%s", out)
	}
	for _, want := range []string{
		"+7.01",  // avg pct no div (aproximado)
		"+225.00 USD",
		"+250.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_NoMonthsWithResults_ShowsZeroAverages(t *testing.T) {
	repo := gainSetup()
	repo.months = map[divKey][]domain.YearMonth{} // sen meses con resultado
	repo.summaries[sumKey{10, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 1200, EstimatedHolding: 1200, // sen result
	}
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 2000, EstimatedHolding: 2000,
	}
	out, err := runWith(repo, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "sen meses con resultados") {
		t.Errorf("saída non sinala falta de medias:\n%s", out)
	}
}

func TestRun_TargetMonthWithoutResults_ShowsDashes(t *testing.T) {
	repo := gainSetup()
	// Eliminar os resultados de 04/2026 pero deixar holding > 0
	repo.summaries[sumKey{10, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 1200, EstimatedHolding: 1300,
	}
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 2000, EstimatedHolding: 2100,
	}
	out, err := runWith(repo, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Count(out, "—") < 4 {
		t.Errorf("saída non contén guións suficientes para Resultado/G&P/Índice:\n%s", out)
	}
}

func TestRun_PartialResults_AnnotatesCoverage(t *testing.T) {
	repo := gainSetup()
	// Vanguard sen resultado para 04/2026
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 2000, EstimatedHolding: 2100,
	}
	out, err := runWith(repo, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "(1/2 activos con resultado)") {
		t.Errorf("saída non anota cobertura parcial:\n%s", out)
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
