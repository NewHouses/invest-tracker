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

func TestRun_EndToEnd_ListsTransactionsWithInitialAndOrdered(t *testing.T) {
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
	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 9999, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	// Inserto transaccións en orde non-cronolóxica para verificar que se ordean por data.
	for _, tx := range []domain.Transaction{
		{AssetID: id1, AmountUSD: 300, Month: 6, Year: 2026},
		{AssetID: id1, AmountUSD: 500, Month: 4, Year: 2026},
		{AssetID: id1, AmountUSD: -200, Month: 5, Year: 2026},
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
		"4 entradas (incluíndo a compra inicial)",
		"Compras: 1800.00 USD", // 1000 inicial + 500 + 300
		"Vendas: 200.00 USD",
		"+1600.00 USD",
		"Compra/Venda",
		"COMPRA",
		"VENDA",
		"1000.00 USD", // inicial
		"500.00 USD",
		"300.00 USD",
		"—", // ID da compra inicial
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Verifica orde cronolóxica dentro do corpo da táboa (evita falsas
	// coincidencias na liña de totais "Compras: 1800.00 USD ...").
	tableStart := strings.Index(output, "Cantidade")
	if tableStart < 0 {
		t.Fatalf("non se atopou cabeceira da táboa:\n%s", output)
	}
	table := output[tableStart:]
	pos1000 := strings.Index(table, "1000.00 USD")
	pos500 := strings.Index(table, "500.00 USD")
	pos200 := strings.Index(table, "200.00 USD")
	pos300 := strings.Index(table, "300.00 USD")
	if !(pos1000 < pos500 && pos500 < pos200 && pos200 < pos300) {
		t.Errorf("filas non en orde cronolóxica: 1000=%d 500=%d 200=%d 300=%d\n%s",
			pos1000, pos500, pos200, pos300, table)
	}

	if strings.Contains(output, "9999") {
		t.Errorf("saída non debería conter datos de Vanguard:\n%s", output)
	}
}

func TestRun_EndToEnd_AssetWithoutTransactionsShowsOnlyInitial(t *testing.T) {
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

	output := out.String()
	if !strings.Contains(output, "1 entradas (incluíndo a compra inicial)") {
		t.Errorf("saída non mostra exactamente 1 entrada:\n%s", output)
	}
	if !strings.Contains(output, "1000.00 USD") {
		t.Errorf("saída non mostra a compra inicial 1000.00 USD:\n%s", output)
	}
	if !strings.Contains(output, "Compras: 1000.00 USD") {
		t.Errorf("totais non inclúen a compra inicial:\n%s", output)
	}
}
