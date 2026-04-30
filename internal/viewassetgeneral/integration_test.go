package viewassetgeneral_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewassetgeneral"
)

func TestRun_EndToEnd_HistoryFromDB(t *testing.T) {
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
	// Outro activo (non debe afectar)
	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 9999, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: id1, AmountUSD: 200, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: id1, ResultUSD: 1500, Month: 4, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	// Selecciona o primeiro activo (AAPL)
	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := viewassetgeneral.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()

	if !strings.Contains(output, "Acción — AAPL") {
		t.Errorf("saída non contén nome do activo:\n%s", output)
	}
	if !strings.Contains(output, "2 mes(es) con resultado") {
		t.Errorf("saída non sinala 2 meses:\n%s", output)
	}

	// Lifetime: invested=1200, last result=1500 → gain=300, pct=25.00%
	for _, want := range []string{
		"1200.00 USD", // total invested lifetime
		"+300.00 USD", // total gain
		"+25.00%",     // total pct
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Avg: pct (10 + 15.38)/2 ≈ 12.69, gain (100+200)/2 = 150
	for _, want := range []string{
		"+12.69",
		"+150.00 USD",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Filas
	for _, want := range []string{
		"1000.00",
		"1100.00",
		"+100.00",
		"+10.00%",
		"1300.00",
		"1500.00",
		"+200.00",
		"+15.38%",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Vanguard non debe aparecer no contido
	if strings.Contains(output, "9999") {
		t.Errorf("saída non debería incluír Vanguard:\n%s", output)
	}
}
