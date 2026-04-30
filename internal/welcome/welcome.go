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

func WelcomeMessage() string {
	return "============================================\n" +
		"  " + Title + "\n" +
		"  Ferramenta CLI para o seguimento mensual\n" +
		"============================================\n" +
		"\n" +
		"Benvida! Escolle unha opción:"
}

func Options() []Option {
	return []Option{
		{Key: 1, Label: "Engadir novo activo"},
		{Key: 2, Label: "Editar un activo"},
		{Key: 3, Label: "Eliminar activo"},
		{Key: 4, Label: "Engadir nova transacción"},
		{Key: 5, Label: "Eliminar transacción"},
		{Key: 6, Label: "Pechar mes (resultados)"},
		{Key: 7, Label: "Limpar mes"},
		{Key: 8, Label: "Engadir resultado mensual"},
		{Key: 9, Label: "Eliminar resultado mensual"},
		{Key: 10, Label: "Engadir dividendo mensual"},
		{Key: 11, Label: "Eliminar dividendo mensual"},
		{Key: 12, Label: "Ver informe mensual dun activo"},
		{Key: 13, Label: "Ver resultado xeral dun activo"},
		{Key: 14, Label: "Ver informe mensual por tipo"},
		{Key: 15, Label: "Ver resultado xeral (historial total)"},
		{Key: 16, Label: "Ver informe mensual total"},
		{Key: 17, Label: "Ver historial dun activo"},
		{Key: 0, Label: "Saír"},
	}
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
