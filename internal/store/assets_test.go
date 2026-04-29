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

func TestStore_UpdateAsset(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	if err := s.UpdateAsset(domain.Asset{
		ID: id, Type: domain.Accion, Name: "AAPL Updated", AmountUSD: 1500, Month: 6, Year: 2027,
	}); err != nil {
		t.Fatalf("UpdateAsset: %v", err)
	}

	got, err := s.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, esperabamos 1", len(got))
	}
	want := domain.Asset{
		ID: id, Type: domain.Accion, Name: "AAPL Updated", AmountUSD: 1500, Month: 6, Year: 2027,
	}
	if got[0] != want {
		t.Errorf("got %+v, esperabamos %+v", got[0], want)
	}
}

func TestStore_UpdateAsset_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	err = s.UpdateAsset(domain.Asset{
		ID: 999, Type: domain.Accion, Name: "ghost", AmountUSD: 1, Month: 1, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro por id inexistente")
	}
}

func TestStore_UpdateAsset_RejectsInvalidMonth(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	err = s.UpdateAsset(domain.Asset{
		ID: id, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 13, Year: 2026,
	})
	if err == nil {
		t.Fatal("esperabamos erro por mes inválido")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "constraint") {
		t.Errorf("erro inesperado: %v", err)
	}
}

func TestStore_DeleteAsset_CascadesChildren(t *testing.T) {
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

	for _, tx := range []domain.Transaction{
		{AssetID: id1, AmountUSD: 500, Month: 2, Year: 2026},
		{AssetID: id1, AmountUSD: 300, Month: 3, Year: 2026},
		{AssetID: id2, AmountUSD: 100, Month: 2, Year: 2026},
	} {
		if _, err := s.InsertTransaction(tx); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1600, Month: 2, Year: 2026},
		{AssetID: id2, ResultUSD: 2100, Month: 2, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	if err := s.DeleteAsset(id1); err != nil {
		t.Fatalf("DeleteAsset: %v", err)
	}

	assets, err := s.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(assets) != 1 || assets[0].ID != id2 {
		t.Errorf("permanece %v, esperabamos só id=%d", assets, id2)
	}

	txsRemoved, err := s.ListTransactionsByAsset(id1)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset(id1): %v", err)
	}
	if len(txsRemoved) != 0 {
		t.Errorf("esperabamos 0 transaccións de id=%d, got %d", id1, len(txsRemoved))
	}

	txsKept, err := s.ListTransactionsByAsset(id2)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset(id2): %v", err)
	}
	if len(txsKept) != 1 {
		t.Errorf("esperabamos 1 transacción de id=%d, got %d", id2, len(txsKept))
	}
}

func TestStore_DeleteAsset_NoRow(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	err = s.DeleteAsset(999)
	if err == nil {
		t.Fatal("esperabamos erro por id inexistente")
	}
}
