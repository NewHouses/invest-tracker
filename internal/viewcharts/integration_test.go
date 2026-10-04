package viewcharts_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewcharts"
)

func TestRun_EndToEnd_Distribution_AndTotal(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// 2 acciones + 1 indice. Resultados en 04 e 05/2026. Dividendos.
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
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1100, Month: 4, Year: 2026},
		{AssetID: msftID, ResultUSD: 550, Month: 4, Year: 2026},
		{AssetID: vanID, ResultUSD: 2100, Month: 4, Year: 2026},
		{AssetID: aaplID, ResultUSD: 1200, Month: 5, Year: 2026},
		{AssetID: msftID, ResultUSD: 605, Month: 5, Year: 2026},
		{AssetID: vanID, ResultUSD: 2150, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}
	for _, d := range []domain.Dividend{
		{AmountUSD: 20, Month: 4, Year: 2026},
		{AmountUSD: 50, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertDividend(d); err != nil {
			t.Fatalf("InsertDividend: %v", err)
		}
	}

	// Probar chart 1 (distribución) seguido de chart 6 (total) e despois Voltar.
	r := bufio.NewReader(strings.NewReader("1\n6\n0\n"))
	var out bytes.Buffer
	if err := viewcharts.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		// Chart 1 (distribución): aportes lifetime AAPL=1000, MSFT=500, Vanguard=2000 → total 3500
		"Distribución de aportes",
		"$3500.00", // total aportado
		"AAPL", "MSFT", "Vanguard",
		// Chart 6 (total): mes 04 e 05 con dividendos acumulados
		"Evolución do resultado total",
		"Resultado total da carteira",
		"Resultado + dividendos acum.",
		"04/2026",
		"05/2026",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
