package viewtransactions

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":      "listing assets: %w",
		"listando transaccións: %w": "listing transactions: %w",

		"\n--- Transaccións dun activo ---\n":                                       "\n--- Asset transactions ---\n",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"  Transaccións de %s — %s\n":                                               "  Transactions for %s — %s\n",
		"  Total: %d entradas (incluíndo a compra inicial)\n":                       "  Total: %d entries (including the initial buy)\n",
		"  Compras: %s · Vendas: %s · Neto: %s\n":                                   "  Buys: %s · Sells: %s · Net: %s\n",
		"  ID\tAno\tMes\tCompra/Venda\tCantidade\t":                                 "  ID\tYear\tMonth\tBuy/Sell\tAmount\t",
		"COMPRA": "BUY",
		"VENDA":  "SELL",
	})
}
