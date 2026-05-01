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
		"Benvida! Escolle unha opción:\n" +
		"(escribe ':q' ou 'cancelar' en calquera prompt para voltar ao menú)"
	if got := WelcomeMessage(); got != want {
		t.Errorf("WelcomeMessage mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestCategories_Stable(t *testing.T) {
	want := []Category{
		{Key: 1, Label: "Operacións con activos", Options: []Option{
			{Key: 1, Label: "Engadir activo"},
			{Key: 2, Label: "Editar activo"},
			{Key: 3, Label: "Eliminar activo"},
		}},
		{Key: 2, Label: "Operacións con transaccións", Options: []Option{
			{Key: 1, Label: "Engadir transacción"},
			{Key: 2, Label: "Engadir transaccións en serie"},
			{Key: 3, Label: "Editar transacción"},
			{Key: 4, Label: "Eliminar transacción"},
		}},
		{Key: 3, Label: "Operacións de resultados", Options: []Option{
			{Key: 1, Label: "Engadir resultado"},
			{Key: 2, Label: "Engadir dividendo"},
			{Key: 3, Label: "Pechar mes"},
			{Key: 4, Label: "Eliminar resultado"},
			{Key: 5, Label: "Eliminar dividendo"},
			{Key: 6, Label: "Limpar mes"},
		}},
		{Key: 4, Label: "Informes", Options: []Option{
			{Key: 1, Label: "Ver historial dun activo"},
			{Key: 2, Label: "Ver transaccións dun activo"},
			{Key: 3, Label: "Ver informe mensual dun activo"},
			{Key: 4, Label: "Ver resultado xeral dun activo"},
			{Key: 5, Label: "Ver informe mensual por tipo"},
			{Key: 6, Label: "Ver informe mensual total"},
			{Key: 7, Label: "Ver resultado xeral total"},
		}},
	}
	got := Categories()
	if len(got) != len(want) {
		t.Fatalf("Categories() devolveu %d entradas, esperabamos %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Key != w.Key || got[i].Label != w.Label {
			t.Errorf("Categories()[%d] = {%d, %q}, queremos {%d, %q}",
				i, got[i].Key, got[i].Label, w.Key, w.Label)
			continue
		}
		if len(got[i].Options) != len(w.Options) {
			t.Errorf("Categories()[%d].Options len = %d, esperabamos %d",
				i, len(got[i].Options), len(w.Options))
			continue
		}
		for j, oj := range w.Options {
			if got[i].Options[j] != oj {
				t.Errorf("Categories()[%d].Options[%d] = %+v, queremos %+v",
					i, j, got[i].Options[j], oj)
			}
		}
	}
}

func TestTopOptions_ContainsAllCategoriesAndSair(t *testing.T) {
	top := TopOptions()
	if len(top) != len(Categories())+1 {
		t.Fatalf("TopOptions len = %d, esperabamos %d", len(top), len(Categories())+1)
	}
	for i, cat := range Categories() {
		if top[i].Key != cat.Key || top[i].Label != cat.Label {
			t.Errorf("TopOptions[%d] = %+v, queremos {%d, %q}",
				i, top[i], cat.Key, cat.Label)
		}
	}
	last := top[len(top)-1]
	if last.Key != 0 || last.Label != "Saír" {
		t.Errorf("TopOptions[last] = %+v, queremos {0, Saír}", last)
	}
}

func TestCategoryOptions_HasVoltar(t *testing.T) {
	for _, cat := range Categories() {
		opts := CategoryOptions(cat.Key)
		if len(opts) != len(cat.Options)+1 {
			t.Errorf("CategoryOptions(%d) len = %d, esperabamos %d",
				cat.Key, len(opts), len(cat.Options)+1)
			continue
		}
		last := opts[len(opts)-1]
		if last.Key != 0 || last.Label != "Voltar" {
			t.Errorf("CategoryOptions(%d) last = %+v, queremos {0, Voltar}", cat.Key, last)
		}
	}
}

func TestCategoryOptions_UnknownReturnsNil(t *testing.T) {
	if got := CategoryOptions(99); got != nil {
		t.Errorf("CategoryOptions(99) = %v, queremos nil", got)
	}
}

func TestRender_ContainsEveryOption(t *testing.T) {
	opts := TopOptions()
	rendered := Render(opts)
	for _, o := range opts {
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
	opts := TopOptions()
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
	opts := TopOptions()
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
	opts := TopOptions()
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
