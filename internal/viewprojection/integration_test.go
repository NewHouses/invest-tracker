package viewprojection_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewprojection"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// A data de inicio reutilízase do activo máis antigo, non do primeiro inserido.
func TestRun_EndToEnd_StartDateFromOldestAsset(t *testing.T) {
	s := openStore(t)

	for _, a := range []domain.Asset{
		{Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 7, Year: 2026},
		{Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 3, Year: 2025},
	} {
		if _, err := s.InsertAsset(a); err != nil {
			t.Fatalf("InsertAsset %s: %v", a.Name, err)
		}
	}

	r := bufio.NewReader(strings.NewReader("40000\n1\n"))
	var out bytes.Buffer
	if err := viewprojection.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Data de inicio dos investimentos: 03/2025 (activo máis antigo)",
		// 03/2025 + 239 meses = 02/2045.
		"03/2025 → 02/2045",
		"240 meses",
		"Capital final",
		"Capital total",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(output))
		}
	}
}

func TestRun_EndToEnd_NoAssets_PromptsForStartDate(t *testing.T) {
	s := openStore(t)

	// mes, ano, salario, retorno.
	r := bufio.NewReader(strings.NewReader("1\n2026\n40000\n1\n"))
	var out bytes.Buffer
	if err := viewprojection.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Aínda non hai activos: indica a data de inicio dos investimentos.",
		"01/2026 → 12/2045",
		"$269100.00", // salario anual final: 40000 · 1.1^20
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, truncate(output))
		}
	}
}

func TestFirstInvestmentMonth_EmptyStore(t *testing.T) {
	s := openStore(t)

	ym, ok, err := s.FirstInvestmentMonth()
	if err != nil {
		t.Fatalf("FirstInvestmentMonth: %v", err)
	}
	if ok {
		t.Errorf("esperabamos ok=false sen activos, got %+v", ym)
	}
}
