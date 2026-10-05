package repartoaporte_test

import (
	"math"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
)

func TestAvailableTypes_DevolveTiposPresentesEnOrdeEstable(t *testing.T) {
	assets := []domain.Asset{
		{ID: 1, Type: domain.Fondo},
		{ID: 2, Type: domain.Accion},
		{ID: 3, Type: domain.CopyTrading},
		{ID: 4, Type: domain.Accion},
	}

	got := repartoaporte.AvailableTypes(assets)
	want := []domain.AssetType{domain.Accion, domain.CopyTrading, domain.Fondo}
	if len(got) != len(want) {
		t.Fatalf("tipos = %v, queremos %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tipos = %v, queremos %v", got, want)
		}
	}
}

func TestAllocate_RepartePorTipoEActivo(t *testing.T) {
	selection := []repartoaporte.TypeSelection{
		{Type: domain.Accion, Assets: []domain.Asset{{ID: 10, Name: "AAPL"}, {ID: 11, Name: "MSFT"}}},
		{Type: domain.Indice, Assets: []domain.Asset{{ID: 12, Name: "Vanguard"}}},
	}

	got, err := repartoaporte.Allocate(1000, selection)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if got.Total != 1000 || len(got.Types) != 2 {
		t.Fatalf("reparto inesperado: %+v", got)
	}
	if got.Types[0].Type != domain.Accion || got.Types[0].Label != "Acción" || got.Types[0].Amount != 500 {
		t.Fatalf("tipo Acción inesperado: %+v", got.Types[0])
	}
	if got.Types[0].Assets[0].Amount != 250 || got.Types[0].Assets[1].Amount != 250 {
		t.Fatalf("activos Acción inesperados: %+v", got.Types[0].Assets)
	}
	if got.Types[1].Type != domain.Indice || got.Types[1].Amount != 500 || got.Types[1].Assets[0].Amount != 500 {
		t.Fatalf("tipo Índice inesperado: %+v", got.Types[1])
	}
}

func TestAllocate_ErrosDeValidacion(t *testing.T) {
	cases := []struct {
		name      string
		total     float64
		selection []repartoaporte.TypeSelection
		want      string
	}{
		{"total cero", 0, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}}}}, "maior ca 0"},
		{"total NaN", math.NaN(), []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}}}}, "maior ca 0"},
		{"sen selección", 100, nil, "polo menos un tipo"},
		{"sen activos", 100, []repartoaporte.TypeSelection{{Type: domain.Accion}}, "non ten activos"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repartoaporte.Allocate(tc.total, tc.selection)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("erro = %v, queremos que conteña %q", err, tc.want)
			}
		})
	}
}
