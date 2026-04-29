package editasset_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/editasset"
)

type fakeRepo struct {
	assets  []domain.Asset
	updated []domain.Asset
	listEr  error
	updEr   error
}

func (f *fakeRepo) ListAssets() ([]domain.Asset, error) {
	if f.listEr != nil {
		return nil, f.listEr
	}
	return f.assets, nil
}

func (f *fakeRepo) UpdateAsset(a domain.Asset) error {
	if f.updEr != nil {
		return f.updEr
	}
	f.updated = append(f.updated, a)
	return nil
}

func runWith(assets []domain.Asset, input string) (string, *fakeRepo, error) {
	repo := &fakeRepo{assets: assets}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := editasset.Run(r, &buf, repo)
	return buf.String(), repo, err
}

var sampleAssets = []domain.Asset{
	{ID: 10, Type: domain.Accion, Name: "AAPL", AmountUSD: 1000, Month: 3, Year: 2026},
	{ID: 11, Type: domain.Indice, Name: "Vanguard", AmountUSD: 2000, Month: 4, Year: 2026},
}

// Inputs: asset selection, field choice, then per-field new value(s)
// Field 1 = Nome, 2 = Data (mes+ano), 3 = Cantidade

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n3\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Editar activo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PrintsAssetList(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n3\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{"[1] Acción — AAPL", "[2] Índice — Vanguard"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_PrintsSelectedAssetDetails(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n3\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Activo seleccionado: Acción — AAPL (1000.00 USD, 03/2026)") {
		t.Errorf("saída non contén os detalles do activo:\n%s", out)
	}
}

func TestRun_PrintsFieldMenu(t *testing.T) {
	out, _, err := runWith(sampleAssets, "1\n3\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, want := range []string{
		"Que campo queres editar?",
		"[1] Nome",
		"[2] Data (mes/ano)",
		"[3] Cantidade (USD)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("saída non contén %q:\n%s", want, out)
		}
	}
}

func TestRun_EditsName_HappyPath(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "1\n1\nAAPL Renamed\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("repo.updated len = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.ID != 10 || got.Name != "AAPL Renamed" || got.AmountUSD != 1000 || got.Month != 3 || got.Year != 2026 {
		t.Errorf("updated = %+v, esperabamos só Name cambiado", got)
	}
	if !strings.Contains(out, "✓ Activo #10 actualizado") {
		t.Errorf("saída non contén confirmación:\n%s", out)
	}
	if !strings.Contains(out, "AAPL Renamed") {
		t.Errorf("saída non mostra novo nome:\n%s", out)
	}
}

func TestRun_EditsDate_HappyPath(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "1\n2\n6\n2027\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("repo.updated len = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.Month != 6 || got.Year != 2027 || got.Name != "AAPL" || got.AmountUSD != 1000 {
		t.Errorf("updated = %+v, esperabamos só Date cambiado", got)
	}
	if !strings.Contains(out, "06/2027") {
		t.Errorf("saída non mostra nova data:\n%s", out)
	}
}

func TestRun_EditsAmount_HappyPath(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "1\n3\n1500\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("repo.updated len = %d, esperabamos 1", len(repo.updated))
	}
	got := repo.updated[0]
	if got.AmountUSD != 1500 || got.Name != "AAPL" || got.Month != 3 || got.Year != 2026 {
		t.Errorf("updated = %+v, esperabamos só Amount cambiado", got)
	}
	if !strings.Contains(out, "1500.00 USD") {
		t.Errorf("saída non mostra nova cantidade:\n%s", out)
	}
}

func TestRun_AcceptsCommaDecimalForAmount(t *testing.T) {
	_, repo, err := runWith(sampleAssets, "1\n3\n1500,50\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.updated) != 1 || repo.updated[0].AmountUSD != 1500.50 {
		t.Errorf("AmountUSD = %v, esperabamos 1500.50", repo.updated)
	}
}

func TestRun_RejectsInvalidFieldChoice(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "1\n9\n1\nNovo\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Selección non válida, escolle 1-3") {
		t.Errorf("saída non contén o erro de selección:\n%s", out)
	}
	// Recupérase e finalmente edita o nome
	if len(repo.updated) != 1 || repo.updated[0].Name != "Novo" {
		t.Errorf("expected Name=Novo tras recuperación, got %+v", repo.updated)
	}
}

func TestRun_RejectsEmptyName(t *testing.T) {
	out, repo, err := runWith(sampleAssets, "1\n1\n\nFinal\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non pode estar baleiro") {
		t.Errorf("saída non contén erro de nome baleiro:\n%s", out)
	}
	if len(repo.updated) != 1 || repo.updated[0].Name != "Final" {
		t.Errorf("expected Name=Final, got %+v", repo.updated)
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
	if len(repo.updated) != 0 {
		t.Errorf("repo.updated debería estar baleiro, got %v", repo.updated)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, repo, err := runWith(sampleAssets, "1\n3\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(repo.updated) != 0 {
		t.Errorf("repo.updated debería estar baleiro, got %v", repo.updated)
	}
}

func TestRun_PropagatesRepoError(t *testing.T) {
	repo := &fakeRepo{
		assets: sampleAssets,
		updEr:  errors.New("boom"),
	}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader("1\n3\n1500\n"))
	err := editasset.Run(r, &buf, repo)
	if err == nil {
		t.Fatal("esperabamos erro do repo")
	}
	if !strings.Contains(err.Error(), "actualizando activo") {
		t.Errorf("erro = %v, esperabamos wrap", err)
	}
}
