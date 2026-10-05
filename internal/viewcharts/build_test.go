package viewcharts_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewcharts"
)

func TestBuildDistribution_CalculaTotalEtiquetasEPorcentaxes(t *testing.T) {
	got, err := viewcharts.BuildDistribution(setup())
	if err != nil {
		t.Fatalf("BuildDistribution: %v", err)
	}
	if !got.HasAssets || got.Total != 4100 || len(got.Items) != 3 {
		t.Fatalf("distribución inesperada: %+v", got)
	}
	if got.Items[0].Label != "Acción — AAPL" || got.Items[0].Value != 1500 {
		t.Fatalf("primeiro item inesperado: %+v", got.Items[0])
	}
	if math.Abs(got.Items[0].Pct-36.58536585365854) > 0.000001 {
		t.Fatalf("pct = %v", got.Items[0].Pct)
	}
}

func TestBuildAssetEvolution_DevolveSeriesConGaps(t *testing.T) {
	repo := setup()
	repo.summaries[sumKey{10, 2026, 5}] = domain.MonthlySummary{TotalInvestedUpTo: 1500}

	got, err := viewcharts.BuildAssetEvolution(repo, repo.assets[0])
	if err != nil {
		t.Fatalf("BuildAssetEvolution: %v", err)
	}
	if got.Title != "Acción — AAPL" || len(got.Months) != 2 || len(got.Series) != 2 {
		t.Fatalf("gráfica inesperada: %+v", got)
	}
	if got.Series[0].Values[0] != 1100 || !math.IsNaN(got.Series[0].Values[1]) {
		t.Fatalf("valores de resultado inesperados: %v", got.Series[0].Values)
	}
	if got.Series[1].Values[1] != 1500 {
		t.Fatalf("aportes inesperados: %v", got.Series[1].Values)
	}
}

func TestBuildTypeAggregated_AgregaResultadosEAportes(t *testing.T) {
	got, err := viewcharts.BuildTypeAggregated(setup(), domain.Accion)
	if err != nil {
		t.Fatalf("BuildTypeAggregated: %v", err)
	}
	if !got.HasAssets || got.Title != "Tipo: Acción" || len(got.Series) != 2 {
		t.Fatalf("gráfica inesperada: %+v", got)
	}
	if got.Series[0].Values[0] != 1750 || got.Series[0].Values[1] != 2400 {
		t.Fatalf("resultado agregado inesperado: %v", got.Series[0].Values)
	}
	if got.Series[1].Values[0] != 1600 || got.Series[1].Values[1] != 2100 {
		t.Fatalf("aporte agregado inesperado: %v", got.Series[1].Values)
	}
}

func TestBuildAssetsOfType_CreaUnhaSeriePorActivoConResultado(t *testing.T) {
	repo := setup()
	repo.summaries[sumKey{11, 2026, 4}] = domain.MonthlySummary{TotalInvestedUpTo: 600}
	repo.summaries[sumKey{11, 2026, 5}] = domain.MonthlySummary{TotalInvestedUpTo: 600}

	got, err := viewcharts.BuildAssetsOfType(repo, domain.Accion)
	if err != nil {
		t.Fatalf("BuildAssetsOfType: %v", err)
	}
	if !got.HasAssets || len(got.Series) != 1 || got.Series[0].Label != "AAPL" {
		t.Fatalf("series inesperadas: %+v", got.Series)
	}
}

func TestBuildAllTypes_UsaOrdeEstableETenGaps(t *testing.T) {
	repo := setup()
	repo.summaries[sumKey{12, 2026, 5}] = domain.MonthlySummary{TotalInvestedUpTo: 2000}

	got, err := viewcharts.BuildAllTypes(repo)
	if err != nil {
		t.Fatalf("BuildAllTypes: %v", err)
	}
	if !got.HasAssets || len(got.Series) != 2 || got.Series[0].Label != "Acción" || got.Series[1].Label != "Índice" {
		t.Fatalf("series inesperadas: %+v", got.Series)
	}
	if got.Series[0].Values[0] != 1750 || got.Series[0].Values[1] != 2400 {
		t.Fatalf("valores Acción inesperados: %v", got.Series[0].Values)
	}
	if got.Series[1].Values[0] != 2100 || !math.IsNaN(got.Series[1].Values[1]) {
		t.Fatalf("valores Índice inesperados: %v", got.Series[1].Values)
	}
}

func TestBuildTotal_SumaResultadosEDividendosAcumulados(t *testing.T) {
	got, err := viewcharts.BuildTotal(setup())
	if err != nil {
		t.Fatalf("BuildTotal: %v", err)
	}
	if !got.HasAssets || got.Title != "Resultado total da carteira" || len(got.Series) != 1 {
		t.Fatalf("gráfica inesperada: %+v", got)
	}
	want := []float64{3860, 4630}
	for i := range want {
		if got.Series[0].Values[i] != want[i] {
			t.Fatalf("total[%d] = %v, queremos %v", i, got.Series[0].Values[i], want[i])
		}
	}
}

func TestBuilders_MarcanEstadoBaleiro(t *testing.T) {
	dist, err := viewcharts.BuildDistribution(&fakeRepo{})
	if err != nil {
		t.Fatalf("BuildDistribution: %v", err)
	}
	if dist.HasAssets || dist.Total != 0 || len(dist.Items) != 0 {
		t.Fatalf("distribución baleira inesperada: %+v", dist)
	}

	chart, err := viewcharts.BuildAllTypes(&fakeRepo{})
	if err != nil {
		t.Fatalf("BuildAllTypes: %v", err)
	}
	if chart.HasAssets || len(chart.Months) != 0 || len(chart.Series) != 0 {
		t.Fatalf("gráfica baleira inesperada: %+v", chart)
	}
}
