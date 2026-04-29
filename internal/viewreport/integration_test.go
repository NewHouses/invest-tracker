package viewreport_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewreport"
)

func TestRun_EndToEnd_ShowsTableFromDB(t *testing.T) {
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
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1800, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}

	input := "1\n4\n2026\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := viewreport.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Acción — AAPL · 04/2026",
		"Investido ata o mes",
		"1500.00 USD",
		"Investido este mes",
		"500.00 USD",
		"Resultado",
		"1800.00 USD",
		"+300.00 USD",
		"+20.00%",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
