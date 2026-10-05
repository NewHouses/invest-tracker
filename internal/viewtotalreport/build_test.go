package viewtotalreport_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalreport"
)

func preto(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestBuild_NormalCase(t *testing.T) {
	report, err := viewtotalreport.Build(gainSetup(), 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.Period != (domain.YearMonth{Year: 2026, Month: 4}) || report.TotalAssets != 2 || report.AssetsActive != 2 || report.AssetsWithResult != 2 || report.Partial {
		t.Fatalf("metadatos inesperados: %+v", report)
	}
	checks := map[string][2]float64{
		"TotalInvested":             {report.TotalInvested, 3200},
		"InvestedInMonth":           {report.InvestedInMonth, 200},
		"InvestedPlusPrevDividends": {report.InvestedPlusPrevDividends, 220},
		"Dividends":                 {report.Dividends, 50},
		"DividendsPrev":             {report.DividendsPrev, 20},
		"HoldingNoDiv":              {report.HoldingNoDiv, 3400},
		"HoldingWithDiv":            {report.HoldingWithDiv, 3420},
		"BaseNoDiv":                 {report.BaseNoDiv, 3400},
		"BaseWithDiv":               {report.BaseWithDiv, 3420},
		"ResultNoDiv":               {report.ResultNoDiv, 3650},
		"ResultWithDiv":             {report.ResultWithDiv, 3700},
		"GainNoDiv":                 {report.GainNoDiv, 250},
		"GainWithDiv":               {report.GainWithDiv, 280},
		"PctNoDiv":                  {report.PctNoDiv, 7.35},
		"PctWithDiv":                {report.PctWithDiv, 8.19},
	}
	for name, gotWant := range checks {
		if !preto(gotWant[0], gotWant[1]) {
			t.Fatalf("%s=%.2f, esperabamos %.2f", name, gotWant[0], gotWant[1])
		}
	}
	if !report.HasResults || !report.HasMetrics || !report.HasPctWithDiv {
		t.Fatalf("flags inesperados: %+v", report)
	}
	if report.Averages.Months != 2 || !preto(report.Averages.GainNoDiv, 225) || !preto(report.Averages.GainWithDiv, 250) {
		t.Fatalf("medias inesperadas: %+v", report.Averages)
	}
}

func TestBuild_PartialResults(t *testing.T) {
	repo := gainSetup()
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{
		TotalInvestedUpTo: 2000, EstimatedHolding: 2100, HasPrevResult: true,
	}
	report, err := viewtotalreport.Build(repo, 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !report.Partial || report.AssetsWithResult != 1 || report.AssetsActive != 2 {
		t.Fatalf("parcialidade inesperada: %+v", report)
	}
	if !preto(report.HoldingNoDiv, 3400) || !preto(report.BaseNoDiv, 1300) || !preto(report.ResultNoDiv, 1500) {
		t.Fatalf("base parcial inesperada: %+v", report)
	}
	if !preto(report.GainNoDiv, 200) || !preto(report.PctNoDiv, 15.38) || !preto(report.GainWithDiv, 230) || !preto(report.PctWithDiv, 17.42) {
		t.Fatalf("métricas parciais inesperadas: %+v", report)
	}
}

func TestBuild_NoResults(t *testing.T) {
	repo := gainSetup()
	repo.summaries[sumKey{10, 2026, 4}] = domain.MonthlySummary{TotalInvestedUpTo: 1200, EstimatedHolding: 1300}
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{TotalInvestedUpTo: 2000, EstimatedHolding: 2100}
	repo.months = map[divKey][]domain.YearMonth{}
	report, err := viewtotalreport.Build(repo, 2026, 4)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if report.HasResults || report.HasMetrics || report.HasPctWithDiv || report.Partial {
		t.Fatalf("flags sen resultado inesperados: %+v", report)
	}
	if !preto(report.HoldingNoDiv, 3400) || report.ResultNoDiv != 0 || report.GainNoDiv != 0 || report.Averages.Months != 0 {
		t.Fatalf("métricas sen resultado inesperadas: %+v", report)
	}
}

func TestBuild_AssetWithoutPrevMonthResult_UsesEstimatedHolding(t *testing.T) {
	repo := &fakeRepo{
		assets: twoAssets,
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 2}: {TotalInvestedUpTo: 1000, EstimatedHolding: 1000, HasPrevResult: true},
			{11, 2026, 2}: {TotalInvestedUpTo: 500, EstimatedHolding: 500, Result: 500, HasResult: true, HasPrevResult: true},
			{10, 2026, 3}: {TotalInvestedUpTo: 1000, EstimatedHolding: 1000, Result: 1100, HasResult: true, HasPrevResult: true},
			{11, 2026, 3}: {TotalInvestedUpTo: 500, EstimatedHolding: 500, Result: 510, HasResult: true, HasPrevResult: true},
		},
		months: map[divKey][]domain.YearMonth{
			{2026, 3}: {{Year: 2026, Month: 3}},
		},
	}
	report, err := viewtotalreport.Build(repo, 2026, 3)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !preto(report.HoldingNoDiv, 1500) || !preto(report.BaseNoDiv, 1500) || !preto(report.ResultNoDiv, 1610) {
		t.Fatalf("base esperada contra holding: %+v", report)
	}
	if !preto(report.GainNoDiv, 110) || !preto(report.PctNoDiv, 7.33) {
		t.Fatalf("G/P inflada ou pct incorrecto: %+v", report)
	}
}
