package viewtypereport

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":        "listing assets: %w",
		"calculando resumo de %s: %w": "calculating summary for %s: %w",

		"\n--- Informe mensual por tipo ---\n":                           "\n--- Monthly report by type ---\n",
		"Non hai activos de tipo %s.\n":                                  "There are no %s assets.\n",
		"Activos de tipo %s:\n":                                          "%s assets:\n",
		"Non hai activos de tipo %s con capital investido en %02d/%d.\n": "There are no %s assets with invested capital in %02d/%d.\n",
		"  Tipo: %s · %02d/%d\n":                                         "  Type: %s · %02d/%d\n",
		"  Activos incluídos: %d\n":                                      "  Included assets: %d\n",
		"    - %s (no activo: %s)\n":                                     "    - %s (holding: %s)\n",
		"  Investido ata o mes\t%s\n":                                    "  Invested up to the month\t%s\n",
		"  Investido este mes\t%s\n":                                     "  Invested this month\t%s\n",
		"  No activo\t%s\n":                                              "  Holding\t%s\n",
		"  Resultado\t—":                                                 "  Result\t—",
		"  Gañanzas/Perdas\t—":                                           "  Gain/Loss\t—",
		"  Índice\t—":                                                    "  Return\t—",
		"  Resultado\t%s\n":                                              "  Result\t%s\n",
		"  Gañanzas/Perdas\t%s\n":                                        "  Gain/Loss\t%s\n",
		"  Índice\t%+.2f%%\n":                                            "  Return\t%+.2f%%\n",
		"  Índice\tn/a":                                                  "  Return\tn/a",
		"  Resultado (parc.)\t%s  (%d/%d activos)\n":                     "  Result (partial)\t%s  (%d/%d assets)\n",
		"  Gañanzas/Perdas (parc.)\t%s\n":                                "  Gain/Loss (partial)\t%s\n",
		"  Índice (parc.)\t%+.2f%%\n":                                    "  Return (partial)\t%+.2f%%\n",
		"  Índice (parc.)\tn/a":                                          "  Return (partial)\tn/a",
	})
}
