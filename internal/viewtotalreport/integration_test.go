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
	//   resultSumPrev (03) = 1100 + 2100 = 3200
	//   div 04 = 50, divPrev (03) = 20
	//   HoldingNoDiv = 3200 + 200 = 3400
	//   HoldingWithDiv = 3400 + 20 = 3420
	//   ResultNoDiv = 3650, ResultWithDiv = 3700
	//   GainNoDiv = 250, PctNoDiv = 250/3400 ≈ 7.35%
	//   GainWithDiv = 280, PctWithDiv = 280/3420 ≈ 8.19%
	//   Investimento + dividendos prev. mes = 200 + 20 = 220
	for _, want := range []string{
		"Informe total · 04/2026",
		"3200.00 USD", // total investido
		"200.00 USD",  // investido este mes
		"220.00 USD",  // invest + div prev
		"3400.00 USD", // no activo sen div
		"3420.00 USD", // no activo con div
		"50.00 USD",   // dividendos este mes
		"3650.00 USD", // resultado sen div
		"3700.00 USD", // resultado con div
		"+250.00 USD",
		"+280.00 USD",
		"+7.35%",
		"+8.19%",
		"2 mes(es) con resultado",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
