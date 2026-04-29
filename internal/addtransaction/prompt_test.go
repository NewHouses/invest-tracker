package addtransaction_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addtransaction"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	invs   []domain.Investment
	saved  []domain.Transaction
	listEr error
	saveEr error
}

func (f *fakeRepo) ListInvestments() ([]domain.Investment, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.invs, nil
}

func (f *fakeRepo) InsertTransaction(t domain.Transaction) (int64, error) {
	if f.saveEr != nil {
		return 0, f.saveEr
	}
	f.saved = append(f.saved, t)
	return int64(len(f.saved)), nil
}

func runWith(invs []domain.Investment, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{invs: invs}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addtransaction.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleInvs = []domain.Investment{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 1, Year: 2026},
	{ID: 11, Type: domain.Indice, Name: "Vanguard S&P 500", AmountUSD: 2000, Month: 1, Year: 2026},
}

const validInput = "1\n500.00\n5\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleInvs, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir nova transacción") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsInvestmentList(t *testing.T) {
	out, _, err := runWith(sampleInvs, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"Investimentos:",
		"[1] Acción — AAPL",
		"[2] Índice — Vanguard S&P 500",
	}
	for _, e := range expects {
		if !strings.Contains(out, e) {
			t.Errorf("saída non contén %q:\n%s", e, out)
		}
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	out, _, err := runWith(sampleInvs, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"Selecciona (1-2):",
		"Cantidade (USD):",
		"Mes (1-12):",
		"Ano:",
	}
	for _, e := range expects {
		if !strings.Contains(out, e) {
			t.Errorf("saída non contén %q:\n%s", e, out)
		}
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, repo, err := runWith(sampleInvs, validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "✓ Transacción gardada") {
		t.Errorf("saída non contén a confirmación:\n%s", out)
	}
	if !strings.Contains(out, "AAPL") {
		t.Errorf("saída non contén o nome:\n%s", out)
	}
	if !strings.Contains(out, "#1") {
		t.Errorf("saída non contén o ID:\n%s", out)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
}

func TestRun_HappyPath_SavesCorrectTransaction(t *testing.T) {
	_, repo, err := runWith(sampleInvs, "2\n750.25\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("repo.saved tamaño = %d, esperabamos 1", len(repo.saved))
	}
	got := repo.saved[0]
	want := domain.Transaction{
		InvestmentID: 11,
		AmountUSD:    750.25,
		Month:        5,
		Year:         2026,
	}
	if got != want {
		t.Errorf("guardado = %+v, queremos %+v", got, want)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	_, repo, err := runWith(sampleInvs, "1\n750,25\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].AmountUSD != 750.25 {
		t.Errorf("AmountUSD = %v, esperabamos 750.25", repo.saved)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, repo, err := runWith(sampleInvs, "99\n1\n500\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
	if len(repo.saved) != 1 || repo.saved[0].InvestmentID != 10 {
		t.Errorf("expected save sobre id=10, got %+v", repo.saved)
	}
}

func TestRun_RecoversFromInvalidSelection(t *testing.T) {
	_, repo, err := runWith(sampleInvs, "0\n3\nabc\n2\n100\n5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].InvestmentID != 11 {
		t.Errorf("expected save sobre id=11, got %+v", repo.saved)
	}
}

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai investimentos") {
		t.Errorf("saída non contén a suxestión:\n%s", out)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleInvs, "1\n500\n5\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("repo.saved debería estar baleiro, got %v", repo.saved)
	}
}
