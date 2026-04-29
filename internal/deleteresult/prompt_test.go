package deleteresult_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/deleteresult"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	assets         []domain.Asset
	resultsByAsset map[int64][]domain.MonthlyResult
	deletedIDs     []int64
	listAEr        error
	listREr        error
	delEr          error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listAEr != nil {
		return nil, f.listAEr
	}
	return f.assets, nil
}

func (f *fakeRepo) ListMonthlyResultsByAsset(id int64) ([]domain.MonthlyResult, error) {
	if f.listREr != nil {
		return nil, f.listREr
	}
	return f.resultsByAsset[id], nil
}

func (f *fakeRepo) DeleteMonthlyResult(id int64) error {
	if f.delEr != nil {
		return f.delEr
	}
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func runWith(assets []domain.Asset, results map[int64][]domain.MonthlyResult, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets, resultsByAsset: results}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := deleteresult.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

func sampleResults() map[int64][]domain.MonthlyResult {
	return map[int64][]domain.MonthlyResult{
		10: {
			{ID: 200, AssetID: 10, ResultUSD: 1100, Month: 4, Year: 2026},
			{ID: 201, AssetID: 10, ResultUSD: 1200, Month: 5, Year: 2026},
		},
	}
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleResults(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Eliminar resultado mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleResults(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsResultList(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleResults(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Resultados de Acción — AAPL:",
		"[1] 04/2026 — 1100.00 USD",
		"[2] 05/2026 — 1200.00 USD",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsSelection(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleResults(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona o resultado a eliminar (1-2):") {
		t.Errorf("saída non contén o prompt esperado:\n%s", out)
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, _, err := runWith(sampleAssets, sampleResults(), "1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"✓ Resultado", "eliminado", "04/2026", "1100.00 USD"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HappyPath_DeletesCorrectResult(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleResults(), "1\n2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 201 {
		t.Errorf("deletedIDs = %v, esperabamos [201]", repo.deletedIDs)
	}
}

func TestRun_PrintsErrorOnInvalidResultSelection(t *testing.T) {
	out, repo, err := runWith(sampleAssets, sampleResults(), "1\n99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 200 {
		t.Errorf("deletedIDs = %v, esperabamos [200]", repo.deletedIDs)
	}
}

func TestRun_RecoversFromInvalidAssetSelection(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleResults(), "99\n1\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 200 {
		t.Errorf("deletedIDs = %v, esperabamos [200]", repo.deletedIDs)
	}
}

func TestRun_EmptyAssets_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, nil, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Aínda non hai activos") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_NoResults_PrintsHint(t *testing.T) {
	out, repo, err := runWith(sampleAssets, map[int64][]domain.MonthlyResult{}, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Este activo non ten resultados mensuais rexistrados") {
		t.Errorf("saída non contén suxestión:\n%s", out)
	}
	if len(repo.deletedIDs) != 0 {
		t.Errorf("deletedIDs debería estar baleiro, got %v", repo.deletedIDs)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, sampleResults(), "1\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
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
		assets:         sampleAssets,
		resultsByAsset: sampleResults(),
		delEr:          errors.New("boom"),
	}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader("1\n1\n"))
	err := deleteresult.Run(r, &buf, repo)
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "eliminando resultado") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}
