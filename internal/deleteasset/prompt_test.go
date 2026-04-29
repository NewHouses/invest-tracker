package deleteasset_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/deleteasset"
	"invest-tracker/internal/domain"
)

type fakeRepo struct {
	assets     []domain.Asset
	deletedIDs []int64
	listEr     error
	delEr      error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) DeleteAsset(id int64) error {
	if f.delEr != nil {
		return f.delEr
	}
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func runWith(assets []domain.Asset, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := deleteasset.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL"},
	{ID: 11, Type: domain.Indice, Name: "Vanguard"},
}

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Eliminar activo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PromptsSelection(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selecciona (1-2):") {
		t.Errorf("saída non contén o prompt:\n%s", out)
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"✓ Activo", "eliminado", "Acción — AAPL", "transaccións e resultados"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_HappyPath_DeletesCorrectAsset(t *testing.T) {
	_, repo, err := runWith(sampleAssets, "2\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 11 {
		t.Errorf("deletedIDs = %v, esperabamos [11]", repo.deletedIDs)
	}
}

func TestRun_PrintsErrorOnInvalidSelection(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "99\n1\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida") {
		t.Errorf("saída non contén o erro:\n%s", out)
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 10 {
		t.Errorf("deletedIDs = %v, esperabamos [10]", repo.deletedIDs)
	}
}

func TestRun_EmptyList_PrintsHint(t *testing.T) {
	out, repo, err := runWith(nil, "")
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

func TestRun_EOFEmpty_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, "")
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
		assets: sampleAssets,
		delEr:  errors.New("boom"),
	}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader("1\n"))
	err := deleteasset.Run(r, &buf, repo)
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "eliminando activo") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}
