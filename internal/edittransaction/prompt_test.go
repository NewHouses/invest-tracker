package edittransaction_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/edittransaction"
)

type fakeRepo struct {
	assets    []domain.Asset
	txByAsset map[int64][]domain.Transaction
	updated   []domain.Transaction
	listAEr   error
	listTEr   error
	updEr     error
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

func (f *fakeRepo) UpdateTransaction(tx domain.Transaction) error {
	if f.updEr != nil {
		return f.updEr
	}
	f.updated = append(f.updated, tx)
	return nil
}

func runWith(assets []domain.Asset, txs map[int64][]domain.Transaction, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, txByAsset: txs}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := edittransaction.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026},
	{ID: 11, Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 3, Year: 2026},
}

func sampleTxs() map[int64][]domain.Transaction {
	return map[int64][]domain.Transaction{
		10: {
			{ID: 100, AssetID: 10, AmountUSD: 500, Month: 4, Year: 2026},   // COMPRA
			{ID: 101, AssetID: 10, AmountUSD: -200, Month: 5, Year: 2026},  // VENDA
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Editar unha transacción") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
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
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Transaccións de Acción — AAPL:",
		"[1] 04/2026 — COMPRA 500.00 USD",
		"[2] 05/2026 — VENDA 200.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsFieldMenu(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Que campo queres editar?",
		"[1] Cantidade (USD)",
		"[2] Tipo (Compra/Venda)",
		"[3] Data (mes/ano)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EditsAmount_KeepsSignForCompra(t *testing.T) {
	// Tx[0] = COMPRA 500. Cambiar cantidade a 750 → debe seguir sendo COMPRA (positivo).
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated tamaño = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.ID != 100 || got.AmountUSD != 750 || got.Month != 4 || got.Year != 2026 {
		t.Errorf("updated = %+v", got)
	}
}

func TestRun_EditsAmount_KeepsSignForVenda(t *testing.T) {
	// Tx[1] = VENDA -200. Cambiar cantidade a 350 → debe seguir VENDA → -350.
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n2\n1\n350\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated tamaño = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.AmountUSD != -350 {
		t.Errorf("AmountUSD = %v, esperabamos -350 (VENDA preservada)", got.AmountUSD)
	}
}

func TestRun_EditsType_FlipsSign_CompraToVenda(t *testing.T) {
	// Tx[0] = COMPRA 500. Cambiar tipo a Venda (2) → -500.
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n2\n2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated tamaño = %d, esperabamos 1", len(repo.updated))
	}
	if repo.updated[0].AmountUSD != -500 {
		t.Errorf("AmountUSD = %v, esperabamos -500", repo.updated[0].AmountUSD)
	}
}

func TestRun_EditsType_FlipsSign_VendaToCompra(t *testing.T) {
	// Tx[1] = VENDA -200. Cambiar tipo a Compra (1) → +200.
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n2\n2\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated tamaño = %d, esperabamos 1", len(repo.updated))
	}
	if repo.updated[0].AmountUSD != 200 {
		t.Errorf("AmountUSD = %v, esperabamos 200", repo.updated[0].AmountUSD)
	}
}

func TestRun_EditsDate_HappyPath(t *testing.T) {
	// Tx[0] = 04/2026. Cambiar data a 06/2027.
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n3\n6\n2027\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated tamaño = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.Month != 6 || got.Year != 2027 || got.AmountUSD != 500 {
		t.Errorf("updated = %+v", got)
	}
}

func TestRun_EditsDate_RejectsBeforeAssetCreation(t *testing.T) {
	// Asset creado en 03/2026. Tentar cambiar tx data a 12/2025 → rexeita.
	out, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n3\n12\n2025\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non pode ser anterior á data do activo (03/2026)") {
		t.Errorf("saída non contén o erro de data:\n%s", out)
	}
	// Tras o re-prompt válido (4/2026), debe gardar.
	if len(repo.updated) != 1 || repo.updated[0].Month != 4 || repo.updated[0].Year != 2026 {
		t.Errorf("updated = %+v", repo.updated)
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"✓ Transacción #100 actualizada", "COMPRA 750.00 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_NoAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.updated) != 0 {
		t.Errorf("updated debería estar baleiro, got %v", repo.updated)
	}
}

func TestRun_NoTransactions_PrintsHint(t *testing.T) {
	out, repo, err := runWith(sampleAssets, map[int64][]domain.Transaction{}, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non ten transaccións extras rexistradas") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.updated) != 0 {
		t.Errorf("updated debería estar baleiro, got %v", repo.updated)
	}
}

func TestRun_RejectsInvalidTxSelection(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleTxs(), "1\n99\n1\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.updated) != 1 || repo.updated[0].ID != 100 {
		t.Errorf("updated = %+v, esperabamos id=100", repo.updated)
	}
}

func TestRun_RejectsInvalidFieldChoice(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n9\n1\n750\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida, escolle 1-3") {
		t.Errorf("saída non contén o erro de campo:\n%s", out)
	}
	if len(repo.updated) != 1 {
		t.Errorf("expected 1 save tras recuperación, got %d", len(repo.updated))
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleTxs(), "1\n1\n1\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.updated) != 0 {
		t.Errorf("updated debería estar baleiro, got %v", repo.updated)
	}
}

func TestRun_PropagatesRepoError(t *testing.T) {
	repo := &fakeRepo{
		assets:    sampleAssets,
		txByAsset: sampleTxs(),
		updEr:     errors.New("boom"),
	}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader("1\n1\n1\n750\n"))
	err := edittransaction.Run(r, &buf, repo)
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "actualizando transacción") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}
