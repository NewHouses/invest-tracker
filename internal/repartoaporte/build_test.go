package repartoaporte_test

import (
	"math"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
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

	got, err := repartoaporte.Allocate(1000, selection, i18n.Galician)
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

func TestAllocate_ConPorcentaxesDoUsuario(t *testing.T) {
	selection := []repartoaporte.TypeSelection{
		{Type: domain.Accion, Assets: []domain.Asset{{ID: 10, Name: "AAPL"}, {ID: 11, Name: "MSFT"}}, Percents: []float64{70, 20}},
		{Type: domain.Indice, Assets: []domain.Asset{{ID: 12, Name: "Vanguard"}}, Percents: []float64{10}},
	}

	got, err := repartoaporte.Allocate(1000, selection, i18n.Galician)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	acc, idx := got.Types[0], got.Types[1]
	if acc.Percent != 90 || acc.Amount != 900 || acc.Assets[0].Amount != 700 || acc.Assets[1].Amount != 200 || acc.Assets[0].Percent != 70 {
		t.Fatalf("Acción inesperado: %+v", acc)
	}
	if idx.Percent != 10 || idx.Amount != 100 || idx.Assets[0].Amount != 100 {
		t.Fatalf("Índice inesperado: %+v", idx)
	}
}

// Os importes van en centavos e suman sempre o total: os centavos que sobran
// ao redondear van a quen máis perdeu.
func TestAllocate_CentavosExactos(t *testing.T) {
	three := []domain.Asset{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}, {ID: 3, Name: "C"}}
	cases := []struct {
		name     string
		total    float64
		percents []float64
		want     []float64
	}{
		{"terzos a partes iguais", 100, nil, []float64{33.34, 33.33, 33.33}},
		{"porcentaxes con decimais", 10, []float64{33.33, 33.33, 33.34}, []float64{3.33, 3.33, 3.34}},
		{"o resto vai a quen máis perdeu", 1, []float64{10, 45, 45}, []float64{0.10, 0.45, 0.45}},
		{"activo ao 0 %", 99.99, []float64{0, 50, 50}, []float64{0, 50, 49.99}},
		{"33,33 × 3 reparte coma se sumasen 100", 1000, []float64{33.33, 33.33, 33.33}, []float64{333.34, 333.33, 333.33}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repartoaporte.Allocate(tc.total, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: three, Percents: tc.percents}}, i18n.Galician)
			if err != nil {
				t.Fatalf("Allocate: %v", err)
			}
			sum := 0.0
			for i, a := range got.Types[0].Assets {
				if math.Abs(a.Amount-tc.want[i]) > 1e-9 {
					t.Fatalf("importes = %+v, queremos %v", got.Types[0].Assets, tc.want)
				}
				sum += a.Amount
			}
			if math.Abs(sum-tc.total) > 1e-9 || math.Abs(got.Types[0].Amount-tc.total) > 1e-9 {
				t.Fatalf("suma %.2f / tipo %.2f, queremos %.2f", sum, got.Types[0].Amount, tc.total)
			}
		})
	}
}

func TestEqualPercents_PartesIguaisPorTipoEActivo(t *testing.T) {
	selection := []repartoaporte.TypeSelection{
		{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}},
		{Type: domain.Indice, Assets: []domain.Asset{{ID: 3}}},
		{Type: domain.Fondo, Assets: []domain.Asset{{ID: 4}}},
	}
	got := repartoaporte.EqualPercents(selection)
	want := [][]float64{{16.67, 16.67}, {33.33}, {33.33}}
	sum := 0.0
	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > 1e-12 {
				t.Fatalf("porcentaxes = %v, queremos %v", got, want)
			}
			sum += got[i][j]
		}
	}
	if !repartoaporte.SumsTo100(sum, 4) {
		t.Fatalf("suman %v", sum)
	}
}

