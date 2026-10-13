package deleteresult

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Eliminar resultado mensual ---\n": "\n--- Delete monthly result ---\n",
		"listando activos: %w":                   "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"listando resultados: %w":                              "listing results: %w",
		"Este activo non ten resultados mensuais rexistrados.": "This asset has no monthly results recorded.",
		"eliminando resultado: %w":                             "deleting result: %w",
		"✓ Resultado #%d eliminado (%02d/%d — %.2f USD)\n":     "✓ Result #%d deleted (%02d/%d — %.2f USD)\n",
		"Resultados de %s — %s:\n":                             "Results for %s — %s:\n",
		"  [%d] %02d/%d — %.2f USD\n":                          "  [%d] %02d/%d — %.2f USD\n",
		"Selecciona o resultado a eliminar (1-%d): ":           "Select the result to delete (1-%d): ",
		"⚠ Selección non válida":                               "⚠ Invalid selection",
	})
}
