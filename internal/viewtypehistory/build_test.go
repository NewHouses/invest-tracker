package viewtypehistory_test

import (
	"math"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtypehistory"
)

func preto(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("got %.4f, want %.4f", got, want)
	}
}

func TestBuild_DevolveCamposCalculados(t *testing.T) {
	h, err := viewtypehistory.Build(gainSetup(), domain.Accion)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if h.Type != domain.Accion || h.AssetCount != 2 || len(h.Rows) != 2 {
		t.Fatalf("historia inesperada: %+v", h)
	}
	preto(t, h.TotalInvested, 1500)
	preto(t, h.AvgIndexPct, 13.3333)
	preto(t, h.AvgGain, 212.5)
	if !h.HasAverages {
		t.Fatal("esperabamos medias dispoñibles")
	}
	preto(t, h.TotalGain, 425)
	if !h.HasTotalGain {
		t.Fatal("esperabamos G/P total dispoñible")
	}

	row := h.Rows[1]
	if row.Period.Year != 2026 || row.Period.Month != 5 {
		t.Fatalf("periodo inesperado: %+v", row.Period)
	}
	preto(t, row.Aporte, 0)
	preto(t, row.Holding, 1650)
	preto(t, row.Result, 1925)
	preto(t, row.Gain, 275)
	preto(t, row.GainPct, 16.6667)
	if !row.HasMetrics {
		t.Fatal("esperabamos métricas na segunda fila")
	}
}

func TestBuild_SenActivosDoTipo(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{{ID: 12, Type: domain.Indice, Name: "Vanguard"}}}
	h, err := viewtypehistory.Build(repo, domain.Accion)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if h.AssetCount != 0 || len(h.Rows) != 0 {
		t.Fatalf("historia inesperada: %+v", h)
	}
}
