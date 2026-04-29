package addasset_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/addasset"
	"invest-tracker/internal/domain"
)

type fakeSaver struct {
	saved []domain.Investment
	err   error
}

func (f *fakeSaver) InsertInvestment(inv domain.Investment) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.saved = append(f.saved, inv)
	return int64(len(f.saved)), nil
}

func runWith(input string) (string, *fakeSaver, error) {
	saver := &fakeSaver{}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := addasset.Run(r, &buf, saver)
	return buf.String(), saver, err
}

const validInput = "1\nAAPL\n1234.56\n4\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir novo activo") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	out, _, err := runWith(validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"[1] Acción",
		"[2] Índice",
		"[3] Copy-trading",
		"[4] Fondo",
		"Nome:",
		"Cantidade (USD):",
		"Mes (1-12):",
		"Ano:",
	}
	for _, e := range expects {
		if !strings.Contains(out, e) {
			t.Errorf("saída non contén %q:\n%s", e, out)
		}
	}
}

func TestRun_PrintsConfirmation(t *testing.T) {
	out, saver, err := runWith(validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "✓ Investimento gardado") {
		t.Errorf("saída non contén a confirmación:\n%s", out)
	}
	if !strings.Contains(out, "AAPL") {
		t.Errorf("saída non contén o nome:\n%s", out)
	}
	if !strings.Contains(out, "#1") {
		t.Errorf("saída non contén o ID:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Fatalf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_HappyPath_SavesCorrectInvestment(t *testing.T) {
	_, saver, err := runWith("2\nVanguard S&P 500\n1500.50\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(saver.saved) != 1 {
		t.Fatalf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
	got := saver.saved[0]
	want := domain.Investment{
		Type:      domain.Indice,
		Name:      "Vanguard S&P 500",
		AmountUSD: 1500.50,
		Month:     4,
		Year:      2026,
	}
	if got != want {
		t.Errorf("guardado = %+v, queremos %+v", got, want)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	_, saver, err := runWith("1\nAAPL\n1500,50\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(saver.saved) != 1 || saver.saved[0].AmountUSD != 1500.50 {
		t.Errorf("AmountUSD = %v, esperabamos 1500.50", saver.saved)
	}
}

func TestRun_PrintsErrorOnInvalidType(t *testing.T) {
	out, saver, err := runWith("9\n1\nAAPL\n100\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Tipo non válido") {
		t.Errorf("saída non contén o erro de tipo:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Errorf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_PrintsErrorOnEmptyName(t *testing.T) {
	out, saver, err := runWith("1\n\nAAPL\n100\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "non pode estar baleiro") {
		t.Errorf("saída non contén o erro de nome baleiro:\n%s", out)
	}
	if len(saver.saved) != 1 || saver.saved[0].Name != "AAPL" {
		t.Errorf("Name esperado AAPL, got %v", saver.saved)
	}
}

func TestRun_PrintsErrorOnInvalidAmount(t *testing.T) {
	out, saver, err := runWith("1\nAAPL\nabc\n100\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Cantidade non válida") {
		t.Errorf("saída non contén o erro de cantidade:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Errorf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_PrintsErrorOnInvalidMonth(t *testing.T) {
	out, saver, err := runWith("1\nAAPL\n100\n13\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Mes non válido") {
		t.Errorf("saída non contén o erro de mes:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Errorf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_PrintsErrorOnInvalidYear(t *testing.T) {
	out, saver, err := runWith("1\nAAPL\n100\n4\n1800\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Ano non válido") {
		t.Errorf("saída non contén o erro de ano:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Errorf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_RecoversFromInvalidThenValid(t *testing.T) {
	_, saver, err := runWith("9\n2\nVanguard\n1500\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(saver.saved) != 1 || saver.saved[0].Type != domain.Indice {
		t.Errorf("expected Indice saved, got %+v", saver.saved)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, saver, err := runWith("1\nAAPL\n100\n4\n")
	if err == nil {
		t.Fatal("esperabamos erro por entrada truncada")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("esperabamos io.EOF, got %v", err)
	}
	if len(saver.saved) != 0 {
		t.Errorf("saver debería estar baleiro, got %v", saver.saved)
	}
}
