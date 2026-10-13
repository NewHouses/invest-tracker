package closemonth

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Pechar mes (resultados) ---\n":                                       "\n--- Close month (results) ---\n",
		"Non hai activos con capital investido para %02d/%d.\n":                     "There are no assets with invested capital for %02d/%d.\n",
		"Vanse pedir os resultados para %d activo(s) en %02d/%d.\n":                 "Results will be requested for %d asset(s) in %02d/%d.\n",
		"(Deixa baleiro para saltar un activo.)":                                    "(Leave blank to skip an asset.)",
		"\n[%d/%d] %s — %s (no activo: %.2f USD)\n":                                 "\n[%d/%d] %s — %s (holding: %.2f USD)\n",
		"   Xa hai un resultado rexistrado este mes: %.2f USD. Baleiro mantenno.\n": "   There is already a result recorded this month: %.2f USD. Blank keeps it.\n",
		"   ↷ Saltado.":                  "   ↷ Skipped.",
		"gardando resultado para %s: %w": "saving result for %s: %w",
		"   ✓ Gardado #%d — Gañanzas/Perdas: %+.2f USD (%+.2f%%)\n":          "   ✓ Saved #%d — Gain/Loss: %+.2f USD (%+.2f%%)\n",
		"   ✓ Gardado #%d — Gañanzas/Perdas: %+.2f USD (n/a%%)\n":            "   ✓ Saved #%d — Gain/Loss: %+.2f USD (n/a%%)\n",
		"\n✓ Pechouse %02d/%d: %d resultado(s) gardado(s), %d saltado(s).\n": "\n✓ Closed %02d/%d: %d result(s) saved, %d skipped.\n",
		"listando activos: %w":                                             "listing assets: %w",
		"calculando resumo de %s: %w":                                      "calculating summary for %s: %w",
		"   Resultado (USD, baleiro = saltar): ":                           "   Result (USD, blank = skip): ",
		"   ⚠ Resultado non válido (debe ser > 0, ou baleiro para saltar)": "   ⚠ Invalid result (must be > 0, or blank to skip)",
	})
}
