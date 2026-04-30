package viewtransactions_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtransactions"
)

func TestRun_EndToEnd_ListsTransactions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset AAPL: %v", err)
	}
	// Outro activo (non debe aparecer)
	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 9999, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	for _, tx := range []domain.Transaction{
		{AssetID: id1, AmountUSD: 500, Month: 4, Year: 2026},
		{AssetID: id1, AmountUSD: -200, Month: 5, Year: 2026},
		{AssetID: id1, AmountUSD: 300, Month: 6, Year: 2026},
	} {
		if _, err := s.InsertTransaction(tx); err != nil {
			t.Fatalf("InsertTransaction: %v", err)
		}
	}

	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := viewtransactions.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()

	for _, want := range []string{
		"Transaccións de Acción — AAPL",
		"3 transacción(s)",
		"Compras: 800.00 USD",
		"Vendas: 200.00 USD",
		"+600.00 USD",
		"COMPRA",
		"VENDA",
		"500.00 USD",
		"200.00 USD",
		"300.00 USD",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// O outro activo non debe aparecer
	if strings.Contains(output, "9999") {
		t.Errorf("saída non debería conter datos de Vanguard:\n%s", output)
	}
}

func TestRun_EndToEnd_AssetWithoutTransactions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}

	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := viewtransactions.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(out.String(), "non ten transaccións rexistradas") {
		t.Errorf("saída non contén suxestión:\n%s", out.String())
	}
}
