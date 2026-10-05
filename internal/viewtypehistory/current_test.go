package viewtypehistory_test

import (
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtypehistory"
)

func TestCurrent_ActivosDoTipoCreadosDespois_DevolveNil(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 6}}, summaries: map[sumKey]domain.MonthlySummary{}, months: map[int64][]domain.YearMonth{}}
	row, err := viewtypehistory.Current(repo, domain.Accion, domain.YearMonth{Year: 2026, Month: 5})
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
		months:    map[int64][]domain.YearMonth{10: {{Year: 2026, Month: 5}}},
		summaries: map[sumKey]domain.MonthlySummary{},
	}
	row, err := viewtypehistory.Current(repo, domain.Accion, domain.YearMonth{Year: 2026, Month: 5})
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
		months:    map[int64][]domain.YearMonth{10: {{Year: 2026, Month: 6}}},
		summaries: map[sumKey]domain.MonthlySummary{},
	}
	row, err := viewtypehistory.Current(repo, domain.Accion, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}

func TestCurrent_MesAberto_FiltraTipoESumaActivosCreados(t *testing.T) {
	now := domain.YearMonth{Year: 2026, Month: 5}
	repo := &fakeRepo{
		assets: []domain.Asset{
			{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1},
			{ID: 11, Type: domain.Accion, Name: "MSFT", Year: 2026, Month: 5},
			{ID: 12, Type: domain.Indice, Name: "Vanguard", Year: 2026, Month: 1},
			{ID: 13, Type: domain.Accion, Name: "Futura", Year: 2026, Month: 6},
		},
		months: map[int64][]domain.YearMonth{
			10: {{Year: 2026, Month: 4}},
			11: {},
			12: {},
			13: {},
		},
		summaries: map[sumKey]domain.MonthlySummary{
			{10, 2026, 5}: {InvestedInMonth: 100, EstimatedHolding: 1100},
			{11, 2026, 5}: {InvestedInMonth: 300, EstimatedHolding: 300},
			{12, 2026, 5}: {InvestedInMonth: 999, EstimatedHolding: 999},
		},
	}
	row, err := viewtypehistory.Current(repo, domain.Accion, now)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row == nil {
		t.Fatal("esperabamos fila actual")
	}
	if row.Period != now || row.HasMetrics || row.Result != 0 || row.Gain != 0 || row.GainPct != 0 {
		t.Fatalf("fila inesperada: %#v", row)
	}
	preto(t, row.Aporte, 400)
	preto(t, row.Holding, 1400)
}

func TestCurrent_SenNadaQueMostrar_DevolveNil(t *testing.T) {
	repo := &fakeRepo{
		assets:    []domain.Asset{{ID: 10, Type: domain.Accion, Name: "AAPL", Year: 2026, Month: 1}},
		months:    map[int64][]domain.YearMonth{},
		summaries: map[sumKey]domain.MonthlySummary{{10, 2026, 5}: {}},
	}
	row, err := viewtypehistory.Current(repo, domain.Accion, domain.YearMonth{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if row != nil {
		t.Fatalf("esperabamos nil, got %#v", row)
	}
}
