package addasset_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addasset"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsToSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	input := "3\nCopy Trader X\n2500.75\n4\n2026\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := addasset.Run(r, &out, s); err != nil {
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

	list, err := s2.ListInvestments()
	if err != nil {
		t.Fatalf("ListInvestments: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(list))
	}

	got := list[0]
	if got.ID <= 0 {
		t.Errorf("ID = %d, esperabamos > 0", got.ID)
	}
	want := domain.Investment{
		ID:        got.ID,
		Type:      domain.CopyTrading,
		Name:      "Copy Trader X",
		AmountUSD: 2500.75,
		Month:     4,
		Year:      2026,
	}
	if got != want {
		t.Errorf("fila gardada = %+v, queremos %+v", got, want)
	}
}
