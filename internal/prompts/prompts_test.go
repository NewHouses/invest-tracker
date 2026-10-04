package prompts_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/prompts"
)

func newReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestReadLine_TrimsWhitespace(t *testing.T) {
	got, err := prompts.ReadLine(newReader("  hola  \n"))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != "hola" {
		t.Errorf("got %q, esperabamos %q", got, "hola")
	}
}

func TestReadLine_EOFEmpty_ReturnsEOF(t *testing.T) {
	_, err := prompts.ReadLine(newReader(""))
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestReadLine_CancelSentinels(t *testing.T) {
	cases := []string{
		":q\n",
		"cancelar\n",
		"  :q  \n",   // espazos
		"CANCELAR\n", // maiúsculas
		":Q\n",
	}
	for _, in := range cases {
		_, err := prompts.ReadLine(newReader(in))
		if !errors.Is(err, prompts.ErrCancelled) {
			t.Errorf("input %q: esperabamos ErrCancelled, got %v", in, err)
		}
	}
}

func TestAmount_PropagatesCancellation(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.Amount(newReader(":q\n"), &w)
	if !errors.Is(err, prompts.ErrCancelled) {
		t.Errorf("esperabamos ErrCancelled, got %v", err)
	}
}

func TestMonth_PropagatesCancellation(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.Month(newReader(":q\n"), &w)
	if !errors.Is(err, prompts.ErrCancelled) {
		t.Errorf("esperabamos ErrCancelled, got %v", err)
	}
}

func TestYear_PropagatesCancellation(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.Year(newReader(":q\n"), &w)
	if !errors.Is(err, prompts.ErrCancelled) {
		t.Errorf("esperabamos ErrCancelled, got %v", err)
	}
}

func TestAmount_AcceptsDecimal(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Amount(newReader("123.45\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 123.45 {
		t.Errorf("got %v, esperabamos 123.45", v)
	}
}

func TestAmount_AcceptsComma(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Amount(newReader("123,45\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 123.45 {
		t.Errorf("got %v, esperabamos 123.45", v)
	}
}

func TestAmount_RejectsZeroOrNegative(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Amount(newReader("0\n-5\n10\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 10 {
		t.Errorf("got %v, esperabamos 10", v)
	}
	if !strings.Contains(w.String(), "Cantidade non válida") {
		t.Errorf("esperabamos mensaxe de erro, got: %s", w.String())
	}
}

func TestAmount_RejectsNonNumeric(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Amount(newReader("abc\n50\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 50 {
		t.Errorf("got %v, esperabamos 50", v)
	}
}

func TestMonth_AcceptsValidRange(t *testing.T) {
	var w bytes.Buffer
	for _, m := range []string{"1", "6", "12"} {
		v, err := prompts.Month(newReader(m+"\n"), &w)
		if err != nil {
			t.Fatalf("erro con %q: %v", m, err)
		}
		expect, _ := parseInt(m)
		if v != expect {
			t.Errorf("got %d, esperabamos %d", v, expect)
		}
	}
}

func TestMonth_RejectsOutOfRange(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Month(newReader("0\n13\n5\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 5 {
		t.Errorf("got %d, esperabamos 5", v)
	}
	if !strings.Contains(w.String(), "Mes non válido") {
		t.Errorf("esperabamos mensaxe de erro, got: %s", w.String())
	}
}

func TestYear_AcceptsValidRange(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Year(newReader("2026\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 2026 {
		t.Errorf("got %d, esperabamos 2026", v)
	}
}

func TestYear_RejectsOutOfRange(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Year(newReader("1800\n2200\n2026\n"), &w)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 2026 {
		t.Errorf("got %d, esperabamos 2026", v)
	}
	if !strings.Contains(w.String(), "Ano non válido") {
		t.Errorf("esperabamos mensaxe de erro, got: %s", w.String())
	}
}

func TestNonNegativeAmount_UsesLabelAndAcceptsZero(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.NonNegativeAmount(newReader("0\n"), &w, "Salario anual inicial (USD)")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 0 {
		t.Errorf("got %v, esperabamos 0", v)
	}
	if !strings.Contains(w.String(), "Salario anual inicial (USD): ") {
		t.Errorf("esperabamos a etiqueta no prompt, got: %s", w.String())
	}
}

func TestNonNegativeAmount_AcceptsComma(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.NonNegativeAmount(newReader("40000,50\n"), &w, "Salario")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 40000.50 {
		t.Errorf("got %v, esperabamos 40000.50", v)
	}
}

func TestNonNegativeAmount_RejectsNegativeEmptyAndNonNumeric(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.NonNegativeAmount(newReader("-1\n\nabc\n40000\n"), &w, "Salario")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 40000 {
		t.Errorf("got %v, esperabamos 40000", v)
	}
	if got := strings.Count(w.String(), "Cantidade non válida"); got != 3 {
		t.Errorf("esperabamos 3 avisos, got %d: %s", got, w.String())
	}
}

func TestNonNegativeAmount_PropagatesCancellation(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.NonNegativeAmount(newReader(":q\n"), &w, "Salario")
	if !errors.Is(err, prompts.ErrCancelled) {
		t.Errorf("esperabamos ErrCancelled, got %v", err)
	}
}

func TestPercent_ReturnsFraction(t *testing.T) {
	cases := []struct {
		input string
		want  float64
	}{
		{"1\n", 0.01},
		{"1,5\n", 0.015},
		{"0\n", 0},
		{"-5\n", -0.05},
	}
	for _, c := range cases {
		var w bytes.Buffer
		v, err := prompts.Percent(newReader(c.input), &w, "Retorno mensual esperado (%)", -100)
		if err != nil {
			t.Fatalf("input %q: erro inesperado: %v", c.input, err)
		}
		if v != c.want {
			t.Errorf("input %q: got %v, esperabamos %v", c.input, v, c.want)
		}
	}
}

func TestPercent_RejectsAtOrBelowMin(t *testing.T) {
	var w bytes.Buffer
	v, err := prompts.Percent(newReader("-100\n-150\n\nabc\n2\n"), &w, "Retorno", -100)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if v != 0.02 {
		t.Errorf("got %v, esperabamos 0.02", v)
	}
	if got := strings.Count(w.String(), "Porcentaxe non válida"); got != 4 {
		t.Errorf("esperabamos 4 avisos, got %d: %s", got, w.String())
	}
	if !strings.Contains(w.String(), "maior ca -100%") {
		t.Errorf("esperabamos o límite na mensaxe, got: %s", w.String())
	}
}

func TestPercent_PropagatesCancellation(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.Percent(newReader("cancelar\n"), &w, "Retorno", -100)
	if !errors.Is(err, prompts.ErrCancelled) {
		t.Errorf("esperabamos ErrCancelled, got %v", err)
	}
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("non numeric")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
