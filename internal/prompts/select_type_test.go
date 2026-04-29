package prompts_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

func TestSelectAssetType_PrintsMenu(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.SelectAssetType(bufio.NewReader(strings.NewReader("1\n")), &w)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	for _, want := range []string{
		"Tipo de investimento:",
		"[1] Acción",
		"[2] Índice",
		"[3] Copy-trading",
		"[4] Fondo",
	} {
		if !strings.Contains(w.String(), want) {
			t.Errorf("saída non contén %q:\n%s", want, w.String())
		}
	}
}

func TestSelectAssetType_HappyPaths(t *testing.T) {
	cases := []struct {
		input string
		want  domain.AssetType
	}{
		{"1\n", domain.Accion},
		{"2\n", domain.Indice},
		{"3\n", domain.CopyTrading},
		{"4\n", domain.Fondo},
	}
	for _, c := range cases {
		var w bytes.Buffer
		got, err := prompts.SelectAssetType(bufio.NewReader(strings.NewReader(c.input)), &w)
		if err != nil {
			t.Errorf("input=%q erro=%v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("input=%q got=%v, esperabamos %v", c.input, got, c.want)
		}
	}
}

func TestSelectAssetType_Reprompts(t *testing.T) {
	var w bytes.Buffer
	got, err := prompts.SelectAssetType(bufio.NewReader(strings.NewReader("9\nabc\n2\n")), &w)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if got != domain.Indice {
		t.Errorf("got %v, esperabamos Indice", got)
	}
	if !strings.Contains(w.String(), "Tipo non válido") {
		t.Errorf("saída non contén o erro:\n%s", w.String())
	}
}

func TestSelectAssetType_EOFEmpty(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.SelectAssetType(bufio.NewReader(strings.NewReader("")), &w)
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
