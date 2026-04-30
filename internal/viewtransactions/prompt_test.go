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
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

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

func TestRun_PrintsTotals(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Compras: 500 + 300 = 800. Vendas: 200. Neto: +600.
	for _, want := range []string{
		"3 transacción(s)",
		"Compras: 800.00 USD",
		"Vendas: 200.00 USD",
		"+600.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTableHeaders(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"ID", "Ano", "Mes", "Tipo", "Cantidade"} {
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
	// A tx de venda ten amount=-200 pero debe aparecer como 200.00 USD na liña.
	// Asegúrome que aparece "200.00" (sen signo) — o "-200" non debe aparecer.
	if strings.Contains(out, "-200.00 USD") {
		t.Errorf("non debería mostrar amount negativo na fila de venda:\n%s", out)
	}
	if !strings.Contains(out, "200.00 USD") {
		t.Errorf("saída non mostra absoluto da venda:\n%s", out)
	}
}

func TestRun_PrintsAllTransactions(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// As 3 transaccións teñen IDs 100, 101, 102 e cantidades 500, 200 e 300.
	for _, want := range []string{"100", "101", "102", "500.00", "300.00"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_AssetWithoutTxs_PrintsHint(t *testing.T) {
	// Selecciona Vanguard (id=11) que non ten txs no map
	out, _, err := runWith(sampleAssets, sampleTxs(), "2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Vanguard non ten transaccións rexistradas") {
		t.Errorf("saída non contén suxestión:\n%s", out)
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
