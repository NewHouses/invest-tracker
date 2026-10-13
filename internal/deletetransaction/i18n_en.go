package deletetransaction

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Eliminar transacción ---\n": "\n--- Delete transaction ---\n",
		"listando activos: %w":             "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"listando transaccións: %w": "listing transactions: %w",
		"Este activo non ten transaccións extras rexistradas (a compra inicial non é unha transacción).": "This asset has no extra transactions recorded (the initial buy is not a transaction).",
		"eliminando transacción: %w":                         "deleting transaction: %w",
		"✓ Transacción #%d eliminada (%02d/%d — %.2f USD)\n": "✓ Transaction #%d deleted (%02d/%d — %.2f USD)\n",
		"Transaccións de %s — %s:\n":                         "Transactions for %s — %s:\n",
		"  [%d] %02d/%d — %.2f USD\n":                        "  [%d] %02d/%d — %.2f USD\n",
		"Selecciona a transacción a eliminar (1-%d): ":       "Select the transaction to delete (1-%d): ",
		"⚠ Selección non válida":                             "⚠ Invalid selection",
	})
}
