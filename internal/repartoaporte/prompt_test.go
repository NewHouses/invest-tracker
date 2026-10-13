package repartoaporte_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
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
	out, err := runWith(&fakeRepo{assets: sampleAssets}, "1000\n1\n1\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Repartir aporte mensual") {
		t.Errorf("saída non contén cabeceira:\n%s", out)
	}
}

func TestRun_PromptsAmount(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets}, "1000\n1\n1\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade (USD):") {
		t.Errorf("saída non solicita a cantidade:\n%s", out)
	}
}

// Sen activos creados, o reparto faise entre activos novos.
func TestRun_NoAssets_SplitsAmongNewAssets(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "1000\n2\nNovo índice\ns\n4\nNovo fondo\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Aínda non hai activos creados",
		"Novo índice (novo)", "Novo fondo (novo)",
		"50.00 %", "500.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_DistributesAcrossSelectedTypes(t *testing.T) {
	// Total 1000, tipos: Acción (1) + Índice (2). Por tipo: 500.
	// Acción: 1,2 (AAPL+MSFT) → 250 cada. Índice: 1 (Vanguard) → 500.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n\n\n")
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

func TestRun_ShowsProposedSplit(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Proposta a partes iguais: 50 % por tipo, 25 % por activo de Acción.
	for _, want := range []string{"Reparto proposto", "Acción — AAPL", " 25.00 %", "Índice — Vanguard", " 50.00 %", "Queres cambiar as porcentaxes?"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_CustomPercents(t *testing.T) {
	// AAPL 70 %, MSFT 20 %, Vanguard 10 %.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n\ns\n70\n20,5\n9.5 %\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"700.00 USD", "205.00 USD", "95.00 USD", "90.50 %", "905.00 USD", "9.50 %"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PercentsMustAddUpTo100(t *testing.T) {
	// Primeiro suman 80; na segunda volta Intro mantén o valor anterior.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1,2\n1,2\n1\n\ns\n50\n20\n10\n\n40\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"⚠ As porcentaxes suman 80.00 %; teñen que sumar 100 %.", "MSFT [20.00]", "500.00 USD", "400.00 USD", "100.00 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_RejectsInvalidPercent(t *testing.T) {
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1\n1\n\ns\nabc\n150\n-1\n100\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.Count(out, "⚠ Porcentaxe non válida"); got != 3 {
		t.Errorf("avisos de porcentaxe non válida = %d, queremos 3:\n%s", got, out)
	}
	if !strings.Contains(out, "1000.00 USD") {
		t.Errorf("saída non contén o reparto:\n%s", out)
	}
}

func TestRun_AddsNewAssetToItsType(t *testing.T) {
	// Acción: AAPL e un activo novo (NVDA); Fondo: outro activo novo.
	out, err := runWith(&fakeRepo{assets: sampleAssets},
		"1000\n1\n1\ns\n1\nNVDA\ns\n4\nNovo fondo\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"Acción — NVDA (novo)", "Fondo — Novo fondo (novo)", "NVDA (novo)", "250.00 USD", "500.00 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Fund1") {
		t.Errorf("Fund1 non se seleccionou:\n%s", out)
	}
}

func TestRun_NewAssetNeedsName(t *testing.T) {
	out, err := runWith(&fakeRepo{}, "100\n1\n\nETF\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "⚠ O nome non pode estar baleiro") || !strings.Contains(out, "ETF (novo)") {
		t.Errorf("saída inesperada:\n%s", out)
	}
}

func TestRun_English(t *testing.T) {
	defer i18n.SetCurrent(i18n.Current())
	i18n.SetCurrent(i18n.English)
	out, err := runWith(&fakeRepo{assets: sampleAssets}, "1000\n1\n1\ny\n1\nNVDA\nn\ny\n60\n40\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Do you want to add assets you have not created yet? [y/N]:",
		"Proposed split:", "Stock — NVDA (new)",
		"Do you want to change the percentages? [y/N]:",
		"Monthly contribution split: 1000.00 USD", "600.00 USD", "400.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestRun_SingleTypeSingleAsset(t *testing.T) {
	repo := &fakeRepo{assets: []domain.Asset{
		{ID: 10, Type: domain.Accion, Name: "AAPL"},
	}}
	out, err := runWith(repo, "300\n1\n1\n\n\n")
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
		"1000\n1,2,3\n1\n1\n1\n\n\n")
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
		"1000\nabc\n1\n1\n\n\n")
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
		"1000\n99\n1\n1\n\n\n")
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
		"1000\n1\nabc\n1\n\n\n")
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
		"1000\n1,1\n1,1,2\n\n\n")
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
		"abc\n100\n1\n1\n\n\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade non válida") {
		t.Errorf("saída non rexeita cantidade non válida:\n%s", out)
	}
}
