package deleteasset

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Eliminar activo ---\n": "\n--- Delete asset ---\n",
		"listando activos: %w":        "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"eliminando activo: %w": "deleting asset: %w",
		"✓ Activo #%d eliminado: %s — %s (e todas as súas transaccións e resultados mensuais)\n": "✓ Asset #%d deleted: %s — %s (and all its transactions and monthly results)\n",
	})
}
