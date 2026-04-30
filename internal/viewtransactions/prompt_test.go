package viewtransactions_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtransactions"
)

type fakeRepo struct {
	assets    []domain.Asset
	txByAsset map[int64][]domain.Transaction
	listAEr   error
	listTEr   error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listAEr != nil {
		return nil, f.listAEr
	}
	return f.assets, nil
}

func (f *fakeRepo) ListTransactionsByAsset(id int64) ([]domain.Transaction, error) {
	if f.listTEr != nil {
		return nil, f.listTEr
	}
	return f.txByAsset[id], nil
}

func runWith(assets []domain.Asset, txs map[int64][]domain.Transaction, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, txByAsset: txs}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := viewtransactions.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026},
	{ID: 11, Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 1, Year: 2026},
}

// AAPL: 3 txs en 04, 05 (venda), 06 do 2026.
// Coa compra inicial en 03/2026 → 4 entradas.
func sampleTxs() map[int64][]domain.Transaction {
	return map[int64][]domain.Transaction{
		10: {
			{ID: 100, AssetID: 10, AmountUSD: 500, Month: 4, Year: 2026},
			{ID: 101, AssetID: 10, AmountUSD: -200, Month: 5, Year: 2026},
			{ID: 102, AssetID: 10, AmountUSD: 300, Month: 6, Year: 2026},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Transaccións dun activo") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsAssetTitle(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Transaccións de Acción — AAPL") {
		t.Errorf("saída non contén título do activo:\n%s", out)
	}
}

func TestRun_IncludesInitialPurchaseInTotals(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Inicial 1000 + tx compras 500 + 300 = 1800. Vendas 200. Neto +1600.
	for _, want := range []string{
		"4 entradas (incluíndo a compra inicial)",
		"Compras: 1800.00 USD",
		"Vendas: 200.00 USD",
		"+1600.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsCompraVendaColumnHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Compra/Venda") {
		t.Errorf("saída non contén cabeceira da columna 'Compra/Venda':\n%s", out)
	}
}

func TestRun_PrintsTableHeaders(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"ID", "Ano", "Mes", "Compra/Venda", "Cantidade"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsCompraAndVendaLabels(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "COMPRA") {
		t.Errorf("saída non contén COMPRA:\n%s", out)
	}
	if !strings.Contains(out, "VENDA") {
		t.Errorf("saída non contén VENDA:\n%s", out)
	}
}

func TestRun_PrintsAbsoluteAmountForVenda(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(out, "-200.00 USD") {
		t.Errorf("non debería mostrar amount negativo na fila de venda:\n%s", out)
	}
	if !strings.Contains(out, "200.00 USD") {
		t.Errorf("saída non mostra absoluto da venda:\n%s", out)
	}
}

func TestRun_PrintsInitialPurchaseRow(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// AAPL creouse en 03/2026 con cantidade 1000.
	if !strings.Contains(out, "1000.00 USD") {
		t.Errorf("saída non mostra a compra inicial 1000.00 USD:\n%s", out)
	}
	// O ID da compra inicial é "—" porque non é unha fila da táboa transactions.
	if !strings.Contains(out, "—") {
		t.Errorf("saída non mostra '—' como id da compra inicial:\n%s", out)
	}
}

func TestRun_OrdersByYearMonth(t *testing.T) {
	// Tx en 6/2026 (id baixo) e tx en 4/2026 (id alto) → debe ordenarse por data, non id.
	txs := map[int64][]domain.Transaction{
		10: {
			{ID: 200, AssetID: 10, AmountUSD: 600, Month: 6, Year: 2026},
			{ID: 100, AssetID: 10, AmountUSD: 400, Month: 4, Year: 2026},
			{ID: 300, AssetID: 10, AmountUSD: 500, Month: 5, Year: 2026},
		},
	}
	out, _, err := runWith(sampleAssets, txs, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Restrinxir busca ao corpo da táboa para evitar falsas coincidencias na cabeceira.
	tableStart := strings.Index(out, "Cantidade")
	if tableStart < 0 {
		t.Fatalf("non se atopou cabeceira da táboa:\n%s", out)
	}
	table := out[tableStart:]
	pos1000 := strings.Index(table, "1000.00 USD")
	pos400 := strings.Index(table, "400.00 USD")
	pos500 := strings.Index(table, "500.00 USD")
	pos600 := strings.Index(table, "600.00 USD")
	if pos1000 < 0 || pos400 < 0 || pos500 < 0 || pos600 < 0 {
		t.Fatalf("non se atoparon todas as cantidades:\n%s", table)
	}
	if !(pos1000 < pos400 && pos400 < pos500 && pos500 < pos600) {
		t.Errorf("filas non en orde cronolóxica: 1000=%d 400=%d 500=%d 600=%d\n%s",
			pos1000, pos400, pos500, pos600, table)
	}
}

func TestRun_AssetWithoutTxs_ShowsOnlyInitial(t *testing.T) {
	// Sen txs explícitas: só aparece a compra inicial (1 fila).
	out, _, err := runWith(sampleAssets, map[int64][]domain.Transaction{}, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "1 entradas (incluíndo a compra inicial)") {
		t.Errorf("saída non mostra exactamente 1 entrada:\n%s", out)
	}
	if !strings.Contains(out, "1000.00 USD") {
		t.Errorf("saída non mostra a compra inicial 1000.00 USD:\n%s", out)
	}
}

func TestRun_NoAssets_PrintsHint(t *testing.T) {
	out, _, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_RecoversFromInvalidSelection(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if !strings.Contains(out, "Acción — AAPL") {
		t.Errorf("non chegou a mostrar tras recuperarse:\n%s", out)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, _, err := runWith(sampleAssets, sampleTxs(), "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
