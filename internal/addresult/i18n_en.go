package addresult

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir resultado mensual ---\n": "\n--- Add monthly result ---\n",
		"listando activos: %w":                  "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Non hai activos con capital investido en %02d/%d.\n":                       "There are no assets with invested capital in %02d/%d.\n",
		"gardando resultado: %w": "saving result: %w",
		"✓ Resultado gardado #%d sobre %s — %s: %.2f USD — %02d/%d\n": "✓ Result saved #%d for %s — %s: %.2f USD — %02d/%d\n",
		"No activo: %.2f USD\n":                                 "Holding: %.2f USD\n",
		"Gañanzas/Perdas: %+.2f USD (%+.2f%%)\n":                "Gain/Loss: %+.2f USD (%+.2f%%)\n",
		"Gañanzas/Perdas: %+.2f USD (n/a%%)\n":                  "Gain/Loss: %+.2f USD (n/a%%)\n",
		"Investimentos:":                                        "Investments:",
		"  [%d] %s — %s (no activo: %.2f USD)\n":                "  [%d] %s — %s (holding: %.2f USD)\n",
		"Selecciona (1-%d): ":                                   "Select (1-%d): ",
		"⚠ Selección non válida":                                "⚠ Invalid selection",
		"Resultado (USD): ":                                     "Result (USD): ",
		"⚠ Resultado non válido, debe ser un número maior ca 0": "⚠ Invalid result, must be a number greater than 0",
	})
}
