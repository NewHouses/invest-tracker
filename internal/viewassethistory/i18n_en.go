package viewassethistory

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"obtendo meses con resultados: %w": "getting months with results: %w",
		"calculando resumo: %w":            "calculating summary: %w",
		"calculando lifetime: %w":          "calculating lifetime totals: %w",
		"listando activos: %w":             "listing assets: %w",

		"\n--- Reporte histórico dun activo ---\n":                                                            "\n--- Asset history report ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.":                           "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Aínda non hai resultados rexistrados para este activo. Engade un coa operación 'Engadir resultado'.": "There are no results recorded for this asset yet. Add one with the 'Add result' operation.",
		"  %s — %s · %d mes(es) con resultado\n":                                                              "  %s — %s · %d month(s) with result\n",
		"  Total Aportado\t%s\n":                                          "  Total Contributed\t%s\n",
		"  Índice Medio Mensual\t%+.2f%%\n":                               "  Average Monthly Return\t%+.2f%%\n",
		"  Gañanzas/Perdas Medias Mensuais\t%s\n":                         "  Average Monthly Gain/Loss\t%s\n",
		"  Índice Medio Mensual\t—":                                       "  Average Monthly Return\t—",
		"  Gañanzas/Perdas Medias Mensuais\t—":                            "  Average Monthly Gain/Loss\t—",
		"  Total Gañanzas/Perdas\t%s\n":                                   "  Total Gain/Loss\t%s\n",
		"  Total Gañanzas/Perdas\t—":                                      "  Total Gain/Loss\t—",
		"  Ano\tMes\tAporte Mensual\tNo activo\tÍndice\tG/P\tResultado\t": "  Year\tMonth\tMonthly Contribution\tHolding\tReturn\tG/L\tResult\t",
	})
}
