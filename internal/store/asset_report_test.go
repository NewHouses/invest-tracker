package store_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestStore_AssetReport_FullPath(t *testing.T) {
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
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1800, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1900, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.AssetReport(assetID)
	if err != nil {
		t.Fatalf("AssetReport: %v", err)
	}

	if len(got.Rows) != 2 {
		t.Fatalf("len(Rows) = %d, esperabamos 2", len(got.Rows))
	}

	row0 := got.Rows[0]
	if row0.Year != 2026 || row0.Month != 4 ||
		row0.InvestedInMonth != 500 || row0.TotalInvestedUpTo != 1500 ||
		row0.Result != 1800 || row0.Gain != 300 ||
		!almostEqual(row0.GainPct, 20) || !row0.HasGainPct {
		t.Errorf("Rows[0] = %+v", row0)
	}
	row1 := got.Rows[1]
	if row1.Year != 2026 || row1.Month != 5 ||
		row1.InvestedInMonth != 200 || row1.TotalInvestedUpTo != 1700 ||
		row1.Result != 1900 || row1.Gain != 200 ||
		!almostEqual(row1.GainPct, 11.7647) || !row1.HasGainPct {
		t.Errorf("Rows[1] = %+v", row1)
	}

	if got.TotalInvested != 1700 {
		t.Errorf("TotalInvested = %v, esperabamos 1700", got.TotalInvested)
	}
	if !got.HasTotalGain || got.TotalGain != 200 {
		t.Errorf("TotalGain = %v (HasTotalGain=%v), esperabamos 200/true", got.TotalGain, got.HasTotalGain)
	}
	if !got.HasAvgIndex || !almostEqual(got.AvgMonthlyIndexPct, (20+11.7647)/2) {
		t.Errorf("AvgMonthlyIndexPct = %v (HasAvgIndex=%v)", got.AvgMonthlyIndexPct, got.HasAvgIndex)
	}
}

func TestStore_AssetReport_NoResults(t *testing.T) {
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

	got, err := s.AssetReport(assetID)
	if err != nil {
		t.Fatalf("AssetReport: %v", err)
	}
	if len(got.Rows) != 0 {
		t.Errorf("Rows len = %d, esperabamos 0", len(got.Rows))
	}
	if got.TotalInvested != 1000 {
		t.Errorf("TotalInvested = %v, esperabamos 1000", got.TotalInvested)
	}
	if got.HasTotalGain {
		t.Errorf("HasTotalGain = true, esperabamos false")
	}
	if got.HasAvgIndex {
		t.Errorf("HasAvgIndex = true, esperabamos false")
	}
}

func TestStore_AssetReport_DedupesDuplicateMonth(t *testing.T) {
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
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1200, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.AssetReport(assetID)
	if err != nil {
		t.Fatalf("AssetReport: %v", err)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("len(Rows) = %d, esperabamos 1", len(got.Rows))
	}
	if got.Rows[0].Result != 1200 {
		t.Errorf("Result = %v, esperabamos 1200 (último por id)", got.Rows[0].Result)
	}
}

func TestStore_AssetReport_OrdersChronologically(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2025,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1300, Month: 2, Year: 2026},
		{AssetID: assetID, ResultUSD: 1100, Month: 1, Year: 2026},
		{AssetID: assetID, ResultUSD: 1500, Month: 11, Year: 2025},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	got, err := s.AssetReport(assetID)
	if err != nil {
		t.Fatalf("AssetReport: %v", err)
	}
	if len(got.Rows) != 3 {
		t.Fatalf("len(Rows) = %d, esperabamos 3", len(got.Rows))
	}
	want := []struct {
		year, month int
	}{
		{2025, 11},
		{2026, 1},
		{2026, 2},
	}
	for i, w := range want {
		if got.Rows[i].Year != w.year || got.Rows[i].Month != w.month {
			t.Errorf("Rows[%d] = %d/%d, esperabamos %d/%d",
				i, got.Rows[i].Year, got.Rows[i].Month, w.year, w.month)
		}
	}
}