// Ao escribir cada porcentaxe con dúas decimais perdese algo: admítese só
// esa diferenza.
func TestSumsTo100_AdmiteORedondeo(t *testing.T) {
	cases := []struct {
		sum  float64
		n    int
		want bool
	}{
		{100, 1, true},
		{99.99, 3, true},  // 33,33 × 3
		{100.02, 6, true}, // 16,67 × 6
		{99.98, 2, false}, // 49,99 × 2: falta unha centésima de verdade
		{99.9, 3, false},
		{100.5, 10, false},
	}
	for _, tc := range cases {
		if got := repartoaporte.SumsTo100(tc.sum, tc.n); got != tc.want {
			t.Errorf("SumsTo100(%v, %d) = %v, queremos %v", tc.sum, tc.n, got, tc.want)
		}
	}
}

// Os activos que aínda non se crearon (ID 0) entran no reparto do seu tipo e
// saen marcados como novos.
func TestAllocate_ActivosNovos(t *testing.T) {
	selection := []repartoaporte.TypeSelection{
		{Type: domain.Accion, Assets: []domain.Asset{{ID: 10, Name: "AAPL"}, {Type: domain.Accion, Name: "NVDA"}}},
		{Type: domain.Fondo, Assets: []domain.Asset{{Type: domain.Fondo, Name: "Novo fondo"}}},
	}
	got, err := repartoaporte.Allocate(1000, selection, i18n.Galician)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	acc, fondo := got.Types[0], got.Types[1]
	if acc.Assets[0].New || !acc.Assets[1].New || acc.Assets[1].Amount != 250 || acc.Assets[1].Asset.Name != "NVDA" {
		t.Fatalf("Acción inesperado: %+v", acc)
	}
	if !fondo.Assets[0].New || fondo.Amount != 500 || fondo.Label != "Fondo" {
		t.Fatalf("Fondo inesperado: %+v", fondo)
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
		{"porcentaxe por riba de 100", 100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1, Name: "AAPL"}}, Percents: []float64{101}}}, "a porcentaxe de AAPL debe estar entre 0 e 100"},
		{"porcentaxe negativa", 100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1, Name: "AAPL"}, {ID: 2, Name: "MSFT"}}, Percents: []float64{110, -10}}}, "entre 0 e 100"},
		{"non suman 100", 100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}, Percents: []float64{50, 40}}}, "as porcentaxes suman 90.00 %; teñen que sumar 100 %"},
		{"falta unha centésima", 100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}, Percents: []float64{49.99, 49.99}}}, "as porcentaxes suman 99.98 %"},
		{"falta unha porcentaxe", 100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}, Percents: []float64{100}}}, "indica a porcentaxe de cada activo"},
		{"un tipo sen porcentaxes", 100, []repartoaporte.TypeSelection{
			{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}}, Percents: []float64{100}},
			{Type: domain.Indice, Assets: []domain.Asset{{ID: 2}}},
		}, "indica a porcentaxe de cada activo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repartoaporte.Allocate(tc.total, tc.selection, i18n.Galician)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("erro = %v, queremos que conteña %q", err, tc.want)
			}
		})
	}
}

func TestAllocate_ErrosEnIngles(t *testing.T) {
	_, err := repartoaporte.Allocate(100, []repartoaporte.TypeSelection{{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}, Percents: []float64{50, 40}}}, i18n.English)
	if err == nil || err.Error() != "the percentages add up to 90.00%; they must add up to 100%" {
		t.Fatalf("erro = %v", err)
	}
}

func TestAllocate_PorcentaxesEfectivas(t *testing.T) {
	got, err := repartoaporte.Allocate(1000, []repartoaporte.TypeSelection{
		{Type: domain.Accion, Assets: []domain.Asset{{ID: 1}, {ID: 2}}, Percents: []float64{33.33, 33.33}},
		{Type: domain.Indice, Assets: []domain.Asset{{ID: 3}}, Percents: []float64{33.33}},
	}, i18n.Galician)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if p := got.Types[0].Assets[0].Percent; math.Abs(p-100.0/3) > 1e-5 {
		t.Fatalf("porcentaxe efectiva = %v, queremos 33,333...", p)
	}
	if p := got.Types[0].Percent; math.Abs(p-200.0/3) > 1e-5 {
		t.Fatalf("porcentaxe do tipo = %v", p)
	}
}
