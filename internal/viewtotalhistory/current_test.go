package viewtotalhistory_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalhistory"
)

func TestCurrent_ActivosCreadosDespois_DevolveNil(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 6}}, summaries: map[sumKey]domain.MonthlySummary{}, dividends: map[divKey]float64{}}
	row, err := viewtotalhistory.Current(repo, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_MesXaTenResultado_DevolveNil(t *testing.T) {
	repo := &fakeRepo{
		assets:    []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}},
		months:    []domain.YearMonth{{Year: 2026, Month: 5}},
		summaries: map[sumKey]domain.MonthlySummary{},
		dividends: map[divKey]float64{},
	}
	row, err := viewtotalhistory.Current(repo, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_ResultadoPosterior_DevolveNil(t *testing.T) {
	repo := &fakeRepo{
		assets:    []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}},
		months:    []domain.YearMonth{{Year: 2026, Month: 6}},
		summaries: map[sumKey]domain.MonthlySummary{},
		dividends: map[divKey]float64{},
	}
	row, err := viewtotalhistory.Current(repo, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_MesAberto_SumaFondosEDividendos(t *testing.T) {
	now := domain.YearMonth{Year: 2026, Month: 5}
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1},
			{ID: 11, Type: domain.Indice, Name: "Vanguard", Year: 2026, Month: 5},
			{ID: 12, Type: domain.Fondo, Name: "Futuro", Year: 2026, Month: 6},
		},
		months: []domain.YearMonth{{Year: 2026, Month: 4}},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {InvestedInMonth: 100, EstimatedHolding: 1100},
			{11, 2026, 5}: {InvestedInMonth: 300, EstimatedHolding: 300},
		},
		dividends: map[divKey]float64{{2026, 5}: 25},
	}
	row, err := viewtotalhistory.Current(repo, now)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row == nil {
		t.Fatal("esperabamos fila actual")
	}
	if row.Period != now || row.HasMetrics || row.Result != 0 || row.Gain != 0 || row.GainPct != 0 {
		t.Fatalf("fila inesperada: %#v", row)
	}
	preto(t, row.Aporte, 375)
	preto(t, row.Fondos, 1400)
	preto(t, row.Dividends, 25)
}

func TestCurrent_SenNadaQueMostrar_DevolveNil(t *testing.T) {
	repo := &fakeRepo{
		assets:    []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}},
		summaries: map[sumKey]domain.MonthlySummary{{10, 2026, 5}: {}},
		dividends: map[divKey]float64{},
	}
	row, err := viewtotalhistory.Current(repo, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}
