package welcome

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const Title = "Control de Investimentos"

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
		"  Ferramenta CLI para o seguimento mensual\n" +
		"============================================\n" +
		"\n" +
		"Benvida! Escolle unha opción:\n" +
		"(escribe ':q' ou 'cancelar' en calquera prompt para voltar ao menú)"
}

func Categories() []Category {
	return []Category{
		{
			Key:   1,
			Label: "Operacións con activos",
			Options: []Option{
				{Key: 1, Label: "Engadir activo"},
				{Key: 2, Label: "Editar activo"},
				{Key: 3, Label: "Eliminar activo"},
			},
		},
		{
			Key:   2,
			Label: "Operacións con transaccións",
			Options: []Option{
				{Key: 1, Label: "Engadir transacción"},
				{Key: 2, Label: "Engadir transaccións en serie"},
				{Key: 3, Label: "Editar transacción"},
				{Key: 4, Label: "Eliminar transacción"},
			},
		},
		{
			Key:   3,
			Label: "Operacións de resultados",
			Options: []Option{
				{Key: 1, Label: "Engadir resultado"},
				{Key: 2, Label: "Engadir dividendo"},
				{Key: 3, Label: "Pechar mes"},
				{Key: 4, Label: "Eliminar resultado"},
				{Key: 5, Label: "Eliminar dividendo"},
				{Key: 6, Label: "Limpar mes"},
			},
		},
		{
			Key:   4,
			Label: "Informes",
			Options: []Option{
				{Key: 1, Label: "Reporte histórico dun activo"},
				{Key: 2, Label: "Ver transaccións dun activo"},
				{Key: 3, Label: "Ver informe mensual dun activo"},
				{Key: 4, Label: "Ver informe mensual por tipo"},
				{Key: 5, Label: "Ver informe mensual total"},
				{Key: 6, Label: "Ver resultado xeral total"},
			},
		},
	}
}

// TopOptions returns the top-level menu (category labels + "Saír" en 0).
func TopOptions() []Option {
	cats := Categories()
	out := make([]Option, 0, len(cats)+1)
	for _, c := range cats {
		out = append(out, Option{Key: c.Key, Label: c.Label})
	}
	out = append(out, Option{Key: 0, Label: "Saír"})
	return out
}

// CategoryOptions returns the submenu for catKey (its options + "Voltar" en 0).
// Returns nil if catKey doesn't match any category.
func CategoryOptions(catKey int) []Option {
	for _, c := range Categories() {
		if c.Key == catKey {
			out := make([]Option, 0, len(c.Options)+1)
			out = append(out, c.Options...)
			out = append(out, Option{Key: 0, Label: "Voltar"})
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
