package viewtypereport_test

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtypereport"
)

func TestRun_EndToEnd_AggregatesByTypeFromDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	// 2 acciones + 1 indice (este non debe aparecer no informe de Acción)
	aaplID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset AAPL: %v", err)
	}
	msftID, err := s.InsertAsset(domain.Asset{
		Type: domain.Accion, Name: "MSFT", AmountUSD: 500, Month: 4, Year: 2026,
	})
	if err != nil {
		t.Fatalf("InsertAsset MSFT: %v", err)
	}
	if _, err := s.InsertAsset(domain.Asset{
		Type: domain.Indice, Name: "Vanguard", AmountUSD: 9999, Month: 3, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertAsset Vanguard: %v", err)
	}

	// Resultados en 04/2026 para os dous Acción
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: aaplID, ResultUSD: 1100, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertMonthlyResult AAPL: %v", err)
	}
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{
		AssetID: msftID, ResultUSD: 600, Month: 4, Year: 2026,
	}); err != nil {
		t.Fatalf("InsertMonthlyResult MSFT: %v", err)
	}

	// Input: type=1 (Accion), month=4, year=2026
	r := bufio.NewReader(strings.NewReader("1\n4\n2026\n"))
	var out bytes.Buffer
	if err := viewtypereport.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := out.String()

	// Cabeceira
	if !strings.Contains(output, "Tipo: Acción · 04/2026") {
		t.Errorf("saída non contén cabeceira:\n%s", output)
	}
	// Lista de activos de tipo (antes de pedir mes/ano)
	for _, want := range []string{"Activos de tipo Acción:", "- AAPL", "- MSFT"} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}
	if strings.Contains(output, "- Vanguard") {
		t.Errorf("saída non debería listar Vanguard:\n%s", output)
	}

	// Cifras agregadas (esperadas)
	// AAPL: TotalInvestedUpTo=1000, InvestedInMonth=0, holding=1000, result=1100
	// MSFT: TotalInvestedUpTo=500, InvestedInMonth=500, holding=500, result=600
	// Sumas: investido=1500, este mes=500, holding=1500, resultado=1700
	// gain=200, pct=200/1500*100 ≈ 13.33
	for _, want := range []string{
		"Activos incluídos: 2",
		"1500.00 USD", // investido ata o mes / no activo
		"500.00 USD",  // investido este mes
		"1700.00 USD", // resultado agregado
		"+200.00 USD", // ganhanza
		"+13.33%",     // índice
	} {
		if !strings.Contains(output, want) {
			t.Errorf("saída non contén %q:\n%s", want, output)
		}
	}

	// Vanguard non aparece nin cos seus números
	if strings.Contains(output, "9999") {
		t.Errorf("saída non debería incluír cifras de Vanguard:\n%s", output)
	}
}
