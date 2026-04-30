package editasset_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/editasset"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_UpdatesAssetInDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	id, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	// Editar todo: cantidade. Input: asset 1, field 3 (cantidade), 1500
	r := bufio.NewReader(strings.NewReader("1\n3\n1500\n"))
	var out bytes.Buffer
	if err := editasset.Run(r, &out, s); err != nil {
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

	list, err := s2.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(list))
	}
	got := list[0]
	want := domain.Asset{
		ID: id, Type: domain.Accion, Name: "AAPL", AmountUSD: 1500, Month: 3, Year: 2026,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v", got, want)
	}
}

// TestRun_EndToEnd_EditedAmountReflectsInBothInvestedFields verifica que tras
// editar a cantidade inicial dun activo, MonthlySummary devolve o valor novo
// tanto para TotalInvestedUpTo coma para InvestedInMonth no mes da creación.
func TestRun_EndToEnd_EditedAmountReflectsInBothInvestedFields(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	// Edita a cantidade a 1500
	r := bufio.NewReader(strings.NewReader("1\n3\n1500\n"))
	var out bytes.Buffer
	if err := editasset.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verifica MonthlySummary para o mes de creación (03/2026)
	sum, err := s.MonthlySummary(id, 2026, 3)
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if sum.TotalInvestedUpTo != 1500 {
		t.Errorf("TotalInvestedUpTo = %v, esperabamos 1500 (cantidade editada)", sum.TotalInvestedUpTo)
	}
	if sum.InvestedInMonth != 1500 {
		t.Errorf("InvestedInMonth = %v, esperabamos 1500 (cantidade editada)", sum.InvestedInMonth)
	}
	if sum.TotalInvestedUpTo != sum.InvestedInMonth {
		t.Errorf("TotalInvestedUpTo (%v) != InvestedInMonth (%v); deberían ser iguais no primeiro mes",
			sum.TotalInvestedUpTo, sum.InvestedInMonth)
	}
}

func TestRun_EndToEnd_UpdatesNameAndDate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	id, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	// Edita nome
	r := bufio.NewReader(strings.NewReader("1\n1\nApple Inc.\n"))
	var out bytes.Buffer
	if err := editasset.Run(r, &out, s); err != nil {
		t.Fatalf("Run name: %v", err)
	}

	// Edita data
	r2 := bufio.NewReader(strings.NewReader("1\n2\n8\n2025\n"))
	if err := editasset.Run(r2, &out, s); err != nil {
		t.Fatalf("Run date: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	list, _ := s2.ListAssets()
	if len(list) != 1 {
		t.Fatalf("len = %d, esperabamos 1", len(list))
	}
	got := list[0]
	want := domain.Asset{
		ID: id, Type: domain.Accion, Name: "Apple Inc.", AmountUSD: 1000, Month: 8, Year: 2025,
	}
	if got != want {
		t.Errorf("got %+v, esperabamos %+v", got, want)
	}
}
