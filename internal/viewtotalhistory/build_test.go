package viewtotalhistory_test

import (
	"math"
	"path/filepath"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtotalhistory"
)

func preto(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("got %.4f, want %.4f", got, want)
	}
}

func TestBuild_DevolveCamposCalculados(t *testing.T) {
	h, err := viewtotalhistory.Build(gainSetup())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if h.AssetCount != 2 || len(h.Rows) != 2 {
		t.Fatalf("historia inesperada: %+v", h)
	}
	preto(t, h.LifetimeAporte, 3200)
	preto(t, h.AvgIndexPct, 8.0756)
	preto(t, h.AvgGain, 260)
	if !h.HasAverages {
		t.Fatal("esperabamos medias dispoñibles")
	}
	preto(t, h.TotalDividends, 70)
	preto(t, h.CurrentValue, 3720)
	if !h.HasCurrentValue {
		t.Fatal("esperabamos valor actual dispoñible")
	}
	preto(t, h.TotalGain, 520)
	if !h.HasTotalGain {
		t.Fatal("esperabamos G/P total dispoñible")
	}

	row := h.Rows[0]
	preto(t, row.Aporte, 2980)
	preto(t, row.Fondos, 3000)
	preto(t, row.Dividends, 20)
	preto(t, row.Result, 3220)
	preto(t, row.Gain, 220)
	preto(t, row.GainPct, 7.3333)
	if !row.HasMetrics {
		t.Fatal("esperabamos métricas na primeira fila")
	}
}

func TestBuild_EndToEnd_ValorActualConUltimoResultadoPorActivo(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	aID, err := s.InsertAsset(domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Month: 1, Year: 2026})
	if err != nil {
		t.Fatalf("InsertAsset A: %v", err)
	}
	bID, err := s.InsertAsset(domain.Asset{Type: domain.Accion, Name: "B", AmountUSD: 500, Month: 1, Year: 2026})
	if err != nil {
		t.Fatalf("InsertAsset B: %v", err)
	}
	for _, mr := range []domain.MonthlyResult{
		{AssetID: aID, ResultUSD: 1000, Month: 1, Year: 2026},
		{AssetID: bID, ResultUSD: 500, Month: 1, Year: 2026},
		{AssetID: bID, ResultUSD: 500, Month: 2, Year: 2026},
		{AssetID: aID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: bID, ResultUSD: 510, Month: 3, Year: 2026},
	} {
		if _, err := s.InsertMonthlyResult(mr); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}

	h, err := viewtotalhistory.Build(s)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if h.AssetCount != 2 || len(h.Rows) != 3 {
		t.Fatalf("historia inesperada: %+v", h)
	}
	row := h.Rows[2]
	if row.Period.Year != 2026 || row.Period.Month != 3 {
		t.Fatalf("periodo inesperado: %+v", row.Period)
	}
	preto(t, row.Aporte, 0)
	preto(t, row.Fondos, 1500)
	preto(t, row.Result, 1610)
	preto(t, row.Gain, 110)
	preto(t, row.GainPct, 7.3333)
	preto(t, h.AvgIndexPct, 2.4444)
	preto(t, h.AvgGain, 36.6667)
	preto(t, h.TotalGain, 110)
	preto(t, h.CurrentValue, 1610)
	if !h.HasCurrentValue {
		t.Fatal("esperabamos valor actual dispoñible")
	}
}
