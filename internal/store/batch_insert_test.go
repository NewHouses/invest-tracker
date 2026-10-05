package store_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_InsertTransactions(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	items := []domain.Transaction{
		{AssetID: assetID, AmountUSD: 100, Month: 2, Year: 2026},
		{AssetID: assetID, AmountUSD: 200, Month: 3, Year: 2026},
	}
	ids, err := s.InsertTransactions(items)
	if err != nil {
		t.Fatalf("InsertTransactions: %v", err)
	}
	if len(ids) != len(items) {
		t.Fatalf("len(ids) = %d, esperabamos %d", len(ids), len(items))
	}
	for i, id := range ids {
		if id <= 0 {
			t.Fatalf("ids[%d] = %d, esperabamos > 0", i, id)
		}
		items[i].ID = id
	}
	got, err := s.ListTransactionsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset: %v", err)
	}
	if len(got) != len(items) {
		t.Fatalf("filas = %d, esperabamos %d", len(got), len(items))
	}
	for i := range items {
		if got[i] != items[i] {
			t.Fatalf("fila[%d] = %+v, queremos %+v", i, got[i], items[i])
		}
	}
}

func TestStore_InsertTransactions_Empty(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	ids, err := s.InsertTransactions(nil)
	if err != nil {
		t.Fatalf("InsertTransactions baleiro: %v", err)
	}
	if ids != nil {
		t.Fatalf("ids = %#v, esperabamos nil", ids)
	}
}

func TestStore_InsertTransactions_Rollback(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	_, err = s.InsertTransactions([]domain.Transaction{
		{AssetID: assetID, AmountUSD: 100, Month: 2, Year: 2026},
		{AssetID: 999, AmountUSD: 200, Month: 3, Year: 2026},
	})
	if err == nil {
		t.Fatal("esperabamos erro de FK")
	}
	got, err := s.ListTransactionsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("rollback deixou %d transaccións, esperabamos 0", len(got))
	}
}

func TestStore_InsertMonthlyResults(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	items := []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1100, Month: 2, Year: 2026},
		{AssetID: assetID, ResultUSD: 1200, Month: 3, Year: 2026},
	}
	ids, err := s.InsertMonthlyResults(items)
	if err != nil {
		t.Fatalf("InsertMonthlyResults: %v", err)
	}
	if len(ids) != len(items) {
		t.Fatalf("len(ids) = %d, esperabamos %d", len(ids), len(items))
	}
	for i, id := range ids {
		if id <= 0 {
			t.Fatalf("ids[%d] = %d, esperabamos > 0", i, id)
		}
		items[i].ID = id
	}
	got, err := s.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(got) != len(items) {
		t.Fatalf("filas = %d, esperabamos %d", len(got), len(items))
	}
	for i := range items {
		if got[i] != items[i] {
			t.Fatalf("fila[%d] = %+v, queremos %+v", i, got[i], items[i])
		}
	}
}

func TestStore_InsertMonthlyResults_Empty(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	ids, err := s.InsertMonthlyResults(nil)
	if err != nil {
		t.Fatalf("InsertMonthlyResults baleiro: %v", err)
	}
	if ids != nil {
		t.Fatalf("ids = %#v, esperabamos nil", ids)
	}
}

func TestStore_InsertMonthlyResults_Rollback(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	_, err = s.InsertMonthlyResults([]domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1100, Month: 2, Year: 2026},
		{AssetID: 999, ResultUSD: 1200, Month: 3, Year: 2026},
	})
	if err == nil {
		t.Fatal("esperabamos erro de FK")
	}
	got, err := s.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("rollback deixou %d resultados, esperabamos 0", len(got))
	}
}
