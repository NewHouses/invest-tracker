package viewassethistory_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassethistory"
)

func TestCurrent_AssetCreatedAfterNow_DevolveNil(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Year: 2026, Month: 6}
	row, err := viewassethistory.Current(&fakeRepo{summaries: map[sumKey]domain.MonthlySummary{}, months: map[int64][]domain.YearMonth{}}, asset, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_MesXaTenResultado_DevolveNil(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}
	repo := &fakeRepo{months: map[int64][]domain.YearMonth{10: {{Year: 2026, Month: 5}}}, summaries: map[sumKey]domain.MonthlySummary{}}
	row, err := viewassethistory.Current(repo, asset, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_ResultadoPosterior_DevolveNil(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}
	repo := &fakeRepo{months: map[int64][]domain.YearMonth{10: {{Year: 2026, Month: 6}}}, summaries: map[sumKey]domain.MonthlySummary{}}
	row, err := viewassethistory.Current(repo, asset, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_MesAberto_DevolveAporteEHolding(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}
	now := domain.YearMonth{Year: 2026, Month: 5}
	repo := &fakeRepo{
		months: map[int64][]domain.YearMonth{10: {{Year: 2026, Month: 4}}},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {InvestedInMonth: 125, EstimatedHolding: 1125},
		},
	}
	row, err := viewassethistory.Current(repo, asset, now)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row == nil {
		t.Fatal("esperabamos fila actual")
	}
	if row.Period != now || row.HasMetrics || row.Result != 0 || row.Gain != 0 || row.GainPct != 0 {
		t.Fatalf("fila inesperada: %#v", row)
	}
	preto(t, row.Aporte, 125)
	preto(t, row.Holding, 1125)
}

func TestCurrent_SenNadaQueMostrar_DevolveNil(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}
	repo := &fakeRepo{months: map[int64][]domain.YearMonth{}, summaries: map[sumKey]domain.MonthlySummary{{10, 2026, 5}: {}}}
	row, err := viewassethistory.Current(repo, asset, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}
