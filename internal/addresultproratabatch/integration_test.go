package addresultproratabatch_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/addresultproratabatch"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_DistributesAcrossMonthsAndPersists(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// 2 acciones (AAPL=1000, MSFT=500) compradas en 03/2026.
	// + 1 indice que NON debe contar para o reparto.
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

	// type=1 (Accion), mes inicial=4, ano=2026.
	// 04/2026: holdings AAPL=1000, MSFT=500, total=1500. Gain=300 (20%) →
	//          AAPL=1200, MSFT=600. Continuar.
	// 05/2026: holdings AAPL=1200 (prev_result), MSFT=600, total=1800. Gain=180
	//          (10%) → AAPL=1320, MSFT=660. Parar.
	r := bufio.NewReader(strings.NewReader("1\n4\n2026\n300\ns\n180\nn\n"))
	var out bytes.Buffer
	if err := addresultproratabatch.Run(r, &out, s); err != nil {
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

	aaplRes, err := s2.ListMonthlyResultsByAsset(aaplID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset AAPL: %v", err)
	}
	if len(aaplRes) != 2 {
		t.Fatalf("AAPL: got %d resultados, esperabamos 2", len(aaplRes))
	}
	wantAAPL := []struct {
		amount float64
		month  int
	}{
		{1200, 4},
		{1320, 5},
	}
	for i, w := range wantAAPL {
		if !almostEqual(aaplRes[i].ResultUSD, w.amount) || aaplRes[i].Month != w.month || aaplRes[i].Year != 2026 {
			t.Errorf("AAPL[%d] = %+v, esperabamos amount=%.2f month=%d", i, aaplRes[i], w.amount, w.month)
		}
	}

	msftRes, err := s2.ListMonthlyResultsByAsset(msftID)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset MSFT: %v", err)
	}
	if len(msftRes) != 2 {
		t.Fatalf("MSFT: got %d resultados, esperabamos 2", len(msftRes))
	}
	wantMSFT := []struct {
		amount float64
		month  int
	}{
		{600, 4},
		{660, 5},
	}
	for i, w := range wantMSFT {
		if !almostEqual(msftRes[i].ResultUSD, w.amount) || msftRes[i].Month != w.month || msftRes[i].Year != 2026 {
			t.Errorf("MSFT[%d] = %+v, esperabamos amount=%.2f month=%d", i, msftRes[i], w.amount, w.month)
		}
	}

	output := out.String()
	if !strings.Contains(output, "4 resultado(s) e saltáronse 0 mes(es)") {
		t.Errorf("saída non contén resumo correcto:\n%s", output)
	}
	if !strings.Contains(output, "Suma de holdings: 1500.00 USD") {
		t.Errorf("saída non contén suma de 04/2026:\n%s", output)
	}
	if !strings.Contains(output, "Suma de holdings: 1800.00 USD") {
		t.Errorf("saída non contén suma de 05/2026:\n%s", output)
	}
}
