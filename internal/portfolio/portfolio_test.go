package portfolio_test

import (
	"math"
	"path/filepath"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/portfolio"
	"invest-tracker/internal/store"
	"invest-tracker/internal/viewtotalhistory"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("abrindo a base de datos: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func insertAsset(t *testing.T, st *store.Store, a domain.Asset) domain.Asset {
	t.Helper()
	id, err := st.InsertAsset(a)
	if err != nil {
		t.Fatalf("gardando activo %s: %v", a.Name, err)
	}
	a.ID = id
	return a
}

func insertResult(t *testing.T, st *store.Store, asset domain.Asset, year, month int, value float64) {
	t.Helper()
	if _, err := st.InsertMonthlyResult(domain.MonthlyResult{AssetID: asset.ID, Year: year, Month: month, ResultUSD: value}); err != nil {
		t.Fatalf("gardando resultado de %s en %02d/%d: %v", asset.Name, month, year, err)
	}
}

func insertTransaction(t *testing.T, st *store.Store, asset domain.Asset, year, month int, amount float64) {
	t.Helper()
	if _, err := st.InsertTransaction(domain.Transaction{AssetID: asset.ID, Year: year, Month: month, AmountUSD: amount}); err != nil {
		t.Fatalf("gardando transacción de %s en %02d/%d: %v", asset.Name, month, year, err)
	}
}

func TestBuildRowsAndPending(t *testing.T) {
	st := openStore(t)
	a := insertAsset(t, st, domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Year: 2026, Month: 1})
	b := insertAsset(t, st, domain.Asset{Type: domain.Fondo, Name: "Sen resultado", AmountUSD: 500, Year: 2026, Month: 1})
	c := insertAsset(t, st, domain.Asset{Type: domain.Indice, Name: "Vendido", AmountUSD: 1000, Year: 2026, Month: 1})
	d := insertAsset(t, st, domain.Asset{Type: domain.CopyTrading, Name: "Futuro", AmountUSD: 700, Year: 2026, Month: 5})

	insertResult(t, st, a, 2026, 1, 1050)
	insertTransaction(t, st, a, 2026, 2, 200)
	insertResult(t, st, a, 2026, 2, 1300)
	insertResult(t, st, a, 2026, 4, 1400)
	insertResult(t, st, c, 2026, 1, 1000)
	insertTransaction(t, st, c, 2026, 2, -1200)
	insertResult(t, st, c, 2026, 4, 10)

	got, err := portfolio.Build(st, domain.YearMonth{Year: 2026, Month: 4})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got.Period.Year != 2026 || got.Period.Month != 4 {
		t.Fatalf("período=%#v, esperabamos 04/2026", got.Period)
	}
	if len(got.Rows) != 4 {
		t.Fatalf("filas=%d, esperabamos 4: %#v", len(got.Rows), got.Rows)
	}
	if got.Rows[0].Asset.ID != a.ID || got.Rows[1].Asset.ID != b.ID || got.Rows[2].Asset.ID != c.ID || got.Rows[3].Asset.ID != d.ID {
		t.Fatalf("orde inesperada: %#v", got.Rows)
	}

	rowA := got.Rows[0]
	if rowA.TotalInvested != 1200 || rowA.CurrentValue != 1400 || !rowA.HasCurrentValue || rowA.Gain != 200 || !rowA.HasGain || !rowA.HasGainPct || !closeTo(rowA.GainPct, 16.6666666667) {
		t.Fatalf("fila A inesperada: %#v", rowA)
	}
	if rowA.LastResult == nil || rowA.LastResult.Year != 2026 || rowA.LastResult.Month != 4 {
		t.Fatalf("último resultado de A inesperado: %#v", rowA.LastResult)
	}

	rowB := got.Rows[1]
	if rowB.HasCurrentValue || rowB.HasGain || rowB.LastResult != nil {
		t.Fatalf("o activo sen resultados non debe ter métricas: %#v", rowB)
	}
	if !rowB.Pending {
		t.Fatalf("o activo elegible sen resultado debía quedar pendente: %#v", rowB)
	}

	rowC := got.Rows[2]
	if rowC.TotalInvested >= 0 || rowC.HasGainPct {
		t.Fatalf("un total investido non positivo non debe ter porcentaxe: %#v", rowC)
	}
	if rowC.Pending {
		t.Fatalf("o activo con resultado no mes actual non debe estar pendente: %#v", rowC)
	}

	rowD := got.Rows[3]
	if rowD.Pending {
		t.Fatalf("o activo creado despois do período non debe estar pendente: %#v", rowD)
	}
	if got.PendingCount != 1 {
		t.Fatalf("pendentes=%d, esperabamos 1", got.PendingCount)
	}
}

func TestBuildEmptyRowsAreNotNil(t *testing.T) {
	got, err := portfolio.Build(openStore(t), domain.YearMonth{Year: 2026, Month: 4})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if got.Rows == nil || len(got.Rows) != 0 {
		t.Fatalf("filas baleiras=%#v, esperabamos []", got.Rows)
	}
}

