package viewreport_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewreport"
)

func preto(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestBuild_NormalCase(t *testing.T) {
	repo := &fakeRepo{summaries: gainSummary()}
	report, err := viewreport.Build(repo, sampleAssets[0], 2026, 5)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.Asset.Name != "AAPL" || report.Period != (domain.YearMonth{Year: 2026, Month: 5}) {
		t.Fatalf("metadatos inesperados: %+v", report)
	}
	if !report.HasResult || !report.HasGainPct {
		t.Fatalf("flags inesperados: %+v", report)
	}
	if !preto(report.Gain, 300) || !preto(report.GainPct, 20) {
		t.Fatalf("métricas inesperadas: gain=%.2f pct=%.2f", report.Gain, report.GainPct)
	}
}

func TestBuild_PartialResultEquivalentSingleAsset(t *testing.T) {
	repo := &fakeRepo{summaries: map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {EstimatedHolding: 1300, Result: 1500, HasResult: true},
	}}
	report, err := viewreport.Build(repo, sampleAssets[0], 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !report.HasResult || !preto(report.Gain, 200) || !preto(report.GainPct, 15.38) {
		t.Fatalf("resultado inesperado: %+v", report)
	}
}

func TestBuild_NoResult(t *testing.T) {
	repo := &fakeRepo{summaries: map[summaryKey]domain.MonthlySummary{
		{10, 2026, 5}: {EstimatedHolding: 1500, HasResult: false},
	}}
	report, err := viewreport.Build(repo, sampleAssets[0], 2026, 5)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.HasResult || report.HasGainPct || report.Gain != 0 || report.GainPct != 0 {
		t.Fatalf("sen resultado debería deixar métricas baleiras: %+v", report)
	}
}
