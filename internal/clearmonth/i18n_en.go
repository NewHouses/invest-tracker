package clearmonth

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Limpar mes ---\n":                                 "\n--- Clear month ---\n",
		"limpando mes: %w":                                       "clearing month: %w",
		"Non había resultados rexistrados en %02d/%d.\n":         "There were no results recorded in %02d/%d.\n",
		"✓ Limpado mes %02d/%d: %d resultado(s) eliminado(s).\n": "✓ Cleared month %02d/%d: %d result(s) deleted.\n",
	})
}
