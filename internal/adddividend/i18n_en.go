package adddividend

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir dividendo mensual ---\n":                 "\n--- Add monthly dividend ---\n",
		"gardando dividendo: %w":                                "saving dividend: %w",
		"✓ Dividendo gardado #%d: %.2f USD — %02d/%d\n":         "✓ Dividend saved #%d: %.2f USD — %02d/%d\n",
		"Dividendo (USD): ":                                     "Dividend (USD): ",
		"⚠ Dividendo non válido, debe ser un número maior ca 0": "⚠ Invalid dividend, must be a number greater than 0",
	})
}
