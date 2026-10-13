package viewtotalhistory

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":                   "listing assets: %w",
		"obtendo meses con resultados: %w":       "getting months with results: %w",
		"calculando resumo de %s: %w":            "calculating summary for %s: %w",
		"sumando dividendos de %d/%d: %w":        "adding up dividends for %d/%d: %w",
		"calculando lifetime de %s: %w":          "calculating lifetime totals for %s: %w",
		"\n--- Reporte histórico completo ---\n": "\n--- Full history report ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.":                                           "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Aínda non hai resultados rexistrados. Engade resultados mensuais coa operación 'Engadir resultado' ou 'Pechar mes'.": "There are no results recorded yet. Add monthly results with the 'Add result' or 'Close month' operation.",
		"  Reporte histórico completo · %d activo(s) · %d mes(es) con resultado\n":                                            "  Full history report · %d asset(s) · %d month(s) with result\n",
		"  Aporte histórico total\t%s\n": "  Total contributions\t%s\n",
		"  Índice Medio\t%+.2f%%\n":      "  Average Return\t%+.2f%%\n",
		"  G/P Media\t%s\n":              "  Average G/L\t%s\n",
		"  Índice Medio\t—":              "  Average Return\t—",
		"  G/P Media\t—":                 "  Average G/L\t—",
		"  G/P Total\t%s\n":              "  Total G/L\t%s\n",
		"  G/P Total\t—":                 "  Total G/L\t—",
		"  Dividendos totais\t%s\n":      "  Total dividends\t%s\n",
		"  Ano\tMes\tAporte Mensual\tFondos\tÍndice\tG/P\tDividendos\tResultado\t": "  Year\tMonth\tMonthly Contribution\tHoldings\tReturn\tG/L\tDividends\tResult\t",
	})
}
