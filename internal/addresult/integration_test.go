package addresult_test

import (
	"bufio"
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"invest-tracker/internal/addresult"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsAndShowsCalculations(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: assetID, AmountUSD: 500, Month: 2, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	// Input layout: month, year, asset selection, result
	input := "5\n2026\n1\n1800\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := addresult.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"investido: 1500.00 USD",
		"+300.00 USD",
		"+20.00%",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var (
		gotID      int64
		gotAssetID int64
		gotResult  float64
		gotMonth   int
		gotYear    int
	)
	row := db.QueryRow(
		`SELECT id, asset_id, result_usd, month, year FROM monthly_results WHERE asset_id = ?`,
		assetID,
	)
	if err := row.Scan(&gotID, &gotAssetID, &gotResult, &gotMonth, &gotYear); err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if gotID <= 0 {
		t.Errorf("id = %d, esperabamos > 0", gotID)
	}
	if gotAssetID != assetID || gotResult != 1800 || gotMonth != 5 || gotYear != 2026 {
		t.Errorf("fila = (%d, %d, %v, %d, %d); queremos (%d, %d, 1800, 5, 2026)",
			gotID, gotAssetID, gotResult, gotMonth, gotYear, gotID, assetID)
	}
}
