package deletetransaction_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/deletetransaction"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	assets     []domain.Asset
	txByAsset  map[int64][]domain.Transaction
	deletedIDs []int64
	listAEr    error
	listTEr    error
	delEr      error
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

func (f *fakeRepo) DeleteTransaction(id int64) error {
	if f.delEr != nil {
		return f.delEr
	}
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func runWith(assets []domain.Asset, txs map[int64][]domain.Transaction, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, txByAsset: txs}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := deletetransaction.Run(r, &buf, repo)
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
			{ID: 101, AssetID: 10, AmountUSD: 200, Month: 5, Year: 2026},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Eliminar transacción") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsTransactionList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Transaccións de Acción — AAPL:",
		"[1] 04/2026 — 500.00 USD",
		"[2] 05/2026 — 200.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsSelection(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona a transacción a eliminar (1-2):") {
		t.Errorf("saída non contén o prompt esperado:\n%s", out)
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"✓ Transacción", "eliminada", "04/2026", "500.00 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsErrorOnInvalidTxSelection(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleTxs(), "1\n99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 100 {
		t.Errorf("deletedIDs = %v, esperabamos [100]", repo.deletedIDs)
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_NoTransactions_PrintsHint(t *testing.T) {
	out, repo, err := runWith(sampleAssets, map[int64][]domain.Transaction{}, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Este activo non ten transaccións extras rexistradas") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_HappyPath_DeletesCorrectTransaction(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 101 {
		t.Errorf("deletedIDs = %v, esperabamos [101]", repo.deletedIDs)
	}
}

func TestRun_RecoversFromInvalidSelection(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTxs(), "99\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 100 {
		t.Errorf("deletedIDs = %v, esperabamos [100]", repo.deletedIDs)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}