func TestBuildConsistentWithTotalHistory(t *testing.T) {
	st := openStore(t)
	a := insertAsset(t, st, domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Year: 2026, Month: 1})
	b := insertAsset(t, st, domain.Asset{Type: domain.Fondo, Name: "B", AmountUSD: 500, Year: 2026, Month: 1})
	insertResult(t, st, a, 2026, 1, 1100)
	insertResult(t, st, b, 2026, 1, 510)
	insertResult(t, st, a, 2026, 4, 1200)
	insertResult(t, st, b, 2026, 4, 550)
	if _, err := st.InsertDividend(domain.Dividend{AmountUSD: 10, Year: 2026, Month: 1}); err != nil {
		t.Fatalf("gardando dividendo 01/2026: %v", err)
	}
	if _, err := st.InsertDividend(domain.Dividend{AmountUSD: 15, Year: 2026, Month: 4}); err != nil {
		t.Fatalf("gardando dividendo 04/2026: %v", err)
	}

	port, err := portfolio.Build(st, domain.YearMonth{Year: 2026, Month: 4})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	total, err := viewtotalhistory.Build(st)
	if err != nil {
		t.Fatalf("viewtotalhistory.Build: %v", err)
	}
	var currentValue, gain float64
	for _, row := range port.Rows {
		currentValue += row.CurrentValue
		gain += row.Gain
	}
	if !closeTo(currentValue+total.TotalDividends, total.CurrentValue) {
		t.Fatalf("valor actual inconsistente: carteira %.2f + dividendos %.2f, total %.2f", currentValue, total.TotalDividends, total.CurrentValue)
	}
	if !closeTo(gain+total.TotalDividends, total.TotalGain) {
		t.Fatalf("ganancia inconsistente: carteira %.2f + dividendos %.2f, total %.2f", gain, total.TotalDividends, total.TotalGain)
	}
}

func closeTo(got, want float64) bool {
	return math.Abs(got-want) < 0.000001
}

// O valor neto e a G/P teñen que ter en conta os movementos posteriores ao
// último resultado: un aporte do mes aínda sen pechar non é unha perda, un
// activo vendido xa non vale nada e uns restos de redondeo non deixan o
// activo pendente de pechar para sempre.
func TestBuildUsesMovementsAfterLastResult(t *testing.T) {
	st := openStore(t)
	aporte := insertAsset(t, st, domain.Asset{Type: domain.Indice, Name: "Aporte do mes", AmountUSD: 1000, Year: 2026, Month: 1})
	vendido := insertAsset(t, st, domain.Asset{Type: domain.Accion, Name: "Vendido", AmountUSD: 1000, Year: 2026, Month: 1})
	novo := insertAsset(t, st, domain.Asset{Type: domain.Fondo, Name: "Sen resultado", AmountUSD: 300, Year: 2026, Month: 10})
	restos := insertAsset(t, st, domain.Asset{Type: domain.CopyTrading, Name: "Restos", AmountUSD: 100, Year: 2026, Month: 1})

	insertResult(t, st, aporte, 2026, 9, 1200)
	insertTransaction(t, st, aporte, 2026, 10, 500)
	insertResult(t, st, vendido, 2026, 8, 1100)
	insertTransaction(t, st, vendido, 2026, 9, -1150)
	insertResult(t, st, restos, 2026, 8, 100.004)
	insertTransaction(t, st, restos, 2026, 9, -100)

	got, err := portfolio.Build(st, domain.YearMonth{Year: 2026, Month: 10})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	rows := make(map[int64]portfolio.Row, len(got.Rows))
	for _, row := range got.Rows {
		rows[row.Asset.ID] = row
	}

	if r := rows[aporte.ID]; !closeTo(r.TotalInvested, 1500) || !closeTo(r.CurrentValue, 1700) || !closeTo(r.Gain, 200) || !r.HasGain || !closeTo(r.GainPct, 200.0/1500*100) || !r.Pending {
		t.Fatalf("o aporte do mes non pode contar como perda: %#v", r)
	}
	if r := rows[vendido.ID]; r.CurrentValue != 0 || !r.HasCurrentValue || !r.HasGain || !closeTo(r.Gain, 150) || r.HasGainPct || r.Pending {
		t.Fatalf("un activo vendido non pode conservar o valor do último resultado: %#v", r)
	}
	if r := rows[novo.ID]; !closeTo(r.TotalInvested, 300) || r.HasCurrentValue || r.HasGain || !r.Pending {
		t.Fatalf("un activo sen resultados debe mostrar o investido e quedar pendente: %#v", r)
	}
	if r := rows[restos.ID]; r.Pending || r.CurrentValue != 0 {
		t.Fatalf("uns restos de 0,004 USD non poden deixar o activo pendente: %#v", r)
	}
	if got.PendingCount != 2 {
		t.Fatalf("pendentes=%d, esperabamos 2 (aporte do mes e activo novo)", got.PendingCount)
	}

	total, err := viewtotalhistory.Build(st)
	if err != nil {
		t.Fatalf("viewtotalhistory.Build: %v", err)
	}
	// Aporte: 1500 − 150 + 300 + 0. Valor neto: 1200+500, 0, 300 ao custo, 0.
	if !closeTo(total.LifetimeAporte, 1650) || !closeTo(total.CurrentValue, 2000) || !closeTo(total.TotalGain, 350) || !total.HasTotalGain {
		t.Fatalf("KPIs totais inesperados: aporte=%.4f valor=%.4f G/P=%.4f (%v)", total.LifetimeAporte, total.CurrentValue, total.TotalGain, total.HasTotalGain)
	}
}
