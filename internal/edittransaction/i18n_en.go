package edittransaction

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Editar unha transacción ---\n": "\n--- Edit a transaction ---\n",
		"listando activos: %w":                "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"listando transaccións: %w":                            "listing transactions: %w",
		"Este activo non ten transaccións extras rexistradas.": "This asset has no extra transactions recorded.",
		"Transacción seleccionada: %s\n":                       "Selected transaction: %s\n",
		"actualizando transacción: %w":                         "updating transaction: %w",
		"✓ Transacción #%d actualizada: %s\n":                  "✓ Transaction #%d updated: %s\n",
		"Transaccións de %s — %s:\n":                           "Transactions for %s — %s:\n",
		"  [%d] %s\n":                                          "  [%d] %s\n",
		"Selecciona a transacción a editar (1-%d): ":           "Select the transaction to edit (1-%d): ",
		"⚠ Selección non válida":                               "⚠ Invalid selection",
		"Que campo queres editar?":                             "Which field do you want to edit?",
		"  [1] Cantidade (USD)":                                "  [1] Amount (USD)",
		"  [2] Tipo (Compra/Venda)":                            "  [2] Type (Buy/Sell)",
		"  [3] Data (mes/ano)":                                 "  [3] Date (month/year)",
		"> ":                                                   "> ",
		"⚠ Selección non válida, escolle 1-3":                  "⚠ Invalid selection, choose 1-3",
		"COMPRA":                "BUY",
		"VENDA":                 "SELL",
		"%02d/%d — %s %.2f USD": "%02d/%d — %s %.2f USD",
	})
}
