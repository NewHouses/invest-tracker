package store_test

import (
	"database/sql"
	"errors"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_GetAsset(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	want := domain.Asset{Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026}
	id, err := s.InsertAsset(want)
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	want.ID = id

	got, err := s.GetAsset(id)
	if err != nil {
		t.Fatalf("GetAsset: %v", err)
	}
	if got != want {
		t.Fatalf("GetAsset = %+v, queremos %+v", got, want)
	}
}

func TestStore_GetAsset_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.GetAsset(999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetAsset inexistente erro = %v, queremos sql.ErrNoRows", err)
	}
}

func TestStore_GetTransaction(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	want := domain.Transaction{AssetID: assetID, AmountUSD: 250, Month: 2, Year: 2026}
	id, err := s.InsertTransaction(want)
	if err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}
	want.ID = id

	got, err := s.GetTransaction(id)
	if err != nil {
		t.Fatalf("GetTransaction: %v", err)
	}
	if got != want {
		t.Fatalf("GetTransaction = %+v, queremos %+v", got, want)
	}
}

func TestStore_GetTransaction_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.GetTransaction(999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetTransaction inexistente erro = %v, queremos sql.ErrNoRows", err)
	}
}
