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
		"Acción — AAPL",
		"Total investido",
		"1700.00 USD",
		"Total Ganhanzas/Perdas",
		"+200.00 USD",
		"Índice medio mensual",
		"+15.88",
		"Investido ata o mes",
		"Investido este mes",
		"Resultado",
		"1500.00",
		"500.00",
		"+20.00%",
		"+300.00",
		"1800.00",
		"1700.00",
		"200.00",
		"+11.76%",
		"+200.00",
		"1900.00",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
