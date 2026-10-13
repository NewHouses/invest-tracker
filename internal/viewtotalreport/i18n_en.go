package viewtotalreport

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":             "listing assets: %w",
		"obtendo meses con resultados: %w": "getting months with results: %w",
		"calculando %s: %w":                "calculating %s: %w",

		"\n--- Informe mensual total ---\n":                                         "\n--- Total monthly report ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Activos:": "Assets:",
		"Non hai activos con capital investido en %02d/%d.\n":            "There are no assets with invested capital in %02d/%d.\n",
		"  Informe total · %02d/%d\n":                                    "  Total report · %02d/%d\n",
		"  Activos: %d (%d activos en %02d/%d)\n":                        "  Assets: %d (%d assets in %02d/%d)\n",
		"  Total investido ata o mes\t%s\n":                              "  Total invested up to the month\t%s\n",
		"  Investido este mes\t%s\n":                                     "  Invested this month\t%s\n",
		"  Investimento + dividendos prev. mes\t%s\n":                    "  Investment + previous month's dividends\t%s\n",
		"  No activo (sen div)\t%s\n":                                    "  Holding (without div.)\t%s\n",
		"  No activo (con div)\t%s\n":                                    "  Holding (with div.)\t%s\n",
		"  Dividendos este mes\t%s\n":                                    "  Dividends this month\t%s\n",
		"  Resultado (sen div)\t—":                                       "  Result (without div.)\t—",
		"  Resultado total (con div)\t—":                                 "  Total result (with div.)\t—",
		"  Gañanzas/Perdas\t—":                                           "  Gain/Loss\t—",
		"  Gañanzas/Perdas (con div)\t—":                                 "  Gain/Loss (with div.)\t—",
		"  Índice\t—":                                                    "  Return\t—",
		"  Índice (con div)\t—":                                          "  Return (with div.)\t—",
		"  (%d/%d activos con resultado)":                                "  (%d/%d assets with result)",
		"  Resultado (sen div)\t%s%s\n":                                  "  Result (without div.)\t%s%s\n",
		"  Resultado total (con div)\t%s\n":                              "  Total result (with div.)\t%s\n",
		"  Gañanzas/Perdas\t%s\n":                                        "  Gain/Loss\t%s\n",
		"  Gañanzas/Perdas (con div)\t%s\n":                              "  Gain/Loss (with div.)\t%s\n",
		"  Índice\t%+.2f%%\n":                                            "  Return\t%+.2f%%\n",
		"  Índice (con div)\t%+.2f%%\n":                                  "  Return (with div.)\t%+.2f%%\n",
		"  Índice (con div)\tn/a":                                        "  Return (with div.)\tn/a",
		"  Gañanzas/Perdas\tn/a":                                         "  Gain/Loss\tn/a",
		"  Gañanzas/Perdas (con div)\tn/a":                               "  Gain/Loss (with div.)\tn/a",
		"  Índice\tn/a":                                                  "  Return\tn/a",
		"  Promedios mensuais: sen meses con resultados.":                "  Monthly averages: no months with results.",
		"  Promedios mensuais (%d mes(es) con resultado ata %02d/%d):\n": "  Monthly averages (%d month(s) with result up to %02d/%d):\n",
		"    Índice medio mensual (sen div)\t%+.2f%%\n":                  "    Average monthly return (without div.)\t%+.2f%%\n",
		"    Gañanza media mensual (sen div)\t%s\n":                      "    Average monthly gain (without div.)\t%s\n",
		"    Índice medio mensual (con div)\t%+.2f%%\n":                  "    Average monthly return (with div.)\t%+.2f%%\n",
		"    Gañanza media mensual (con div)\t%s\n":                      "    Average monthly gain (with div.)\t%s\n",
	})
}
