package store_test

import (
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func seedInvestment(t *testing.T, s *store.Store) int64 {
	t.Helper()
	id, err := s.InsertInvestment(domain.Investment{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("seed InsertInvestment: %v", err)
	}
	return id
}

func TestStore_InsertAndListTransaction(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	invID := seedInvestment(t, s)

	want := []domain.Transaction{
		{InvestmentID: invID, AmountUSD: 250.50, Month: 2, Year: 2026},
		{InvestmentID: invID, AmountUSD: 100.00, Month: 3, Year: 2026},
	}
	for i := range want {
		id, err := s.InsertTransaction(want[i])
		if err != nil {
			t.Fatalf("InsertTransaction[%d]: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("InsertTransaction[%d] devolveu id=%d, esperabamos > 0", i, id)
		}
		want[i].ID = id
	}

	got, err := s.ListTransactionsByInvestment(invID)
	if err != nil {
		t.Fatalf("ListTransactionsByInvestment: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ListTransactionsByInvestment devolveu %d filas, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("fila[%d] = %+v, queremos %+v", i, got[i], w)
		}
	}
}

func TestStore_RejectsOrphanTransaction(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.InsertTransaction(domain.Transaction{
		InvestmentID: 999, AmountUSD: 100, Month: 1, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro de FK orfa")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "foreign") {
		t.Errorf("esperabamos erro FOREIGN KEY, got: %v", err)
	}
}

func TestStore_RejectsInvalidTransactionMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	invID := seedInvestment(t, s)
	_, err = s.InsertTransaction(domain.Transaction{
		InvestmentID: invID, AmountUSD: 100, Month: 13, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro do CHECK constraint do mes")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("esperabamos erro de constraint, got: %v", err)
	}
}
