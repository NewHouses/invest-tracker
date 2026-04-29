package deletedividend_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/deletedividend"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	dividends  []domain.Dividend
	deletedIDs []int64
	listEr     error
	delEr      error
}

func (f *fakeRepo) ListDividends() ([]domain.Dividend, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.dividends, nil
}

func (f *fakeRepo) DeleteDividend(id int64) error {
	if f.delEr != nil {
		return f.delEr
	}
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func runWith(divs []domain.Dividend, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{dividends: divs}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := deletedividend.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleDividends = []domain.Dividend{
	{ID: 100, AmountUSD: 50.25, Month: 4, Year: 2026},
	{ID: 101, AmountUSD: 75.50, Month: 5, Year: 2026},
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleDividends, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Eliminar dividendo mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsDividendList(t *testing.T) {
	out, _, err := runWith(sampleDividends, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Dividendos:",
		"[1] 04/2026 — 50.25 USD",
		"[2] 05/2026 — 75.50 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsSelection(t *testing.T) {
	out, _, err := runWith(sampleDividends, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona o dividendo a eliminar (1-2):") {
		t.Errorf("saída non contén o prompt esperado:\n%s", out)
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, _, err := runWith(sampleDividends, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"✓ Dividendo", "eliminado", "04/2026", "50.25 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HappyPath_DeletesCorrectDividend(t *testing.T) {
	_, repo, err := runWith(sampleDividends, "2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 101 {
		t.Errorf("deletedIDs = %v, esperabamos [101]", repo.deletedIDs)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, repo, err := runWith(sampleDividends, "99\n1\n")
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

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai dividendos rexistrados") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleDividends, "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_PropagatesRepoError(t *testing.T) {
	repo := &fakeRepo{
		dividends: sampleDividends,
		delEr:     errors.New("boom"),
	}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader("1\n"))
	err := deletedividend.Run(r, &buf, repo)
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "eliminando dividendo") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}
