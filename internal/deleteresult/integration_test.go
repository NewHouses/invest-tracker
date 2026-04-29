package deleteresult_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/deleteresult"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_DeletesFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	id1, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertMonthlyResult[1]: %v", err)
	}
	id2, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1200, Month: 5, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertMonthlyResult[2]: %v", err)
	}

	// Input: activo 1, resultado 2 (o de 05/2026)
	r := bufio.NewReader(strings.NewReader("1\n2\n"))
	var out bytes.Buffer
	if err := deleteresult.Run(r, &out, s); err != nil {
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

	list, err := s2.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(list))
	}
	if list[0].ID != id1 {
		t.Errorf("permanece id=%d, esperabamos id=%d (id2=%d borrado)", list[0].ID, id1, id2)
	}
}
