package store_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestStore_MonthsWithResultsUpTo(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1200, Month: 5, Year: 2026},
		{AssetID: assetID, ResultUSD: 900, Month: 11, Year: 2025},
		{AssetID: assetID, ResultUSD: 1300, Month: 6, Year: 2026}, // máis aló do target
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.MonthsWithResultsUpTo(2026, 5)
	if err != nil {
		t.Fatalf("MonthsWithResultsUpTo: %v", err)
	}
	want := []domain.YearMonth{
		{Year: 2025, Month: 11},
		{Year: 2026, Month: 4},
		{Year: 2026, Month: 5},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, esperabamos %d (%+v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] = %+v, esperabamos %+v", i, got[i], w)
		}
	}
}

func TestStore_MonthsWithResultsUpTo_DedupesAcrossAssets(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, _ := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	id2, _ := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 1, Year: 2026,
	})
	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: id2, ResultUSD: 2100, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, _ := s.MonthsWithResultsUpTo(2026, 4)
	if len(got) != 1 {
		t.Fatalf("len = %d, esperabamos 1 (deduplicado), got %+v", len(got), got)
	}
}

func TestStore_SumDividends(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	for _, d := range []domain.Dividend{
		{AmountUSD: 50.25, Month: 4, Year: 2026},
		{AmountUSD: 30.00, Month: 4, Year: 2026},
		{AmountUSD: 75.00, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertDividend(d); err != nil {
			t.Fatalf("InsertDividend: %v", err)
		}
	}

	got, err := s.SumDividends(2026, 4)
	if err != nil {
		t.Fatalf("SumDividends: %v", err)
	}
	if got != 80.25 {
		t.Errorf("SumDividends(04/2026) = %v, esperabamos 80.25", got)
	}
}

func TestStore_MonthsWithResults_All(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1200, Month: 5, Year: 2026},
		{AssetID: assetID, ResultUSD: 900, Month: 11, Year: 2025},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.MonthsWithResults()
	if err != nil {
		t.Fatalf("MonthsWithResults: %v", err)
	}
	want := []domain.YearMonth{
		{Year: 2025, Month: 11},
		{Year: 2026, Month: 4},
		{Year: 2026, Month: 5},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] = %+v, esperabamos %+v", i, got[i], w)
		}
	}
}

func TestStore_MonthsWithResultsForAsset(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, _ := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	id2, _ := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 1, Year: 2026,
	})
	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: id1, ResultUSD: 1200, Month: 5, Year: 2026},
		{AssetID: id2, ResultUSD: 2100, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.MonthsWithResultsForAsset(id1)
	if err != nil {
		t.Fatalf("MonthsWithResultsForAsset: %v", err)
	}
	want := []domain.YearMonth{
		{Year: 2026, Month: 4},
		{Year: 2026, Month: 5},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] = %+v, esperabamos %+v", i, got[i], w)
		}
	}
}

func TestStore_MonthsWithResults_Empty(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	got, err := s.MonthsWithResults()
	if err != nil {
		t.Fatalf("MonthsWithResults: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, esperabamos 0", len(got))
	}
}

func TestStore_SumDividends_Empty(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	got, err := s.SumDividends(2026, 4)
	if err != nil {
		t.Fatalf("SumDividends: %v", err)
	}
	if got != 0 {
		t.Errorf("SumDividends sen filas = %v, esperabamos 0", got)
	}
}
