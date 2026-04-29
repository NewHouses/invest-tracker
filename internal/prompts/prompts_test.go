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
