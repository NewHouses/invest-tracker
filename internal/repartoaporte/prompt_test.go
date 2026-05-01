package repartoaporte_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
)

type fakeRepo struct {
	assets []domain.Asset
	listEr error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func runWith(repo *fakeRepo, input string) (string, error) {
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := repartoaporte.Run(r, &buf, repo)
	return buf.String(), err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Accion, Name: "MSFT"},
	{ID: 12, Type: domain.Indice, Name: "Vanguard"},
	{ID: 13, Type: domain.Fondo, Name: "Fund1"},
}

func TestRun_PrintsHeader(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets}, "1000\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Repartir aporte mensual") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PromptsAmount(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets}, "1000\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade (USD):") {
		t.Errorf("saída non solicita a cantidade:\n%s", out)
	}
}

func TestRun_NoAssets_PrintsHint(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "1000\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión sen activos:\n%s", out)
	}
}

func TestRun_DistributesAcrossSelectedTypes(t *testing.T) {
	// Total 1000, tipos: Acción (1) + Índice (2). Por tipo: 500.
	// Acción: 1,2 (AAPL+MSFT) → 250 cada. Índice: 1 (Vanguard) → 500.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Reparto de aporte mensual: 1000.00 USD",
		"Acción", "500.00 USD",
		"AAPL", "250.00 USD",
		"MSFT",
		"Índice",
		"Vanguard",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsPerTypeAmountInPrompt(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Cada tipo debe mostrar a cantidade asignada antes da subselección.
	if !strings.Contains(out, "→ 500.00 USD a Acción") {
		t.Errorf("saída non sinala 500.00 USD asignados a Acción:\n%s", out)
	}
	if !strings.Contains(out, "→ 500.00 USD a Índice") {
		t.Errorf("saída non sinala 500.00 USD asignados a Índice:\n%s", out)
	}
}

func TestRun_SingleTypeSingleAsset(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{
		{ID: 10, Type: domain.Accion, Name: "AAPL"},
	}}
	out, err := runWith(repo, "300\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Reparto de aporte mensual: 300.00 USD",
		"Acción", "300.00 USD",
		"AAPL",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_UnevenSplit(t *testing.T) {
	// Total 1000, tres tipos seleccionados: 333.33 cada.
	// Acción: 1 asset → 333.33. Índice: 1 asset → 333.33. Fondo: 1 asset → 333.33.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2,3\n1\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "333.33 USD") {
		t.Errorf("saída non mostra 333.33 USD:\n%s", out)
	}
}

func TestRun_RejectsInvalidTypeSelection(t *testing.T) {
	// "abc" non é válido → re-pregunta. Despois "1" → Acción.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\nabc\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non amosa erro de selección de tipo:\n%s", out)
	}
}

func TestRun_RejectsOutOfRangeIndex(t *testing.T) {
	// "99" → fóra de rango. "1" recupera.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n99\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non rexeita índice fóra de rango:\n%s", out)
	}
}

func TestRun_RejectsInvalidAssetSelection(t *testing.T) {
	// Tipos: 1 (Acción). Activos: "abc" inválido, despois "1" recupera.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1\nabc\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non amosa erro de selección de activos:\n%s", out)
	}
}

func TestRun_DedupsRepeatedIndices(t *testing.T) {
	// "1,1,2" → trata como {1,2}. Verifícase indirectamente:
	// Acción seleccionado unha soa vez, total Acción = 1000 (un só tipo seleccionado).
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,1\n1,1,2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Acción único tipo → 1000 USD; AAPL + MSFT (deduplicados) → 500 cada
	if !strings.Contains(out, "1000.00 USD") {
		t.Errorf("saída non mostra 1000.00 ao tipo único:\n%s", out)
	}
	if !strings.Contains(out, "500.00 USD") {
		t.Errorf("saída non mostra 500.00 por activo (2 tras dedupe):\n%s", out)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	// Sen entrada → Amount lerá EOF.
	_, err := runWith(&fakeRepo{assets: sampleAssets}, "")
	if err == nil {
		t.Fatal("esperabamos erro por entrada baleira")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
}

func TestRun_RejectsInvalidAmount(t *testing.T) {
	// "abc" non é cantidade válida → re-pregunta. Despois "100".
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"abc\n100\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade non válida") {
		t.Errorf("saída non rexeita cantidade non válida:\n%s", out)
	}
}
