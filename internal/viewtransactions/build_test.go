package viewtransactions_test

import (
	"reflect"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtransactions"
)

func TestBuildRows_OrdenaECalculaCampos(t *testing.T) {
	asset := domain.Asset{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026}
	txs := []domain.Transaction{
		{ID: 3, AssetID: 10, AmountUSD: 300, Month: 5, Year: 2026},
		{ID: 2, AssetID: 10, AmountUSD: -200, Month: 4, Year: 2026},
		{ID: 1, AssetID: 10, AmountUSD: 500, Month: 4, Year: 2026},
	}

	got := viewtransactions.BuildRows(asset, txs)
	want := []viewtransactions.Row{
		{IsInitial: true, Year: 2026, Month: 3, Amount: 1000},
		{ID: 1, Year: 2026, Month: 4, Amount: 500},
		{ID: 2, Year: 2026, Month: 4, IsVenda: true, Amount: 200},
		{ID: 3, Year: 2026, Month: 5, Amount: 300},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filas = %+v, queremos %+v", got, want)
	}
}

func TestComputeTotals_CompraVendaNeto(t *testing.T) {
	rows := []viewtransactions.Row{
		{Amount: 1000},
		{Amount: 500},
		{IsVenda: true, Amount: 200},
	}

	got := viewtransactions.ComputeTotals(rows)
	want := viewtransactions.Totals{Compra: 1500, Venda: 200, Neto: 1300}
	if got != want {
		t.Fatalf("totais = %+v, queremos %+v", got, want)
	}
}
