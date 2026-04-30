package addtxbatch_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addtxbatch"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	assets []domain.Asset
	saved  []domain.Transaction
	listEr error
	saveEr error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) InsertTransaction(t domain.Transaction) (int64, error) {
	if f.saveEr != nil {
		return 0, f.saveEr
	}
	f.saved = append(f.saved, t)
	return int64(len(f.saved)), nil
}

func runWith(assets []domain.Asset, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addtxbatch.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026},
}

// Input layout:
// asset(1), mode(1=perTx|2=fixed|3=auto)
// Mode 1: per tx → mes, ano, tipo, cantidade, continuar?
// Mode 2/3: mes inicial, ano inicial, despois por tx: tipo, cantidade, continuar?

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n2\n4\n2026\n1\n100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir varias transaccións") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PrintsModeMenu(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n2\n4\n2026\n1\n100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Modo de entrada:",
		"[1] Establecer mes/ano en cada transacción",
		"[2] Engadir todas no mesmo mes",
		"[3] Unha por mes (auto-incrementa) dende un mes inicial",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}

// Mode 1 (per-tx): each tx asks month/year explicitly
func TestRun_Mode1_PerTx_SavesWithDifferentDates(t *testing.T) {
	// asset=1, mode=1, tx1: mes=4, ano=2026, tipo=1(compra), cant=100, sí
	//                  tx2: mes=5, ano=2026, tipo=2(venda), cant=50, n
	_, repo, err := runWith(sampleAssets, "1\n1\n4\n2026\n1\n100\ns\n5\n2026\n2\n50\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved tamaño = %d, esperabamos 2", len(repo.saved))
	}
	if repo.saved[0] != (domain.Transaction{AssetID: 10, AmountUSD: 100, Month: 4, Year: 2026}) {
		t.Errorf("saved[0] = %+v", repo.saved[0])
	}
	if repo.saved[1] != (domain.Transaction{AssetID: 10, AmountUSD: -50, Month: 5, Year: 2026}) {
		t.Errorf("saved[1] = %+v (esperabamos -50 venda)", repo.saved[1])
	}
}

// Mode 2 (fixed month): one date prompt, all txs share it
func TestRun_Mode2_FixedMonth_SharesDate(t *testing.T) {
	// asset=1, mode=2, mes=4, ano=2026, tx1: tipo=1, cant=100, sí, tx2: tipo=1, cant=200, n
	_, repo, err := runWith(sampleAssets, "1\n2\n4\n2026\n1\n100\ns\n1\n200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved tamaño = %d, esperabamos 2", len(repo.saved))
	}
	for i, tx := range repo.saved {
		if tx.Month != 4 || tx.Year != 2026 {
			t.Errorf("saved[%d] data = %02d/%d, esperabamos 04/2026", i, tx.Month, tx.Year)
		}
	}
	if repo.saved[0].AmountUSD != 100 || repo.saved[1].AmountUSD != 200 {
		t.Errorf("amounts incorrectos: %+v", repo.saved)
	}
}

// Mode 3 (auto-increment): one start date, txs in successive months
// (rollover de mes 12 → 1 do ano seguinte)
func TestRun_Mode3_AutoIncrement_AdvancesMonth(t *testing.T) {
	// Asset creado en 11/2025 para permitir empezar nesa data.
	earlyAssets := []domain.Asset{
		{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 11, Year: 2025},
	}
	// asset=1, mode=3, mes=11, ano=2025
	// tx1 (11/2025): tipo=1, cant=100, sí
	// tx2 (12/2025): tipo=1, cant=200, sí
	// tx3 (01/2026): tipo=1, cant=300, n
	_, repo, err := runWith(earlyAssets, "1\n3\n11\n2025\n1\n100\ns\n1\n200\ns\n1\n300\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved tamaño = %d, esperabamos 3", len(repo.saved))
	}
	want := []struct{ y, m int }{{2025, 11}, {2025, 12}, {2026, 1}}
	for i, w := range want {
		if repo.saved[i].Year != w.y || repo.saved[i].Month != w.m {
			t.Errorf("saved[%d] data = %02d/%d, esperabamos %02d/%d",
				i, repo.saved[i].Month, repo.saved[i].Year, w.m, w.y)
		}
	}
}

func TestRun_ContinueDefaultIsYes(t *testing.T) {
	// Resposta baleira no "continuar?" debe seguir engadindo
	// asset=1, mode=2, mes=4, ano=2026, tx1: tipo=1, cant=100, "" (default sí), tx2: tipo=1, cant=200, n
	_, repo, err := runWith(sampleAssets, "1\n2\n4\n2026\n1\n100\n\n1\n200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved = %d, esperabamos 2 (baleiro = sí)", len(repo.saved))
	}
}

func TestRun_RejectsTxBeforeAssetCreation_Mode1(t *testing.T) {
	// asset creado en 1/2026; tentar tx en 12/2025 → re-prompt; logo válida
	out, repo, err := runWith(sampleAssets, "1\n1\n12\n2025\n4\n2026\n1\n100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non pode ser anterior á data do activo (01/2026)") {
		t.Errorf("saída non contén o erro de data:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].Month != 4 {
		t.Errorf("saved = %+v", repo.saved)
	}
}

func TestRun_PrintsConfirmationSummary(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n2\n4\n2026\n1\n100\ns\n1\n200\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadíronse 2 transacción(s)") {
		t.Errorf("saída non contén o resumo:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, _, err := runWith(sampleAssets, "1\n2\n4\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_RejectsInvalidMode(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n9\n2\n4\n2026\n1\n100\nn\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Modo non válido") {
		t.Errorf("saída non contén o erro de modo:\n%s", out)
	}
}
