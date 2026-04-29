package store_test

import (
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_InsertDividend(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id, err := s.InsertDividend(domain.Dividend{
		AmountUSD: 125.50, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertDividend: %v", err)
	}
	if id <= 0 {
		t.Errorf("got id=%d, esperabamos > 0", id)
	}
}

func TestStore_RejectsInvalidDividendMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.InsertDividend(domain.Dividend{
		AmountUSD: 100, Month: 13, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro do CHECK constraint do mes")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("esperabamos erro de constraint, got: %v", err)
	}
}

func TestStore_ListDividends(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	for _, d := range []domain.Dividend{
		{AmountUSD: 200, Month: 5, Year: 2026},
		{AmountUSD: 100, Month: 4, Year: 2026},
		{AmountUSD: 300, Month: 11, Year: 2025},
	} {
		if _, err := s.InsertDividend(d); err != nil {
			t.Fatalf("InsertDividend: %v", err)
		}
	}

	got, err := s.ListDividends()
	if err != nil {
		t.Fatalf("ListDividends: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d filas, esperabamos 3", len(got))
	}
	want := []struct{ year, month int }{
		{2025, 11},
		{2026, 4},
		{2026, 5},
	}
	for i, w := range want {
		if got[i].Year != w.year || got[i].Month != w.month {
			t.Errorf("fila[%d] = %d/%d, queremos %d/%d",
				i, got[i].Year, got[i].Month, w.year, w.month)
		}
	}
}

func TestStore_DeleteDividend(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, err := s.InsertDividend(domain.Dividend{
		AmountUSD: 100, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertDividend[1]: %v", err)
	}
	id2, err := s.InsertDividend(domain.Dividend{
		AmountUSD: 200, Month: 5, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertDividend[2]: %v", err)
	}

	if err := s.DeleteDividend(id1); err != nil {
		t.Fatalf("DeleteDividend: %v", err)
	}

	got, err := s.ListDividends()
	if err != nil {
		t.Fatalf("ListDividends: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(got))
	}
	if got[0].ID != id2 {
		t.Errorf("permanece id=%d, esperabamos id=%d", got[0].ID, id2)
	}
}

func TestStore_DeleteDividend_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	err = s.DeleteDividend(999)
	if err == nil {
		t.Fatal("esperabamos erro por id inexistente")
	}
}
