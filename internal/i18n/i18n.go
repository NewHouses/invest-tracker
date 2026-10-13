// Package i18n traduce os textos da aplicación entre galego e inglés.
//
// Os textos escríbense en galego e úsanse como clave: T("Engadir activo")
// devolve a tradución rexistrada con Register cando o idioma é o inglés, e o
// propio texto en galego se non. Cada paquete rexistra as súas traducións nun
// ficheiro i18n_en.go:
//
//	func init() {
//		i18n.Register(map[string]string{
//			"Engadir activo": "Add asset",
//		})
//	}
//
// Os formatos de fmt (%s, %d, %.2f...) teñen que ser os mesmos e na mesma orde
// nas dúas versións; o test do catálogo compróbao e tamén que non falte
// ningunha tradución.
//
// A CLI usa o idioma global (SetCurrent, e as funcións T, Sprintf, Fprintf...).
// O servidor web traduce cada petición co idioma do navegador: crea
// lang := i18n.FromAcceptLanguage(...) e usa lang.T, lang.Sprintf e
// lang.Errorf (a variable chámase sempre lang para que o test a recoñeza).
package i18n

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Lang é un idioma da aplicación.
type Lang string

const (
	Galician Lang = "gl"
	English  Lang = "en"
)

var (
	mu      sync.RWMutex
	english = map[string]string{}
	current = Galician
)

// Register engade traducións ao inglés (texto en galego → texto en inglés).
// Chámase desde as funcións init de cada paquete.
func Register(translations map[string]string) {
	mu.Lock()
	defer mu.Unlock()
	for gl, en := range translations {
		if prev, ok := english[gl]; ok && prev != en {
			panic(fmt.Sprintf("i18n: dúas traducións distintas para %q: %q e %q", gl, prev, en))
		}
		english[gl] = en
	}
}

// SetCurrent cambia o idioma global, o da CLI.
func SetCurrent(l Lang) {
	mu.Lock()
	defer mu.Unlock()
	current = l
}

// Current devolve o idioma global, o da CLI.
func Current() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T traduce s (en galego) ao idioma l.
func (l Lang) T(s string) string {
	if l != English {
		return s
	}
	mu.RLock()
	defer mu.RUnlock()
	if en, ok := english[s]; ok {
		return en
	}
	return s
}

// Sprintf formatea co formato traducido ao idioma l.
func (l Lang) Sprintf(format string, args ...any) string {
	return fmt.Sprintf(l.T(format), args...)
}

// Errorf crea un erro co formato traducido ao idioma l (admite %w).
func (l Lang) Errorf(format string, args ...any) error {
	return fmt.Errorf(l.T(format), args...)
}

// T traduce s ao idioma global.
func T(s string) string { return Current().T(s) }

// Sprintf formatea co formato traducido ao idioma global.
func Sprintf(format string, args ...any) string { return Current().Sprintf(format, args...) }

// Errorf crea un erro co formato traducido ao idioma global (admite %w).
func Errorf(format string, args ...any) error { return Current().Errorf(format, args...) }

// Fprintf escribe en w co formato traducido ao idioma global.
func Fprintf(w io.Writer, format string, args ...any) (int, error) {
	return fmt.Fprintf(w, T(format), args...)
}

// Fprint escribe en w o texto traducido ao idioma global.
func Fprint(w io.Writer, s string) (int, error) { return fmt.Fprint(w, T(s)) }

// Fprintln escribe en w o texto traducido ao idioma global e un salto de liña.
func Fprintln(w io.Writer, s string) (int, error) { return fmt.Fprintln(w, T(s)) }

// Parse recoñece un idioma a partir dunha etiqueta como "gl", "en-US" ou
// "gl_ES.UTF-8".
func Parse(s string) (Lang, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "en"):
		return English, true
	case strings.HasPrefix(s, "gl"):
		return Galician, true
	}
	return "", false
}

// FromAcceptLanguage escolle o idioma dunha cabeceira HTTP Accept-Language: o
// primeiro dos admitidos que apareza; galego se non hai ningún.
func FromAcceptLanguage(header string) Lang {
	for _, part := range strings.Split(header, ",") {
		tag, _, _ := strings.Cut(part, ";")
		if l, ok := Parse(tag); ok {
			return l
		}
	}
	return Galician
}

// Detect devolve o idioma do sistema: inglés se o sistema está en inglés e
// galego en calquera outro caso.
func Detect() Lang {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		v := os.Getenv(name)
		if v == "" || v == "C" || v == "POSIX" || strings.HasPrefix(v, "C.") {
			continue
		}
		if l, ok := Parse(v); ok {
			return l
		}
		return Galician
	}
	if l, ok := Parse(systemLocale()); ok {
		return l
	}
	return Galician
}

// preferenceFile é o ficheiro onde se garda o idioma escollido na CLI. Está no
// mesmo cartafol que os datos da aplicación instalada.
func preferenceFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Control de Investimentos", "idioma.txt"), nil
}

// LoadPreference le o idioma escollido na CLI, se se gardou algún.
func LoadPreference() (Lang, bool) {
	path, err := preferenceFile()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return Parse(string(data))
}

// SavePreference garda o idioma escollido na CLI.
func SavePreference(l Lang) error {
	if l != Galician && l != English {
		return errors.New("i18n: idioma descoñecido")
	}
	path, err := preferenceFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(string(l)+"\n"), 0o644)
}
