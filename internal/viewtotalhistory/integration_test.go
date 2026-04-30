package viewtotalhistory_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtotalhistory"
)

func TestRun_EndToEnd_HistoryFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Setup idéntico ao integration test de viewtotalreport para reutilizar cálculos.
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

	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: vanID, ResultUSD: 2100, Month: 3, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult 03/2026: %v", err)
		}
	}

	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: aaplID, AmountUSD: 200, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1500, Month: 4, Year: 2026},
		{AssetID: vanID, ResultUSD: 2150, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult 04/2026: %v", err)
		}
	}

	for _, d := range []domain.Dividend{
		{AmountUSD: 20, Month: 3, Year: 2026},
		{AmountUSD: 50, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertDividend(d); err != nil {
			t.Fatalf("InsertDividend: %v", err)
		}
	}

	r := bufio.NewReader(strings.NewReader(""))
	var out bytes.Buffer
	if err := viewtotalhistory.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()

	// Cabeceira
	if !strings.Contains(output, "Resultado xeral") {
		t.Errorf("saída non contén cabeceira:\n%s", output)
	}
	if !strings.Contains(output, "2 mes(es) con resultado") {
		t.Errorf("saída non sinala 2 meses:\n%s", output)
	}

	// Lifetime: invested=3200 result_final=3650 gain=450 pct=14.06%
	for _, want := range []string{
		"3200.00 USD", // total investido lifetime
		"+450.00 USD", // total gain
		"+14.06%",     // total pct
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Medias: pct sen div ≈7.01, gain sen div=225, pct con div ≈7.76, gain con div=250
	for _, want := range []string{
		"+7.01",
		"+225.00 USD",
		"+7.76",
		"+250.00 USD",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Filas mensuais
	for _, want := range []string{
		// 03/2026
		"3000.00", // invested total ata 03 + holding
		"+200.00", // gain sen div 03
		"+220.00", // gain con div 03
		"+6.67%",
		"+7.33%",
		// 04/2026
		"3400.00", // holding sen div 04
		"3420.00", // holding con div 04
		"3650.00", // result sen div 04
		"3700.00", // result con div 04
		"+250.00", // gain sen div 04
		"+280.00", // gain con div 04
		"+7.35%",
		"+8.19%",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
