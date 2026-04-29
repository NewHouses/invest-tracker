package deleteasset_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/deleteasset"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_DeletesAssetAndCascades(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	id1, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[1]: %v", err)
	}
	id2, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[2]: %v", err)
	}

	for _, tx := range []domain.Transaction{
		{AssetID: id1, AmountUSD: 500, Month: 4, Year: 2026},
		{AssetID: id2, AmountUSD: 100, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertTransaction(tx); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: id1, ResultUSD: 1700, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}

	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := deleteasset.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	assets, err := s2.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(assets) != 1 || assets[0].ID != id2 {
		t.Errorf("permanecen %v, esperabamos só id=%d", assets, id2)
	}

	txsRemoved, err := s2.ListTransactionsByAsset(id1)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset(id1): %v", err)
	}
	if len(txsRemoved) != 0 {
		t.Errorf("esperabamos 0 transaccións de id=%d tras borrado, got %d", id1, len(txsRemoved))
	}

	txsKept, err := s2.ListTransactionsByAsset(id2)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset(id2): %v", err)
	}
	if len(txsKept) != 1 {
		t.Errorf("esperabamos 1 transacción de id=%d, got %d", id2, len(txsKept))
	}
}
