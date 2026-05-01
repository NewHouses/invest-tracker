package addtxmonth_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addtxmonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsTransactionsForMonth(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	aaplID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset AAPL: %v", err)
	}
	msftID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "MSFT", AmountUSD: 500, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset MSFT: %v", err)
	}
	vanID, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	// Aporte 04/2026: AAPL 250, MSFT skip, Vanguard 500.
	r := bufio.NewReader(strings.NewReader("4\n2026\n250\n\n500\n"))
	var out bytes.Buffer
	if err := addtxmonth.Run(r, &out, s); err != nil {
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

	aaplTxs, err := s2.ListTransactionsByAsset(aaplID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset AAPL: %v", err)
	}
	if len(aaplTxs) != 1 || aaplTxs[0].AmountUSD != 250 ||
		aaplTxs[0].Month != 4 || aaplTxs[0].Year != 2026 {
		t.Errorf("AAPL txs = %+v, esperabamos 1×250 en 04/2026", aaplTxs)
	}

	msftTxs, err := s2.ListTransactionsByAsset(msftID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset MSFT: %v", err)
	}
	if len(msftTxs) != 0 {
		t.Errorf("MSFT non debería ter txs (saltado), got %+v", msftTxs)
	}

	vanTxs, err := s2.ListTransactionsByAsset(vanID)
	if err != nil {
		t.Fatalf("ListTransactionsByAsset Vanguard: %v", err)
	}
	if len(vanTxs) != 1 || vanTxs[0].AmountUSD != 500 {
		t.Errorf("Vanguard txs = %+v, esperabamos 1×500", vanTxs)
	}

	output := out.String()
	if !strings.Contains(output, "2 transacción(s) engadidas (750.00 USD total), 1 saltado") {
		t.Errorf("saída non contén resumo correcto:\n%s", output)
	}
}
