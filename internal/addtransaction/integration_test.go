package addtransaction_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addtransaction"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsTransaction(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	invID, err := s.InsertInvestment(domain.Investment{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertInvestment: %v", err)
	}

	input := "1\n550.75\n5\n2026\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := addtransaction.Run(r, &out, s); err != nil {
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

	list, err := s2.ListTransactionsByInvestment(invID)
	if err != nil {
		t.Fatalf("ListTransactionsByInvestment: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d filas, esperabamos 1", len(list))
	}

	got := list[0]
	if got.ID <= 0 {
		t.Errorf("ID = %d, esperabamos > 0", got.ID)
	}
	want := domain.Transaction{
		ID:           got.ID,
		InvestmentID: invID,
		AmountUSD:    550.75,
		Month:        5,
		Year:         2026,
	}
	if got != want {
		t.Errorf("fila gardada = %+v, queremos %+v", got, want)
	}
}
