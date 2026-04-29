package welcome

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestWelcomeMessage_ContainsTitle(t *testing.T) {
	if !strings.Contains(WelcomeMessage(), Title) {
		t.Errorf("WelcomeMessage debe conter o título %q", Title)
	}
}

func TestWelcomeMessage_ContainsGreeting(t *testing.T) {
	if !strings.Contains(WelcomeMessage(), "Benvida") {
		t.Error("WelcomeMessage debe conter 'Benvida'")
	}
}

func TestWelcomeMessage_Snapshot(t *testing.T) {
	want := "============================================\n" +
		"  Control de Investimentos\n" +
		"  Ferramenta CLI para o seguimento mensual\n" +
		"============================================\n" +
		"\n" +
		"Benvida! Escolle unha opción:"
	if got := WelcomeMessage(); got != want {
		t.Errorf("WelcomeMessage mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestOptions_Stable(t *testing.T) {
	want := []Option{
		{Key: 1, Label: "Engadir novo activo"},
		{Key: 2, Label: "Engadir nova transacción"},
		{Key: 3, Label: "Engadir resultado mensual"},
		{Key: 4, Label: "Pechar mes"},
		{Key: 5, Label: "Ver informe mensual"},
		{Key: 6, Label: "Exportar (CSV / HTML)"},
		{Key: 0, Label: "Saír"},
	}
	got := Options()
	if len(got) != len(want) {
		t.Fatalf("Options() devolveu %d entradas, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("Options()[%d] = %+v, queremos %+v", i, got[i], w)
		}
	}
}

func TestRender_ContainsEveryOption(t *testing.T) {
	rendered := Render(Options())
	for _, o := range Options() {
		keyMarker := fmt.Sprintf("[%d]", o.Key)
		if !strings.Contains(rendered, keyMarker) {
			t.Errorf("Render non contén a clave %s", keyMarker)
		}
		if !strings.Contains(rendered, o.Label) {
			t.Errorf("Render non contén a etiqueta %q", o.Label)
		}
	}
}

func TestSelect_AllValidKeys(t *testing.T) {
	opts := Options()
	for _, o := range opts {
		input := fmt.Sprint(o.Key)
		got, err := Select(opts, input)
		if err != nil {
			t.Errorf("Select(%q) erro inesperado: %v", input, err)
			continue
		}
		if got != o {
			t.Errorf("Select(%q) = %+v, queremos %+v", input, got, o)
		}
	}
}

func TestSelect_TrimsWhitespace(t *testing.T) {
	cases := []struct {
		input   string
		wantKey int
	}{
		{"  3  ", 3},
		{"\t1 ", 1},
		{"3\n", 3},
		{" 0\r\n", 0},
	}
	opts := Options()
	for _, c := range cases {
		got, err := Select(opts, c.input)
		if err != nil {
			t.Errorf("Select(%q) erro inesperado: %v", c.input, err)
			continue
		}
		if got.Key != c.wantKey {
			t.Errorf("Select(%q).Key = %d, queremos %d", c.input, got.Key, c.wantKey)
		}
	}
}

func TestSelect_InvalidInputs(t *testing.T) {
	opts := Options()
	cases := []string{
		"",
		"   ",
		"abc",
		"99",
		"-1",
		"1 2",
		"1,2",
		"1.0",
	}
	for _, in := range cases {
		_, err := Select(opts, in)
		if !errors.Is(err, ErrInvalidOption) {
			t.Errorf("Select(%q) erro = %v, queremos ErrInvalidOption", in, err)
		}
	}
}
