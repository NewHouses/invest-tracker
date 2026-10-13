package addtransaction

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir nova transacción ---\n": "\n--- Add new transaction ---\n",
		"listando activos: %w":                 "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Sobre %s — %s\n":          "For %s — %s\n",
		"COMPRA":                   "BUY",
		"VENDA":                    "SELL",
		"gardando transacción: %w": "saving transaction: %w",
		"✓ Transacción gardada #%d sobre %s — %s: %s %.2f USD — %02d/%d\n": "✓ Transaction saved #%d for %s — %s: %s %.2f USD — %02d/%d\n",
	})
}
