package addresultprorata_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addresultprorata"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_DistributesGainAndPersists(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// 2 acciones (holdings 1000 + 500 = 1500) + 1 indice que NON debe entrar.
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
	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 9999, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	// type=1 (Accion), mes=4, ano=2026, ganhanza total=300 (20% sobre 1500)
	r := bufio.NewReader(strings.NewReader("1\n4\n2026\n300\n"))
	var out bytes.Buffer
	if err := addresultprorata.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verifica resultados gardados:
	// AAPL: 1000 → 1200 (+200)
	// MSFT: 500 → 600 (+100)
	aaplRes, err := s.ListMonthlyResultsByAsset(aaplID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset AAPL: %v", err)
	}
	if len(aaplRes) != 1 || !almostEqual(aaplRes[0].ResultUSD, 1200) {
		t.Errorf("AAPL: got %+v, esperabamos result=1200", aaplRes)
	}

	msftRes, err := s.ListMonthlyResultsByAsset(msftID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset MSFT: %v", err)
	}
	if len(msftRes) != 1 || !almostEqual(msftRes[0].ResultUSD, 600) {
		t.Errorf("MSFT: got %+v, esperabamos result=600", msftRes)
	}

	// Confirma cabeceira na saída
	output := out.String()
	if !strings.Contains(output, "Suma de holdings: 1500.00 USD") {
		t.Errorf("saída non contén suma esperada:\n%s", output)
	}
	if !strings.Contains(output, "+20.00%") {
		t.Errorf("saída non contén pct esperado:\n%s", output)
	}
}
