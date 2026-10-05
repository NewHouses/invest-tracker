package closemonth_test

import (
	"errors"
	"testing"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
)

func TestEligible_FiltraMantendoOrdeEInclueResultadoExistente(t *testing.T) {
	sums := map[summaryKey]domain.MonthlySummary{
		{10, 2026, 4}: {EstimatedHolding: 1500, Result: 1700, HasResult: true},
		{11, 2026, 4}: {EstimatedHolding: 0},
		{12, 2026, 4}: {EstimatedHolding: 800},
	}
	repo := &fakeRepo{assets: threeAssets, summaries: sums}

	got, err := closemonth.Eligible(repo, 2026, 4)
	if err != nil {
		t.Fatalf("Eligible: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("elixibles = %+v, queremos 2", got)
	}
	if got[0].Asset.ID != 10 || got[0].Holding != 1500 || got[0].Result != 1700 || !got[0].HasResult {
		t.Fatalf("primeiro elixible inesperado: %+v", got[0])
	}
	if got[1].Asset.ID != 12 || got[1].Holding != 800 || got[1].HasResult {
		t.Fatalf("segundo elixible inesperado: %+v", got[1])
	}
}

func TestEligible_PropagaErroDoResumo(t *testing.T) {
	repo := &fakeRepo{assets: threeAssets, summaries: map[summaryKey]domain.MonthlySummary{}, sumEr: errors.New("boom")}

	_, err := closemonth.Eligible(repo, 2026, 4)
	if err == nil {
		t.Fatal("esperabamos erro")
	}
}
