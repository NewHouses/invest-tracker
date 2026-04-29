package store_test

import (
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_InsertMonthlyResult(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)

	id, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1100.50, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}
	if id <= 0 {
		t.Errorf("got id=%d, esperabamos > 0", id)
	}
}

func TestStore_TotalInvested_OnlyInitial(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	total, err := s.TotalInvested(assetID)
	if err != nil {
		t.Fatalf("TotalInvested: %v", err)
	}
	if total != 1000 {
		t.Errorf("got %v, esperabamos 1000", total)
	}
}

func TestStore_TotalInvested_WithTransactions(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	for _, amt := range []float64{250.50, 500.00} {
		if _, err := s.InsertTransaction(domain.Transaction{
			AssetID: assetID, AmountUSD: amt, Month: 2, Year: 2026,
		}); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}

	total, err := s.TotalInvested(assetID)
	if err != nil {
		t.Fatalf("TotalInvested: %v", err)
	}
	if total != 1750.50 {
		t.Errorf("got %v, esperabamos 1750.50", total)
	}
}

func TestStore_RejectsOrphanResult(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	_, err = s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: 999, ResultUSD: 1000, Month: 1, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro de FK orfa")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "foreign") {
		t.Errorf("esperabamos erro FOREIGN KEY, got: %v", err)
	}
}

func TestStore_RejectsInvalidResultMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	_, err = s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1000, Month: 13, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro do CHECK constraint do mes")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("esperabamos erro de constraint, got: %v", err)
	}
}

func TestStore_ListMonthlyResultsByAsset(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1300, Month: 5, Year: 2026},
		{AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1500, Month: 11, Year: 2025},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
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

func TestStore_DeleteMonthlyResult(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
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

	if err := s.DeleteMonthlyResult(id1); err != nil {
		t.Fatalf("DeleteMonthlyResult: %v", err)
	}

	got, err := s.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(got))
	}
	if got[0].ID != id2 {
		t.Errorf("permanece id=%d, esperabamos id=%d", got[0].ID, id2)
	}
}

func TestStore_DeleteMonthlyResult_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	err = s.DeleteMonthlyResult(999)
	if err == nil {
		t.Fatal("esperabamos erro por id inexistente")
	}
}
