package viewcharts

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Ver gráficas ---\n":                 "\n--- View charts ---\n",
		"  [1] Distribución de aportes":            "  [1] Contribution breakdown",
		"  [2] Evolución dun activo":               "  [2] Asset evolution",
		"  [3] Evolución dun tipo (agregada)":      "  [3] Type evolution (aggregated)",
		"  [4] Evolución dos activos dun tipo":     "  [4] Evolution of the assets of a type",
		"  [5] Evolución dos tipos":                "  [5] Evolution by type",
		"  [6] Evolución do resultado total":       "  [6] Total result evolution",
		"  [0] Voltar":                             "  [0] Back",
		"> ":                                       "> ",
		"⚠ Selección non válida":                   "⚠ Invalid selection",
		"↷ Cancelado.":                             "↷ Cancelled.",
		"\n--- Distribución de aportes ---":        "\n--- Contribution breakdown ---",
		"\n--- Evolución dun activo ---":           "\n--- Asset evolution ---",
		"\n--- Evolución dun tipo (agregada) ---":  "\n--- Type evolution (aggregated) ---",
		"\n--- Evolución dos activos dun tipo ---": "\n--- Evolution of the assets of a type ---",
		"\n--- Evolución dos tipos ---":            "\n--- Evolution by type ---",
		"\n--- Evolución do resultado total ---":   "\n--- Total result evolution ---",
		"\nTotal aportado: %s\n\n":                 "\nTotal contributed: %s\n\n",
		"Aínda non hai activos.":                   "There are no assets yet.",
		"Aínda non hai aportes rexistrados.":       "There are no contributions recorded yet.",
		"Aínda non hai resultados rexistrados.":    "There are no results recorded yet.",
		"Non hai activos de tipo %s.\n":            "There are no %s assets.\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Aínda non hai resultados rexistrados para este activo.":                    "There are no results recorded yet for this asset.",
		"Aínda non hai resultados rexistrados para ningún tipo.":                    "There are no results recorded yet for any type.",
		"Aínda non hai resultados rexistrados para activos de tipo %s.\n":           "There are no results recorded yet for %s assets.\n",

		// Títulos e series das gráficas (tamén na web).
		"Resultado":                    "Result",
		"Aporte acumulado":             "Cumulative contribution",
		"Resultado agregado":           "Aggregated result",
		"Resultado + dividendos acum.": "Result + cum. dividends",
		"Resultado por tipo":           "Result by type",
		"Resultado total da carteira":  "Total portfolio result",
		"Todos os activos":             "All assets",
		"Tipo: %s":                     "Type: %s",
		"Activos de tipo %s":           "%s assets",

		"listando activos: %w":          "listing assets: %w",
		"obtendo meses: %w":             "getting months: %w",
		"obtendo meses de %s: %w":       "getting months for %s: %w",
		"calculando resumo: %w":         "calculating summary: %w",
		"calculando resumo de %s: %w":   "calculating summary for %s: %w",
		"calculando lifetime de %s: %w": "calculating lifetime totals for %s: %w",
		"sumando dividendos: %w":        "adding up dividends: %w",
	})
}
