package clearmonth_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/clearmonth"
)

type fakeRepo struct {
	calls    []callRecord
	count    int64
	delEr    error
}

type callRecord struct {
	year  int
	month int
}

func (f *fakeRepo) DeleteMonthlyResultsByMonth(year, month int) (int64, error) {
	f.calls = append(f.calls, callRecord{year: year, month: month})
	if f.delEr != nil {
		return 0, f.delEr
	}
	return f.count, nil
}

func runWith(count int64, delEr error, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{count: count, delEr: delEr}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := clearmonth.Run(r, &buf, repo)
	return buf.String(), repo, err
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(0, nil, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Limpar mes") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsMonthAndYear(t *testing.T) {
	out, _, err := runWith(0, nil, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Mes (1-12):", "Ano:"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_CallsRepoWithCorrectArgs(t *testing.T) {
	_, repo, err := runWith(3, nil, "5\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("repo.calls len = %d, esperabamos 1", len(repo.calls))
	}
	if repo.calls[0] != (callRecord{year: 2026, month: 5}) {
		t.Errorf("repo.calls[0] = %+v, esperabamos {2026, 5}", repo.calls[0])
	}
}

func TestRun_PrintsConfirmationWithCount(t *testing.T) {
	out, _, err := runWith(3, nil, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "✓ Limpado mes 04/2026: 3 resultado(s) eliminado(s).") {
		t.Errorf("saída non contén confirmación esperada:\n%s", out)
	}
}

func TestRun_PrintsHintWhenZero(t *testing.T) {
	out, _, err := runWith(0, nil, "4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Non había resultados rexistrados en 04/2026.") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if strings.Contains(out, "✓ Limpado") {
		t.Errorf("non debería mostrar confirmación con cero filas:\n%s", out)
	}
}

func TestRun_PropagatesRepoError(t *testing.T) {
	_, _, err := runWith(0, errors.New("boom"), "4\n2026\n")
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "limpando mes") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, repo, err := runWith(0, nil, "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.calls) != 0 {
		t.Errorf("repo non debería terse chamado: %v", repo.calls)
	}
}
