package addtxmonth_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addtxmonth"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	assets   []domain.Asset
	saved    []domain.Transaction
	listEr   error
	insertEr error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) InsertTransaction(tx domain.Transaction) (int64, error) {
	if f.insertEr != nil {
		return 0, f.insertEr
	}
	f.saved = append(f.saved, tx)
	return int64(len(f.saved)), nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addtxmonth.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Accion, Name: "MSFT"},
	{ID: 12, Type: domain.Indice, Name: "Vanguard"},
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"4\n2026\n100\n200\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir transaccións do mes") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PromptsMonthAndYear(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"4\n2026\n100\n200\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_NoAssets_PrintsHint(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
}

func TestRun_PromptsEachAssetByName(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"4\n2026\n100\n200\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Acción — AAPL:",
		"Acción — MSFT:",
		"Índice — Vanguard:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén o prompt %q:\n%s", want, out)
		}
	}
}

func TestRun_AddsTransactionForEachAsset(t *testing.T) {
	repo := &fakeRepo{assets: sampleAssets}
	_, err := runWith(repo, "4\n2026\n100\n200\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("saved=%d, esperabamos 3", len(repo.saved))
	}
	want := []struct {
		assetID int64
		amount  float64
	}{
		{10, 100},
		{11, 200},
		{12, 300},
	}
	for i, w := range want {
		got := repo.saved[i]
		if got.AssetID != w.assetID || got.AmountUSD != w.amount ||
			got.Month != 4 || got.Year != 2026 {
			t.Errorf("saved[%d] = %+v, esperabamos asset=%d amount=%.2f", i, got, w.assetID, w.amount)
		}
	}
}

func TestRun_SkipsOnEmptyInput(t *testing.T) {
	// AAPL: 100, MSFT: skip, Vanguard: 300 → só 2 saved.
	repo := &fakeRepo{assets: sampleAssets}
	out, err := runWith(repo, "4\n2026\n100\n\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 2 {
		t.Errorf("saved=%d, esperabamos 2 (saltouse MSFT)", len(repo.saved))
	}
	for _, tx := range repo.saved {
		if tx.AssetID == 11 {
			t.Errorf("MSFT non debería estar gardado: %+v", tx)
		}
	}
	if !strings.Contains(out, "1 saltado") {
		t.Errorf("saída non sinala 1 saltado:\n%s", out)
	}
}

func TestRun_RejectsInvalidAmount(t *testing.T) {
	// AAPL: "abc" inválido, despois "100". Resto baleiros.
	repo := &fakeRepo{assets: sampleAssets}
	out, err := runWith(repo, "4\n2026\nabc\n100\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Valor non válido") {
		t.Errorf("saída non rexeita valor non válido:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].AmountUSD != 100 {
		t.Errorf("saved = %+v, esperabamos só AAPL=100", repo.saved)
	}
}

func TestRun_RejectsZeroAmount(t *testing.T) {
	// AAPL: "0" non válido (debe ser > 0), despois "50". Outros saltados.
	repo := &fakeRepo{assets: sampleAssets}
	out, err := runWith(repo, "4\n2026\n0\n50\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Valor non válido") {
		t.Errorf("saída non rexeita 0:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].AmountUSD != 50 {
		t.Errorf("saved = %+v, esperabamos só AAPL=50", repo.saved)
	}
}

func TestRun_PrintsSummary(t *testing.T) {
	repo := &fakeRepo{assets: sampleAssets}
	out, err := runWith(repo, "4\n2026\n100\n200\n300\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "3 transacción(s) engadidas (600.00 USD total)") {
		t.Errorf("saída non contén o resumo correcto:\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, err := runWith(&fakeRepo{assets: sampleAssets}, "4\n2026\n100\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{
		{ID: 10, Type: domain.Accion, Name: "AAPL"},
	}}
	_, err := runWith(repo, "4\n2026\n100,50\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].AmountUSD != 100.50 {
		t.Errorf("saved = %+v, esperabamos amount=100.50", repo.saved)
	}
}
