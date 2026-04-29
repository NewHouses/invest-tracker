package closemonth_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsResultsForActiveAssets(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Activo 1: AAPL, 1000 USD investido en 03/2026 → activo en 04/2026.
	id1, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[1]: %v", err)
	}
	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: id1, AmountUSD: 500, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}

	// Activo 2: Vanguard, 2000 USD investido en 04/2026 → activo en 04/2026.
	id2, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[2]: %v", err)
	}

	// Activo 3: futuro, 500 USD en 06/2026 → NON activo en 04/2026.
	id3, err := s.InsertAsset(domain.Asset{
		Type: domain.Fondo, Name: "Futuro", AmountUSD: 500, Month: 6, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[3]: %v", err)
	}

	// Input: month=4, year=2026, resultado AAPL=1800, resultado Vanguard=2150
	// (Activo 3 non se pregunta porque non está investido aínda en 04/2026.)
	input := "4\n2026\n1800\n2150\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := closemonth.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Reabre e verifica
	s2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("re-Open: %v", err)
	}
	t.Cleanup(func() { _ = s2.Close() })

	// AAPL debe ter resultado 1800
	sumAAPL, err := s2.MonthlySummary(id1, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary AAPL: %v", err)
	}
	if !sumAAPL.HasResult || sumAAPL.Result != 1800 {
		t.Errorf("AAPL: HasResult=%v, Result=%v, esperabamos true/1800", sumAAPL.HasResult, sumAAPL.Result)
	}

	// Vanguard debe ter resultado 2150
	sumVan, err := s2.MonthlySummary(id2, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary Vanguard: %v", err)
	}
	if !sumVan.HasResult || sumVan.Result != 2150 {
		t.Errorf("Vanguard: HasResult=%v, Result=%v, esperabamos true/2150", sumVan.HasResult, sumVan.Result)
	}

	// Futuro NON debe ter resultado para 04/2026
	sumFut, err := s2.MonthlySummary(id3, 2026, 4)
	if err != nil {
		t.Fatalf("MonthlySummary Futuro: %v", err)
	}
	if sumFut.HasResult {
		t.Errorf("Futuro non debería ter resultado para 04/2026 (era inactivo): %+v", sumFut)
	}

	// Verifica saída terminal
	output := out.String()
	for _, want := range []string{
		"2 activo(s) en 04/2026",
		"AAPL",
		"Vanguard",
		"Pechouse 04/2026: 2 resultado(s) gardado(s), 0 saltado(s)",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Futuro") {
		t.Errorf("saída non debería pedir resultado de Futuro:\n%s", output)
	}
}
