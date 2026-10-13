package welcome

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"invest-tracker/internal/i18n"
)

const Title = "Control de Investimentos"

// LanguageKey é a opción do menú principal para cambiar o idioma.
const LanguageKey = 5

var ErrInvalidOption = errors.New("opción non válida")

type Option struct {
	Key   int
	Label string
}

type Category struct {
	Key     int
	Label   string
	Options []Option
}

func WelcomeMessage() string {
	return "============================================\n" +
		"  " + Title + "\n" +
		"  " + i18n.T("Ferramenta CLI para o seguimento mensual") + "\n" +
		"============================================\n" +
		"\n" +
		i18n.T("Benvida! Escolle unha opción:") + "\n" +
		i18n.T("(escribe ':q' ou 'cancelar' en calquera prompt para voltar ao menú)")
}

func Categories() []Category {
	return []Category{
		{
			Key:   1,
			Label: i18n.T("Operacións con activos"),
			Options: []Option{
				{Key: 1, Label: i18n.T("Engadir activo")},
				{Key: 2, Label: i18n.T("Editar activo")},
				{Key: 3, Label: i18n.T("Eliminar activo")},
			},
		},
		{
			Key:   2,
			Label: i18n.T("Operacións con transaccións"),
			Options: []Option{
				{Key: 1, Label: i18n.T("Engadir transacción")},
				{Key: 2, Label: i18n.T("Engadir transaccións en serie")},
				{Key: 3, Label: i18n.T("Engadir transaccións do mes")},
				{Key: 4, Label: i18n.T("Repartir aporte mensual")},
				{Key: 5, Label: i18n.T("Editar transacción")},
				{Key: 6, Label: i18n.T("Eliminar transacción")},
			},
		},
		{
			Key:   3,
			Label: i18n.T("Operacións de resultados"),
			Options: []Option{
				{Key: 1, Label: i18n.T("Engadir resultado")},
				{Key: 2, Label: i18n.T("Engadir dividendo")},
				{Key: 3, Label: i18n.T("Pechar mes")},
				{Key: 4, Label: i18n.T("Eliminar resultado")},
				{Key: 5, Label: i18n.T("Eliminar dividendo")},
				{Key: 6, Label: i18n.T("Limpar mes")},
			},
		},
		{
			Key:   4,
			Label: i18n.T("Informes"),
			Options: []Option{
				{Key: 1, Label: i18n.T("Reporte histórico dun activo")},
				{Key: 2, Label: i18n.T("Reporte histórico dun tipo")},
				{Key: 3, Label: i18n.T("Ver transaccións dun activo")},
				{Key: 4, Label: i18n.T("Ver informe mensual dun activo")},
				{Key: 5, Label: i18n.T("Ver informe mensual por tipo")},
				{Key: 6, Label: i18n.T("Ver informe mensual total")},
				{Key: 7, Label: i18n.T("Reporte histórico completo")},
				{Key: 8, Label: i18n.T("Ver gráficas")},
				{Key: 9, Label: i18n.T("Proxección a 20 anos")},
			},
		},
	}
}

// TopOptions returns the top-level menu (category labels, the language option
// and "Saír" en 0).
func TopOptions() []Option {
	cats := Categories()
	out := make([]Option, 0, len(cats)+2)
	for _, c := range cats {
		out = append(out, Option{Key: c.Key, Label: c.Label})
	}
	// A etiqueta do idioma vai nos dous idiomas para atopala aínda que non se
	// entenda o idioma actual.
	out = append(out, Option{Key: LanguageKey, Label: "Idioma / Language"})
	out = append(out, Option{Key: 0, Label: i18n.T("Saír")})
	return out
}

// CategoryOptions returns the submenu for catKey (its options + "Voltar" en 0).
// Returns nil if catKey doesn't match any category.
func CategoryOptions(catKey int) []Option {
	for _, c := range Categories() {
		if c.Key == catKey {
			out := make([]Option, 0, len(c.Options)+1)
			out = append(out, c.Options...)
			out = append(out, Option{Key: 0, Label: i18n.T("Voltar")})
			return out
		}
	}
	return nil
}

func Render(opts []Option) string {
	var b strings.Builder
	for _, o := range opts {
		fmt.Fprintf(&b, "  [%d] %s\n", o.Key, o.Label)
	}
	return b.String()
}

func Select(opts []Option, input string) (Option, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || strings.ContainsAny(trimmed, " \t") {
		return Option{}, ErrInvalidOption
	}
	key, err := strconv.Atoi(trimmed)
	if err != nil {
		return Option{}, ErrInvalidOption
	}
	for _, o := range opts {
		if o.Key == key {
			return o, nil
		}
	}
	return Option{}, ErrInvalidOption
}
