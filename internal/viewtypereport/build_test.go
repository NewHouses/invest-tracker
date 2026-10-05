package viewtypereport_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtypereport"
)

func preto(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestBuild_NormalCase(t *testing.T) {
	repo := &fakeRepo{assets: mixedAssets, summaries: gainSummariesAccion()}
	report, err := viewtypereport.Build(repo, domain.Accion, 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.Type != domain.Accion || report.Period != (domain.YearMonth{Year: 2026, Month: 4}) {
		t.Fatalf("metadatos inesperados: %+v", report)
	}
	if len(report.Assets) != 3 || len(report.Active) != 3 || report.WithResult != 3 || report.Partial {
		t.Fatalf("cobertura inesperada: %+v", report)
	}
	if !preto(report.TotalInvested, 2300) || !preto(report.InvestedInMonth, 200) || !preto(report.Holding, 2300) {
		t.Fatalf("totais inesperados: %+v", report)
	}
	if !preto(report.ResultSum, 2600) || !preto(report.Gain, 300) || !preto(report.GainPct, 13.04) || !report.HasGainPct {
		t.Fatalf("resultado inesperado: %+v", report)
	}
}

func TestBuild_PartialResults(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {TotalInvestedUpTo: 1000, EstimatedHolding: 1000, Result: 1100, HasResult: true},
		{11, 2026, 4}: {TotalInvestedUpTo: 500, EstimatedHolding: 500, Result: 600, HasResult: true},
		{13, 2026, 4}: {TotalInvestedUpTo: 800, EstimatedHolding: 800},
	}
	repo := &fakeRepo{assets: mixedAssets, summaries: sums}
	report, err := viewtypereport.Build(repo, domain.Accion, 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !report.Partial || report.WithResult != 2 || len(report.Active) != 3 {
		t.Fatalf("parcialidade inesperada: %+v", report)
	}
	if !preto(report.ResultSum, 1700) || !preto(report.HoldingForResult, 1500) || !preto(report.Gain, 200) || !preto(report.GainPct, 13.33) {
		t.Fatalf("métricas parciais inesperadas: %+v", report)
	}
}

func TestBuild_NoResults(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {TotalInvestedUpTo: 1000, EstimatedHolding: 1000},
		{11, 2026, 4}: {TotalInvestedUpTo: 500, EstimatedHolding: 500},
		{13, 2026, 4}: {TotalInvestedUpTo: 800, EstimatedHolding: 800},
	}
	repo := &fakeRepo{assets: mixedAssets, summaries: sums}
	report, err := viewtypereport.Build(repo, domain.Accion, 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.WithResult != 0 || report.Partial || report.HasGainPct || report.ResultSum != 0 || report.Gain != 0 {
		t.Fatalf("sen resultados debería deixar métricas baleiras: %+v", report)
	}
}
