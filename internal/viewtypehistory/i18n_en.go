package viewtypehistory

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":                                            "listing assets: %w",
		"obtendo meses de %s: %w":                                         "getting months for %s: %w",
		"calculando resumo de %s: %w":                                     "calculating summary for %s: %w",
		"calculando lifetime de %s: %w":                                   "calculating lifetime totals for %s: %w",
		"\n--- Reporte histórico dun tipo ---\n":                          "\n--- Asset type history report ---\n",
		"Non hai activos de tipo %s.\n":                                   "There are no %s assets.\n",
		"Aínda non hai resultados rexistrados para activos de tipo %s.\n": "There are no results recorded yet for %s assets.\n",
		"  Tipo: %s · %d activo(s) · %d mes(es) con resultado\n":          "  Type: %s · %d asset(s) · %d month(s) with result\n",
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
