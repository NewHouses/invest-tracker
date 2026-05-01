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

	// 2 activos, 2 meses con resultados, 2 dividendos.
	//
	// 03/2026: AAPL 1000 (compra inicial) + Vanguard 2000 (compra inicial).
	//          Resultados: AAPL=1100, Vanguard=2100. Dividend=20.
	//          → agg: aporte=(3000−20)=2980, fondos=3000, base=3200, div=20,
	//                 result=3220, +220, +7.33%
	// 04/2026: AAPL +200 tx. Resultados: AAPL=1500, Vanguard=2150. Dividend=50.
	//          → agg: aporte=(200−50)=150, fondos=3400, base=3650, div=50,
	//                 result=3700, +300, +8.82%
	//
	// Lifetime: aporte=3200, totalDiv=70, lastResults=1500+2150=3650.
	// G/P Total = 3650 + 70 - 3200 = +520. Avg pct ≈ +8.08, Avg gain = 260.
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

	for _, want := range []string{
		"Reporte histórico completo",
		"2 activo(s) · 2 mes(es) con resultado",
		// Top summary
		"Aporte histórico total",
		"3200.00 USD",
		"Índice Medio",
		"+8.08%",
		"G/P Media",
		"+260.00 USD",
		"G/P Total",
		"+520.00 USD",
		"Dividendos totais",
		"70.00 USD",
		// Table columns
		"Aporte Mensual",
		"Fondos",
		"Dividendos",
		"Resultado",
		// Row 03 — aporte=2980 (3000−20), fondos=3000
		"2980.00", "3000.00", "+7.33%", "+220.00", "20.00", "3220.00",
		// Row 04 — aporte=150 (200−50), fondos=3400
		"150.00", "3400.00", "+8.82%", "+300.00", "50.00", "3700.00",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
