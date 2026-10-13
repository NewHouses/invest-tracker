package repartoaporte

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Repartir aporte mensual ---\n":                                       "\n--- Split monthly contribution ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"\n→ %.2f USD a %s\n":                                                       "\n→ %.2f USD to %s\n",
		"\nTipos dispoñibles:":                                                      "\nAvailable types:",
		"  [%d] %s\n":                                                               "  [%d] %s\n",
		"Selecciona os tipos (separados por comas, e.g. 1,3): ":                     "Select the types (comma separated, e.g. 1,3): ",
		"Selecciona os activos (separados por comas, e.g. 1,2): ":                   "Select the assets (comma separated, e.g. 1,2): ",
		"⚠ Selección non válida (escribe un ou máis números válidos separados por comas)": "⚠ Invalid selection (type one or more valid numbers separated by commas)",
		"⚠ Selección non válida":                  "⚠ Invalid selection",
		"  Reparto de aporte mensual: %.2f USD\n": "  Monthly contribution split: %.2f USD\n",
		"  %s\t%.2f USD\n":                        "  %s\t%.2f USD\n",
		"    %s\t%.2f USD\n":                      "    %s\t%.2f USD\n",

		"a cantidade total debe ser maior ca 0":   "the total amount must be greater than 0",
		"debes seleccionar polo menos un tipo":    "you must select at least one type",
		"o tipo %s non ten activos seleccionados": "the type %s has no selected assets",
		"listando activos: %w":                    "listing assets: %w",
		"non é un número: %s":                     "not a number: %s",
		"fóra de rango: %d":                       "out of range: %d",
	})
}
