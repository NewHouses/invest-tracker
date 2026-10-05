package viewassethistory_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassethistory"
)

func preto(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("got %.4f, want %.4f", got, want)
	}
}

func TestBuild_DevolveCamposCalculados(t *testing.T) {
	h, err := viewassethistory.Build(gainSetup(), sampleAssets[0])
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if h.Asset.ID != 10 || len(h.Rows) != 2 {
		t.Fatalf("historia inesperada: %+v", h)
	}
	preto(t, h.TotalInvested, 1700)
	preto(t, h.AvgIndexPct, 7.5)
	preto(t, h.AvgGain, 100)
	if !h.HasAverages {
		t.Fatal("esperabamos medias dispoñibles")
	}
	preto(t, h.TotalGain, 200)
	if !h.HasTotalGain {
		t.Fatal("esperabamos G/P total dispoñible")
	}

	row := h.Rows[0]
	if row.Period.Year != 2026 || row.Period.Month != 4 {
		t.Fatalf("periodo inesperado: %+v", row.Period)
	}
	preto(t, row.Aporte, 500)
	preto(t, row.Holding, 1500)
	preto(t, row.Result, 1800)
	preto(t, row.Gain, 300)
	preto(t, row.GainPct, 20)
	if !row.HasMetrics {
		t.Fatal("esperabamos métricas na primeira fila")
	}
}

func TestBuild_SenResultadosDevolveFilasBaleiras(t *testing.T) {
	repo := &fakeRepo{assets: sampleAssets, months: map[int64][]domain.YearMonth{}}
	h, err := viewassethistory.Build(repo, sampleAssets[0])
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(h.Rows) != 0 || h.HasAverages || h.HasTotalGain {
		t.Fatalf("historia baleira inesperada: %+v", h)
	}
}
