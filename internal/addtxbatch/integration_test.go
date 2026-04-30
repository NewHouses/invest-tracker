package addtxbatch_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addtxbatch"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_Mode3_AutoIncrement(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	assetID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 11, Year: 2025,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	// asset=1, mode=3, mes=11, ano=2025
	// tx1: compra 100, sí
	// tx2: compra 200, sí
	// tx3: venda 50, n
	r := bufio.NewReader(strings.NewReader("1\n3\n11\n2025\n1\n100\ns\n1\n200\ns\n2\n50\nn\n"))
	var out bytes.Buffer
	if err := addtxbatch.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	list, err := s2.ListTransactionsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("got %d filas, esperabamos 3", len(list))
	}

	// As 3 transaccións deben estar en 11/2025, 12/2025, 01/2026
	want := []struct {
		amount       float64
		year, month  int
	}{
		{100, 2025, 11},
		{200, 2025, 12},
		{-50, 2026, 1},
	}
	for i, w := range want {
		if list[i].AmountUSD != w.amount || list[i].Year != w.year || list[i].Month != w.month {
			t.Errorf("tx[%d] = %+v, esperabamos %+v", i, list[i], w)
		}
	}

	if !strings.Contains(out.String(), "Engadíronse 3 transacción(s)") {
		t.Errorf("saída non contén resumo de 3 transaccións:\n%s", out.String())
	}
}

func TestRun_EndToEnd_Mode2_FixedMonth(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
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

	// asset=1, mode=2, mes=4, ano=2026, tx1=100, sí, tx2=200, n
	r := bufio.NewReader(strings.NewReader("1\n2\n4\n2026\n1\n100\ns\n1\n200\nn\n"))
	var out bytes.Buffer
	if err := addtxbatch.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	_ = out

	list, err := s.ListTransactionsByAsset(assetID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d, esperabamos 2", len(list))
	}
	for i, tx := range list {
		if tx.Month != 4 || tx.Year != 2026 {
			t.Errorf("tx[%d] data=%02d/%d, esperabamos 04/2026", i, tx.Month, tx.Year)
		}
	}
}
