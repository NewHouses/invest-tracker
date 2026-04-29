package store_test

import (
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_InsertAndList(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	want := []domain.Asset{
		{Type: domain.Accion, Name: "AAPL", AmountUSD: 1000.50, Month: 4, Year: 2026},
		{Type: domain.CopyTrading, Name: "Copy Trader X", AmountUSD: 500.00, Month: 3, Year: 2026},
	}

	for i := range want {
		id, err := s.InsertAsset(want[i])
		if err != nil {
			t.Fatalf("InsertAsset[%d]: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("InsertAsset[%d] devolveu id=%d, esperabamos > 0", i, id)
		}
		want[i].ID = id
	}

	got, err := s.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ListAssets devolveu %d filas, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("fila[%d] = %+v, queremos %+v", i, got[i], w)
		}
	}
}

func TestStore_RejectsInvalidType(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.InsertAsset(domain.Asset{
		Type: "invalido", Name: "X", AmountUSD: 1, Month: 1, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro do CHECK constraint do tipo")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("erro inesperado: %v", err)
	}
}

func TestStore_RejectsInvalidMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "X", AmountUSD: 1, Month: 13, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro do CHECK constraint do mes")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("erro inesperado: %v", err)
	}
}
