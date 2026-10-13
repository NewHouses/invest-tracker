package domain

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		// Tipos de activo.
		"Acción":       "Stock",
		"Índice":       "Index",
		"Copy-trading": "Copy trading",
		"Fondo":        "Fund",

		// Erros da proxección.
		"o salario anual non pode ser negativo":     "the annual salary cannot be negative",
		"o retorno mensual debe ser maior ca -100%": "the monthly return must be greater than -100%",
		"a data de inicio non é válida":             "the start date is not valid",
		"a proxección supera o rango numérico: reduce o retorno, os aportes ou as regras de crecemento": "the projection exceeds the numeric range: reduce the return, the contributions or the growth rules",
		"os anos deben estar entre 1 e %d":                            "the years must be between 1 and %d",
		"o investimento inicial debe ser unha cantidade non negativa": "the initial investment must be a non-negative amount",
		"o retorno mensual debe ser maior ca -100%%":                  "the monthly return must be greater than -100%%",
		"o modo de proxección non é válido":                           "the projection mode is not valid",
		"o aporte mensual debe ser unha cantidade non negativa":       "the monthly contribution must be a non-negative amount",
		"o salario anual debe ser unha cantidade non negativa":        "the annual salary must be a non-negative amount",
		"a porcentaxe de investimento debe estar entre 0 e 100":       "the investment percentage must be between 0 and 100",
		"non se poden definir máis de %d regras":                      "no more than %d rules can be defined",
		"a regra %d ten un tipo de crecemento non válido":             "rule %d has an invalid growth type",
		"a regra %d ten un valor non válido":                          "rule %d has an invalid value",
		"a regra %d ten unha periodicidade non válida":                "rule %d has an invalid frequency",
		"a regra %d ten unha data de inicio non válida":               "rule %d has an invalid start date",
	})
}
