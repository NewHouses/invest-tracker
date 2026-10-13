package viewreport

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"calculando informe: %w": "calculating report: %w",
		"listando activos: %w":   "listing assets: %w",

		"\n--- Informe mensual ---\n": "\n--- Monthly report ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"  Investido ata o mes\t%s\n": "  Invested up to the month\t%s\n",
		"  Investido este mes\t%s\n":  "  Invested this month\t%s\n",
		"  No activo\t%s\n":           "  Holding\t%s\n",
		"  Resultado\t%s\n":           "  Result\t%s\n",
		"  Gañanzas/Perdas\t%s\n":     "  Gain/Loss\t%s\n",
		"  Índice\t%+.2f%%\n":         "  Return\t%+.2f%%\n",
		"  Índice\tn/a":               "  Return\tn/a",
		"  Resultado\t—":              "  Result\t—",
		"  Gañanzas/Perdas\t—":        "  Gain/Loss\t—",
		"  Índice\t—":                 "  Return\t—",
	})
}
