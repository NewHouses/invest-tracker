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

// Corrixir o resultado dun mes substitúe o anterior: antes engadíase unha
// fila nova e, ao borrar a visible, reaparecía o valor vello.
func TestStore_InsertMonthlyResult_ReplacesSameMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assetID := seedAsset(t, s)
	first, err := s.InsertMonthlyResult(domain.MonthlyResult{AssetID: assetID, ResultUSD: 1100, Month: 4, Year: 2026})
	if err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}
	second, err := s.InsertMonthlyResult(domain.MonthlyResult{AssetID: assetID, ResultUSD: 1150, Month: 4, Year: 2026})
	if err != nil {
		t.Fatalf("InsertMonthlyResult (corrección): %v", err)
	}
	ids, err := s.InsertMonthlyResults([]domain.MonthlyResult{
		{AssetID: assetID, ResultUSD: 1175, Month: 4, Year: 2026},
		{AssetID: assetID, ResultUSD: 1200, Month: 5, Year: 2026},
	})
	if err != nil {
		t.Fatalf("InsertMonthlyResults: %v", err)
	}
	if second != first || ids[0] != first || ids[1] == first {
		t.Fatalf("ids = %d, %d, %v; esperabamos reutilizar %d no mesmo mes", first, second, ids, first)
	}

	got, err := s.ListMonthlyResultsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(got) != 2 || got[0].ResultUSD != 1175 || got[1].ResultUSD != 1200 {
		t.Fatalf("resultados = %+v, esperabamos un por mes (1175 en 04, 1200 en 05)", got)
	}

	if err := s.DeleteMonthlyResult(first); err != nil {
		t.Fatalf("DeleteMonthlyResult: %v", err)
	}
	sum, err := s.MonthlySummary(assetID, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if sum.HasResult {
		t.Fatalf("tras borrar o resultado de 04/2026 non debería quedar ningún: %+v", sum)
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

func TestStore_DeleteMonthlyResultsByMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[1]: %v", err)
	}
	id2, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[2]: %v", err)
	}

	// 2 resultados en 04/2026 (un por activo) + 1 resultado en 05/2026
	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: id2, ResultUSD: 2100, Month: 4, Year: 2026},
		{AssetID: id1, ResultUSD: 1200, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	n, err := s.DeleteMonthlyResultsByMonth(2026, 4)
	if err != nil {
		t.Fatalf("DeleteMonthlyResultsByMonth: %v", err)
	}
	if n != 2 {
		t.Errorf("filas borradas = %d, esperabamos 2", n)
	}

	// 04/2026 baleiro
	if got, _ := s.ListMonthlyResultsByAsset(id1); len(got) != 1 || got[0].Month != 5 {
		t.Errorf("AAPL debería ter só o resultado de 05/2026, got %+v", got)
	}
	if got, _ := s.ListMonthlyResultsByAsset(id2); len(got) != 0 {
		t.Errorf("Vanguard debería non ter resultados, got %+v", got)
	}
}

func TestStore_DeleteMonthlyResultsByMonth_NoRows(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// Sen filas: a operación non debe erro, devolve 0.
	n, err := s.DeleteMonthlyResultsByMonth(2026, 4)
	if err != nil {
		t.Fatalf("DeleteMonthlyResultsByMonth: %v", err)
	}
	if n != 0 {
		t.Errorf("filas borradas = %d, esperabamos 0", n)
	}
}
