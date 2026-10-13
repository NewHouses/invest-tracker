package addtxbatch

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir varias transaccións ---\n": "\n--- Add several transactions ---\n",
		"listando activos: %w":                    "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Sobre %s — %s (creado en %02d/%d)\n":                                       "For %s — %s (created in %02d/%d)\n",
		"Introduce a data de inicio:":                                               "Enter the start date:",
		"\n--- Transacción #%d ---\n":                                               "\n--- Transaction #%d ---\n",
		"\n--- Transacción #%d en %02d/%d ---\n":                                    "\n--- Transaction #%d in %02d/%d ---\n",
		"COMPRA":                   "BUY",
		"VENDA":                    "SELL",
		"gardando transacción: %w": "saving transaction: %w",
		"✓ Gardada #%d: %s %.2f USD — %02d/%d\n":             "✓ Saved #%d: %s %.2f USD — %02d/%d\n",
		"\n✓ Engadíronse %d transacción(s) sobre %s — %s.\n": "\n✓ Added %d transaction(s) for %s — %s.\n",
		"Modo de entrada:": "Entry mode:",
		"  [1] Establecer mes/ano en cada transacción":              "  [1] Set month/year in each transaction",
		"  [2] Engadir todas no mesmo mes":                          "  [2] Add all in the same month",
		"  [3] Unha por mes (auto-incrementa) dende un mes inicial": "  [3] One per month (auto-increment) from a start month",
		"> ":                                    "> ",
		"⚠ Modo non válido, escolle 1-3":        "⚠ Invalid mode, choose 1-3",
		"Engadir outra transacción? [S/n]: ":    "Add another transaction? [Y/n]: ",
		"⚠ Resposta non válida, escolle S ou N": "⚠ Invalid response, choose Y or N",
	})
}
