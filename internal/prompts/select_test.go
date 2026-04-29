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

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

func TestSelectAsset_PrintsList(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.SelectAsset(bufio.NewReader(strings.NewReader("1\n")), &w, sampleAssets)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	for _, want := range []string{"Investimentos:", "[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(w.String(), want) {
			t.Errorf("saída non contén %q:\n%s", want, w.String())
		}
	}
}

func TestSelectAsset_HappyPath(t *testing.T) {
	var w bytes.Buffer
	got, err := prompts.SelectAsset(bufio.NewReader(strings.NewReader("2\n")), &w, sampleAssets)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if got != sampleAssets[1] {
		t.Errorf("got %+v, esperabamos %+v", got, sampleAssets[1])
	}
}

func TestSelectAsset_RejectsOutOfRange(t *testing.T) {
	var w bytes.Buffer
	got, err := prompts.SelectAsset(bufio.NewReader(strings.NewReader("99\n0\n2\n")), &w, sampleAssets)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if got != sampleAssets[1] {
		t.Errorf("got %+v, esperabamos %+v", got, sampleAssets[1])
	}
	if !strings.Contains(w.String(), "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", w.String())
	}
}

func TestSelectAsset_RejectsNonNumeric(t *testing.T) {
	var w bytes.Buffer
	got, err := prompts.SelectAsset(bufio.NewReader(strings.NewReader("abc\n1\n")), &w, sampleAssets)
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if got != sampleAssets[0] {
		t.Errorf("got %+v, esperabamos %+v", got, sampleAssets[0])
	}
}

func TestSelectAsset_EOFEmpty_ReturnsEOF(t *testing.T) {
	var w bytes.Buffer
	_, err := prompts.SelectAsset(bufio.NewReader(strings.NewReader("")), &w, sampleAssets)
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}
