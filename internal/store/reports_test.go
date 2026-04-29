package store_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_MonthlySummary_FullPath(t *testing.T) {
	s, err := store.Open(":memory:")
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
	for _, tx := range []domain.Transaction{
		{AssetID: assetID, AmountUSD: 500, Month: 4, Year: 2026},
		{AssetID: assetID, AmountUSD: 200, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertTransaction(tx); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: assetID, ResultUSD: 1800, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}

	got, err := s.MonthlySummary(assetID, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	want := domain.MonthlySummary{
		TotalInvestedUpTo: 1500,
		InvestedInMonth:   500,
		Result:            1800,
		HasResult:         true,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v", got, want)
	}
}

func TestStore_MonthlySummary_NoResultRecorded(t *testing.T) {
	s, err := store.Open(":memory:")
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
	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: assetID, AmountUSD: 500, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	got, err := s.MonthlySummary(assetID, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if got.HasResult {
		t.Errorf("HasResult = true; esperabamos false")
	}
	if got.Result != 0 {
		t.Errorf("Result = %v; esperabamos 0", got.Result)
	}
	if got.TotalInvestedUpTo != 1500 {
		t.Errorf("TotalInvestedUpTo = %v; esperabamos 1500", got.TotalInvestedUpTo)
	}
	if got.InvestedInMonth != 500 {
		t.Errorf("InvestedInMonth = %v; esperabamos 500", got.InvestedInMonth)
	}
}

func TestStore_MonthlySummary_BeforeAnyAsset(t *testing.T) {
	s, err := store.Open(":memory:")
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

	got, err := s.MonthlySummary(assetID, 2026, 1)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	want := domain.MonthlySummary{
		TotalInvestedUpTo: 0,
		InvestedInMonth:   0,
		Result:            0,
		HasResult:         false,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v", got, want)
	}
}

func TestStore_MonthlySummary_OnlyInitial(t *testing.T) {
	s, err := store.Open(":memory:")
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

	got, err := s.MonthlySummary(assetID, 2026, 3)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if got.TotalInvestedUpTo != 1000 {
		t.Errorf("TotalInvestedUpTo = %v; esperabamos 1000", got.TotalInvestedUpTo)
	}
	if got.InvestedInMonth != 1000 {
		t.Errorf("InvestedInMonth = %v; esperabamos 1000", got.InvestedInMonth)
	}
	if got.HasResult {
		t.Errorf("HasResult = true; esperabamos false")
	}
}
