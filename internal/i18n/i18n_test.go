package i18n

import (
	"errors"
	"strings"
	"testing"
)

func TestTranslate(t *testing.T) {
	Register(map[string]string{
		"proba: Ola %s, tes %d activos": "test: Hello %s, you have %d assets",
		"proba: erro lendo %s: %w":      "test: error reading %s: %w",
	})

	if got := Galician.T("proba: Ola %s, tes %d activos"); got != "proba: Ola %s, tes %d activos" {
		t.Errorf("galego = %q, esperabamos o texto orixinal", got)
	}
	if got := English.Sprintf("proba: Ola %s, tes %d activos", "Ana", 3); got != "test: Hello Ana, you have 3 assets" {
		t.Errorf("inglés = %q", got)
	}
	if got := English.T("proba: sen tradución"); got != "proba: sen tradución" {
		t.Errorf("sen tradución = %q, esperabamos o texto en galego", got)
	}
	base := errors.New("boom")
	err := English.Errorf("proba: erro lendo %s: %w", "x.db", base)
	if err.Error() != "test: error reading x.db: boom" || !errors.Is(err, base) {
		t.Errorf("Errorf = %v (Is=%v)", err, errors.Is(err, base))
	}

	defer SetCurrent(Current())
	SetCurrent(English)
	var b strings.Builder
	_, _ = Fprintf(&b, "proba: Ola %s, tes %d activos", "Ana", 1)
	if b.String() != "test: Hello Ana, you have 1 assets" {
		t.Errorf("Fprintf co idioma global = %q", b.String())
	}
}

func TestRegisterRejectsConflictingTranslations(t *testing.T) {
	Register(map[string]string{"proba: conflito": "test: conflict"})
	Register(map[string]string{"proba: conflito": "test: conflict"}) // a mesma tradución non é un erro
	defer func() {
		if recover() == nil {
			t.Fatal("esperabamos un panic por dúas traducións distintas")
		}
	}()
	Register(map[string]string{"proba: conflito": "test: other"})
}

func TestParseAndAcceptLanguage(t *testing.T) {
	for in, want := range map[string]Lang{"gl": Galician, "gl_ES.UTF-8": Galician, "en-US": English, "EN": English} {
		if got, ok := Parse(in); !ok || got != want {
			t.Errorf("Parse(%q) = %q, %v; esperabamos %q", in, got, ok, want)
		}
	}
	if _, ok := Parse("es-ES"); ok {
		t.Error("o castelán non é un idioma da aplicación")
	}
	for header, want := range map[string]Lang{
		"en":                 English,
		"en-GB,en;q=0.9":     English,
		"gl":                 Galician,
		"es-ES,es;q=0.9":     Galician,
		"es-ES,en-US;q=0.8":  English,
		"":                   Galician,
		"gl-ES,en-US;q=0.5,": Galician,
	} {
		if got := FromAcceptLanguage(header); got != want {
			t.Errorf("FromAcceptLanguage(%q) = %q, esperabamos %q", header, got, want)
		}
	}
}

func TestDetectFromEnvironment(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := Detect(); got != English {
		t.Errorf("LANG=en_US: %q, esperabamos inglés", got)
	}
	t.Setenv("LANG", "es_ES.UTF-8")
	if got := Detect(); got != Galician {
		t.Errorf("LANG=es_ES: %q, esperabamos galego", got)
	}
	t.Setenv("LC_ALL", "gl_ES.UTF-8")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := Detect(); got != Galician {
		t.Errorf("LC_ALL ten prioridade: %q, esperabamos galego", got)
	}
}

func TestPreference(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)         // Windows
	t.Setenv("XDG_CONFIG_HOME", dir) // Linux
	t.Setenv("HOME", dir)            // macOS
	if _, ok := LoadPreference(); ok {
		t.Fatal("sen ficheiro non debería haber idioma gardado")
	}
	if err := SavePreference(English); err != nil {
		t.Fatalf("SavePreference: %v", err)
	}
	if got, ok := LoadPreference(); !ok || got != English {
		t.Fatalf("LoadPreference = %q, %v; esperabamos inglés", got, ok)
	}
	if err := SavePreference("fr"); err == nil {
		t.Fatal("un idioma descoñecido debería dar erro")
	}
}
