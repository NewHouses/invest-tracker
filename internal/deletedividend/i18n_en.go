package deletedividend

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Eliminar dividendo mensual ---\n": "\n--- Delete monthly dividend ---\n",
		"listando dividendos: %w":                "listing dividends: %w",
		"Aínda non hai dividendos rexistrados. Engade un primeiro coa operación 'Engadir dividendo'.": "There are no dividends recorded yet. Add one first with the 'Add dividend' operation.",
		"eliminando dividendo: %w":                         "deleting dividend: %w",
		"✓ Dividendo #%d eliminado (%02d/%d — %.2f USD)\n": "✓ Dividend #%d deleted (%02d/%d — %.2f USD)\n",
		"Dividendos:":                                "Dividends:",
		"  [%d] %02d/%d — %.2f USD\n":                "  [%d] %02d/%d — %.2f USD\n",
		"Selecciona o dividendo a eliminar (1-%d): ": "Select the dividend to delete (1-%d): ",
		"⚠ Selección non válida":                     "⚠ Invalid selection",
	})
}
