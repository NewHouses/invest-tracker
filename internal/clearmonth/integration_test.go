package clearmonth_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/clearmonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_DeletesAllResultsForMonth(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	id1, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[1]: %v", err)
	}
	id2, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 1, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset[2]: %v", err)
	}

	// 04/2026: 2 resultados (un por activo). 05/2026: 1 resultado de AAPL.
	for _, mr := range []domain.MonthlyResult{
		{AssetID: id1, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: id2, ResultUSD: 2100, Month: 4, Year: 2026},
		{AssetID: id1, ResultUSD: 1200, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	// Limpar 04/2026
	r := bufio.NewReader(strings.NewReader("4\n2026\n"))
	var out bytes.Buffer
	if err := clearmonth.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✓ Limpado mes 04/2026: 2 resultado(s) eliminado(s).") {
		t.Errorf("saída non contén a confirmación esperada:\n%s", output)
	}

	// 04/2026 baleiro
	if got, _ := s.ListMonthlyResultsByAsset(id1); len(got) != 1 || got[0].Month != 5 {
		t.Errorf("AAPL debería ter só o resultado de 05/2026, got %+v", got)
	}
	if got, _ := s.ListMonthlyResultsByAsset(id2); len(got) != 0 {
		t.Errorf("Vanguard debería non ter resultados, got %+v", got)
	}
}

func TestRun_EndToEnd_NoResultsToDelete(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	r := bufio.NewReader(strings.NewReader("4\n2026\n"))
	var out bytes.Buffer
	if err := clearmonth.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !strings.Contains(out.String(), "Non había resultados rexistrados en 04/2026.") {
		t.Errorf("saída non contén suxestión esperada:\n%s", out.String())
	}
}
