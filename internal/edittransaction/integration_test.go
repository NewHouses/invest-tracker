package edittransaction_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/edittransaction"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_EditsAmount(t *testing.T) {
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
	txID, err := s.InsertTransaction(domain.Transaction{
		AssetID: assetID, AmountUSD: 500, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	// Activo 1, tx 1, campo 1 (cantidade), nova cantidade 750
	r := bufio.NewReader(strings.NewReader("1\n1\n1\n750\n"))
	var out bytes.Buffer
	if err := edittransaction.Run(r, &out, s); err != nil {
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

	list, err := s2.ListTransactionsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d, esperabamos 1", len(list))
	}
	got := list[0]
	want := domain.Transaction{
		ID: txID, AssetID: assetID, AmountUSD: 750, Month: 4, Year: 2026,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v", got, want)
	}
}

func TestRun_EndToEnd_FlipsTypeAndUpdatesDate(t *testing.T) {
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
	txID, err := s.InsertTransaction(domain.Transaction{
		AssetID: assetID, AmountUSD: 500, Month: 4, Year: 2026, // COMPRA
	})
	if err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	// Cambiar tipo a Venda
	r := bufio.NewReader(strings.NewReader("1\n1\n2\n2\n"))
	var out bytes.Buffer
	if err := edittransaction.Run(r, &out, s); err != nil {
		t.Fatalf("Run flip type: %v", err)
	}

	// Cambiar data a 06/2026
	r2 := bufio.NewReader(strings.NewReader("1\n1\n3\n6\n2026\n"))
	if err := edittransaction.Run(r2, &out, s); err != nil {
		t.Fatalf("Run change date: %v", err)
	}

	list, _ := s.ListTransactionsByAsset(assetID)
	if len(list) != 1 {
		t.Fatalf("got %d, esperabamos 1", len(list))
	}
	got := list[0]
	want := domain.Transaction{
		ID: txID, AssetID: assetID, AmountUSD: -500, Month: 6, Year: 2026,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v (venda 500 en 06/2026)", got, want)
	}
}
