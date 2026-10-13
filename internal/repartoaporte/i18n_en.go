package repartoaporte

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Repartir aporte mensual ---\n":                                             "\n--- Split monthly contribution ---\n",
		"Aínda non hai activos creados: o reparto farase entre activos novos.":            "There are no assets created yet: the split will be among new assets.",
		"\nTipos dispoñibles:":                                                            "\nAvailable types:",
		"  [%d] %s\n":                                                                     "  [%d] %s\n",
		"Selecciona os tipos (separados por comas, e.g. 1,3): ":                           "Select the types (comma separated, e.g. 1,3): ",
		"Selecciona os activos (separados por comas, e.g. 1,2): ":                         "Select the assets (comma separated, e.g. 1,2): ",
		"⚠ Selección non válida (escribe un ou máis números válidos separados por comas)": "⚠ Invalid selection (type one or more valid numbers separated by commas)",
		"⚠ Selección non válida":                                                          "⚠ Invalid selection",
		"  Reparto de aporte mensual: %.2f USD\n":                                         "  Monthly contribution split: %.2f USD\n",

		// Activos novos, aínda sen crear.
		"Queres engadir activos que aínda non creaches? [s/N]: ": "Do you want to add assets you have not created yet? [y/N]: ",
		"\nActivo novo (aínda sen crear):":                       "\nNew asset (not created yet):",
		"Nome: ":                                                 "Name: ",
		"⚠ O nome non pode estar baleiro":                        "⚠ The name cannot be empty",
		"Engadir outro activo novo? [s/N]: ":                     "Add another new asset? [y/N]: ",
		" (novo)":                                                " (new)",

		// Porcentaxes de cada activo.
		"\nReparto proposto:":                                       "\nProposed split:",
		"Queres cambiar as porcentaxes? [s/N]: ":                    "Do you want to change the percentages? [y/N]: ",
		"⚠ Resposta non válida, escolle S ou N":                     "⚠ Invalid response, choose Y or N",
		"Porcentaxe de cada activo (Intro para manter a que ten):":  "Percentage for each asset (press Enter to keep the current one):",
		"⚠ As porcentaxes suman %.2f %%; teñen que sumar 100 %%.\n": "⚠ The percentages add up to %.2f%%; they must add up to 100%%.\n",
		"⚠ Porcentaxe non válida, debe ser un número entre 0 e 100": "⚠ Invalid percentage, must be a number between 0 and 100",

		"a cantidade total debe ser maior ca 0":                "the total amount must be greater than 0",
		"debes seleccionar polo menos un tipo":                 "you must select at least one type",
		"o tipo %s non ten activos seleccionados":              "the type %s has no selected assets",
		"indica a porcentaxe de cada activo":                   "enter the percentage for each asset",
		"a porcentaxe de %s debe estar entre 0 e 100":          "the percentage for %s must be between 0 and 100",
		"as porcentaxes suman %.2f %%; teñen que sumar 100 %%": "the percentages add up to %.2f%%; they must add up to 100%%",
		"listando activos: %w":                                 "listing assets: %w",
		"non é un número: %s":                                  "not a number: %s",
		"fóra de rango: %d":                                    "out of range: %d",
	})
}
