package viewassethistory_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassethistory"
)

type fakeRepo struct {
	assets  []domain.Asset
	reports map[int64]domain.AssetReport
	listEr  error
	repEr   error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) AssetReport(id int64) (domain.AssetReport, error) {
	if f.repEr != nil {
		return domain.AssetReport{}, f.repEr
	}
	return f.reports[id], nil
}

func runWith(assets []domain.Asset, reports map[int64]domain.AssetReport, input string) (string, error) {
	repo := &fakeRepo{assets: assets, reports: reports}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewassethistory.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

func gainReport() map[int64]domain.AssetReport {
	rows := []domain.AssetReportRow{
		{Year: 2026, Month: 4, InvestedInMonth: 500, TotalInvestedUpTo: 1500, Result: 1800, Gain: 300, GainPct: 20, HasGainPct: true},
		{Year: 2026, Month: 5, InvestedInMonth: 200, TotalInvestedUpTo: 1700, Result: 1900, Gain: 200, GainPct: 11.7647, HasGainPct: true},
	}
	return map[int64]domain.AssetReport{
		10: {
			Rows:               rows,
			TotalInvested:      1700,
			TotalGain:          200,
			HasTotalGain:       true,
			AvgMonthlyIndexPct: 15.8824,
			HasAvgIndex:        true,
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Historial dun activo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsSelection(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona (1-2):") {
		t.Errorf("saída non contén o prompt:\n%s", out)
	}
}

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_NoResults_PrintsHint(t *testing.T) {
	reports := map[int64]domain.AssetReport{
		10: {TotalInvested: 1000},
	}
	out, err := runWith(sampleAssets, reports, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai resultados rexistrados para este activo") {
		t.Errorf("saída non contén a suxestión sobre resultados:\n%s", out)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, err := runWith(sampleAssets, gainReport(), "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_PrintsHeader_AssetName(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Acción — AAPL") {
		t.Errorf("saída non contén o nome do activo:\n%s", out)
	}
}

func TestRun_PrintsAggregateStats(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Total investido",
		"1700.00 USD",
		"Total Ganhanzas/Perdas",
		"+200.00 USD",
		"Índice medio mensual",
		"+15.88",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableColumns(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Ano",
		"Mes",
		"Investido ata o mes",
		"Investido este mes",
		"Índice",
		"G/P USD",
		"Resultado",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsRowValues(t *testing.T) {
	out, err := runWith(sampleAssets, gainReport(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"1500.00",
		"500.00",
		"+20.00%",
		"+300.00",
		"1800.00",
		"1700.00",
		"200.00",
		"+11.76%",
		"+200.00",
		"1900.00",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_ShowsLossRow(t *testing.T) {
	rows := []domain.AssetReportRow{
		{Year: 2026, Month: 4, InvestedInMonth: 1000, TotalInvestedUpTo: 1000, Result: 800, Gain: -200, GainPct: -20, HasGainPct: true},
	}
	reports := map[int64]domain.AssetReport{
		10: {
			Rows:               rows,
			TotalInvested:      1000,
			TotalGain:          -200,
			HasTotalGain:       true,
			AvgMonthlyIndexPct: -20,
			HasAvgIndex:        true,
		},
	}
	out, err := runWith(sampleAssets, reports, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "-200.00") {
		t.Errorf("saída non mostra perda absoluta:\n%s", out)
	}
	if !strings.Contains(out, "-20.00%") {
		t.Errorf("saída non mostra perda %%:\n%s", out)
	}
}

func TestRun_HandlesZeroInvestedRow(t *testing.T) {
	rows := []domain.AssetReportRow{
		{Year: 2026, Month: 4, InvestedInMonth: 0, TotalInvestedUpTo: 0, Result: 100, Gain: 100, GainPct: 0, HasGainPct: false},
	}
	reports := map[int64]domain.AssetReport{
		10: {
			Rows:          rows,
			TotalInvested: 0,
			HasTotalGain:  true,
			TotalGain:     100,
			HasAvgIndex:   false,
		},
	}
	out, err := runWith(sampleAssets, reports, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "n/a") {
		t.Errorf("saída non contén n/a para o índice:\n%s", out)
	}
}
