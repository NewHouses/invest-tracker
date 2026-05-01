package viewtypehistory_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtypehistory"
)

func TestRun_EndToEnd_AggregatesByTypeFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// 2 acciones cuxos resultados se agregarán + 1 índice que NON debe contar.
	//   AAPL bought 03/2026 1000; tx 04 +500.
	//   MSFT bought 03/2026 500.
	//   Vanguard (Indice) bought 03/2026 9999 — ignorado.
	//
	//   Resultados:
	//     04 AAPL=1800 (holding=1500, gain=+300, +20%)
	//     04 MSFT=550  (holding=500,  gain=+50,  +10%)
	//     → agg 04: aporte=2000 (1000+500+500), holding=2000, result=2350, +350, +17.50%
	//
	//     05 AAPL=1800 (holding=1800, gain=0, 0%) — sen tx en 05
	//     05 MSFT=605  (holding=550,  gain=+55, +10%)
	//     → agg 05: aporte=0, holding=2350, result=2405, +55, ≈+2.34%
	//
	//   Lifetime invested = 1000+500+500 = 2000. Lifetime result = 1800+605 = 2405.
	//   Lifetime gain = +405.
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
	if _, err := s.InsertTransaction(domain.Transaction{
		AssetID: aaplID, AmountUSD: 500, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aaplID, ResultUSD: 1800, Month: 4, Year: 2026},
		{AssetID: msftID, ResultUSD: 550, Month: 4, Year: 2026},
		{AssetID: aaplID, ResultUSD: 1800, Month: 5, Year: 2026},
		{AssetID: msftID, ResultUSD: 605, Month: 5, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	// type=1 (Acción)
	r := bufio.NewReader(strings.NewReader("1\n"))
	var out bytes.Buffer
	if err := viewtypehistory.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Reporte histórico dun tipo",
		"Tipo: Acción · 2 activo(s) · 2 mes(es)",
		// Top summary
		"Total Aportado",
		"2000.00 USD",
		"Total Gañanzas/Perdas",
		"+405.00 USD",
		// Table columns
		"Aporte Mensual",
		"No activo",
		"G/P USD",
		"Resultado",
		// Mes 04 agg: aporte=2000, holding=2000, +17.50%, +350.00, result=2350
		"+17.50%", "+350.00", "2350.00",
		// Mes 05 agg: aporte=0, holding=2350, +2.34%, +55.00, result=2405
		"0.00", "2350.00", "+55.00", "2405.00",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
}
