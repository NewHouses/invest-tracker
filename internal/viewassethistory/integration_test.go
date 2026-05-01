package viewassethistory_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassethistory"
)

func TestRun_EndToEnd_ShowsFullHistoryFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// AAPL bought in 03/2026 with 1000 USD (asset row), then 500 in 04 and 200 in 05.
	// Results: 04 → 1800, 05 → 1900.
	//
	// Holding model:
	//   04: holding = 1000 (asset) + 500 (tx 04) = 1500. G/P = 1800-1500 = +300 (+20%).
	//   05: holding = 1800 (prev result) + 200 (tx 05)  = 2000. G/P = 1900-2000 = -100 (-5%).
	// Total Aportado lifetime = 1700, Total G/P = 1900-1700 = +200.
	// Medias mensuais: pct = (20 + -5)/2 = +7.50%, gain = (300 + -100)/2 = +100.
	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	for _, tx := range []domain.Transaction{
		{AssetID: assetID, AmountUSD: 500, Month: 4, Year: 2026},
		{AssetID: assetID, AmountUSD: 200, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertTransaction(tx); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1800, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1900, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := viewassethistory.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Reporte histórico dun activo",
		"Acción — AAPL",
		// Top summary
		"Total Aportado",
		"1700.00 USD",
		"Índice Medio Mensual",
		"+7.50%",
		"Gañanzas/Perdas Medias Mensuais",
		"+100.00 USD",
		"Total Gañanzas/Perdas",
		"+200.00 USD",
		// Table columns
		"Aporte Mensual",
		"No activo",
		"G/P USD",
		"Resultado",
		// Row values
		"500.00", "1500.00", "+20.00%", "+300.00", "1800.00",
		"200.00", "2000.00", "-5.00%", "-100.00", "1900.00",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
