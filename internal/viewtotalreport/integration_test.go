package viewtotalreport_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtotalreport"
)

func TestRun_EndToEnd_TotalReportFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// 2 activos compradoss en 03/2026
	aaplID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset AAPL: %v", err)
	}
	vanID, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	// Resultados en 03/2026: AAPL=1100, Vanguard=2100
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: vanID, ResultUSD: 2100, Month: 3, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult 03/2026: %v", err)
		}
	}

	// Compras adicionais en 04/2026 só para AAPL
	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: aaplID, AmountUSD: 200, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	// Resultados en 04/2026
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1500, Month: 4, Year: 2026},
		{AssetID: vanID, ResultUSD: 2150, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult 04/2026: %v", err)
		}
	}

	// Dividendos: 20 en 03/2026, 50 en 04/2026
	for _, d := range []domain.Dividend{
		{AmountUSD: 20, Month: 3, Year: 2026},
		{AmountUSD: 50, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertDividend(d); err != nil {
			t.Fatalf("InsertDividend: %v", err)
		}
	}

	r := bufio.NewReader(strings.NewReader("4\n2026\n"))
	var out bytes.Buffer
	if err := viewtotalreport.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()

	// Esperados (recapitulando os cálculos):
	//   totalInvested ata 04 = 1200 + 2000 = 3200
	//   investedInMonth 04 = 200 + 0 = 200
	//   resultSum 04 = 1500 + 2150 = 3650
	//   holding 04 = Σ EstimatedHolding = (1100 + 200) + 2100 = 3400
	//   div 04 = 50, divPrev (03) = 20
	//   HoldingNoDiv = 3400
	//   HoldingWithDiv = 3400 + 20 = 3420
	//   ResultNoDiv = 3650, ResultWithDiv = 3700
	//   GainNoDiv = 250, PctNoDiv = 250/3400 ≈ 7.35%
	//   GainWithDiv = 280, PctWithDiv = 280/3420 ≈ 8.19%
	//   Investimento + dividendos prev. mes = 200 + 20 = 220
	for _, want := range []string{
		"Informe total · 04/2026",
		"$3200.00", // total investido
		"$200.00",  // investido este mes
		"$220.00",  // invest + div prev
		"$3400.00", // no activo sen div
		"$3420.00", // no activo con div
		"$50.00",   // dividendos este mes
		"$3650.00", // resultado sen div
		"$3700.00", // resultado con div
		"+$250.00",
		"+$280.00",
		"+7.35%",
		"+8.19%",
		"2 mes(es) con resultado",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}

// Un activo sen resultado nun mes non debe inflar nin desinflar a G/P dos
// meses seguintes: o informe debe coincidir co reporte histórico completo.
//
//	A=1000 e B=500 creados en 01/2026.
//	Resultados: 01 → A=1000, B=500 · 02 → só B=500 · 03 → A=1100, B=510.
func TestRun_EndToEnd_AssetSkipsMonth(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	aID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "A", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset A: %v", err)
	}
	bID, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "B", AmountUSD: 500, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset B: %v", err)
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aID, ResultUSD: 1000, Month: 1, Year: 2026},
		{AssetID: bID, ResultUSD: 500, Month: 1, Year: 2026},
		{AssetID: bID, ResultUSD: 500, Month: 2, Year: 2026},
		{AssetID: aID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: bID, ResultUSD: 510, Month: 3, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	r := bufio.NewReader(strings.NewReader("3\n2026\n"))
	var out bytes.Buffer
	if err := viewtotalreport.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	output := out.String()

	// 03/2026: base 1000 + 500 = 1500, resultado 1610 → +110 (+7.33%).
	// Medias (01: 0%, 02: 0%, 03: +7.33%) → +2.44% e +$36.67.
	for _, want := range []string{
		"+$110.00",
		"+7.33%",
		"3 mes(es) con resultado",
		"+2.44%",
		"+$36.67",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
	for _, bad := range []string{"+$1110.00", "+222.00%"} {
		if strings.Contains(output, bad) {
			t.Errorf("saída contén %q (G/P inflada):\n%s", bad, output)
		}
	}
}
