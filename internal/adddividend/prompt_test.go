package adddividend_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"invest-tracker/internal/adddividend"
	"invest-tracker/internal/domain"
)

type fakeSaver struct {
	saved []domain.Dividend
	err   error
}

func (f *fakeSaver) InsertDividend(d domain.Dividend) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.saved = append(f.saved, d)
	return int64(len(f.saved)), nil
}

func runWith(input string) (string, *fakeSaver, error) {
	saver := &fakeSaver{}
	var buf bytes.Buffer
	r := bufio.NewReader(strings.NewReader(input))
	err := adddividend.Run(r, &buf, saver)
	return buf.String(), saver, err
}

const validInput = "125.50\n4\n2026\n"

func TestRun_PrintsHeader(t *testing.T) {
	out, _, err := runWith(validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Engadir dividendo mensual") {
		t.Errorf("saída non contén a cabeceira:\n%s", out)
	}
}

func TestRun_PromptsAllFields(t *testing.T) {
	out, _, err := runWith(validInput)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	expects := []string{
		"Dividendo (USD):",
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
	if !strings.Contains(out, "✓ Dividendo gardado") {
		t.Errorf("saída non contén a confirmación:\n%s", out)
	}
	if !strings.Contains(out, "#1") {
		t.Errorf("saída non contén o ID:\n%s", out)
	}
	if !strings.Contains(out, "125.50 USD") {
		t.Errorf("saída non contén a cantidade:\n%s", out)
	}
	if len(saver.saved) != 1 {
		t.Fatalf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
}

func TestRun_HappyPath_SavesCorrectDividend(t *testing.T) {
	_, saver, err := runWith("125.50\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(saver.saved) != 1 {
		t.Fatalf("saver.saved tamaño = %d, esperabamos 1", len(saver.saved))
	}
	got := saver.saved[0]
	want := domain.Dividend{
		AmountUSD: 125.50,
		Month:     4,
		Year:      2026,
	}
	if got != want {
		t.Errorf("guardado = %+v, queremos %+v", got, want)
	}
}

func TestRun_AcceptsCommaDecimal(t *testing.T) {
	_, saver, err := runWith("125,50\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(saver.saved) != 1 || saver.saved[0].AmountUSD != 125.50 {
		t.Errorf("AmountUSD = %v, esperabamos 125.50", saver.saved)
	}
}

func TestRun_PrintsErrorOnInvalidDividend(t *testing.T) {
	out, saver, err := runWith("abc\n0\n125\n4\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "Dividendo non válido") {
		t.Errorf("saída non contén o erro de dividendo:\n%s", out)
	}
	if len(saver.saved) != 1 || saver.saved[0].AmountUSD != 125 {
		t.Errorf("expected 125, got %+v", saver.saved)
	}
}

func TestRun_PrintsErrorOnInvalidMonth(t *testing.T) {
	out, saver, err := runWith("125\n13\n4\n2026\n")
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
	out, saver, err := runWith("125\n4\n1800\n2026\n")
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

func TestRun_RecoversFromInvalidInput(t *testing.T) {
	_, saver, err := runWith("abc\n200,75\n6\n2026\n")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := domain.Dividend{AmountUSD: 200.75, Month: 6, Year: 2026}
	if len(saver.saved) != 1 || saver.saved[0] != want {
		t.Errorf("got %+v, esperabamos [%+v]", saver.saved, want)
	}
}

func TestRun_EOFMidFlow_ReturnsError(t *testing.T) {
	_, saver, err := runWith("125\n4\n")
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
