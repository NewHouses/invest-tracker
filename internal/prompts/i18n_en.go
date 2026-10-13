package prompts

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"operación cancelada polo usuario":                      "operation cancelled by the user",
		"Cantidade (USD): ":                                     "Amount (USD): ",
		"⚠ Cantidade non válida, debe ser un número maior ca 0": "⚠ Invalid amount, must be a number greater than 0",
		"%s: ": "%s: ",
		"⚠ Cantidade non válida, debe ser un número maior ou igual a 0": "⚠ Invalid amount, must be a number greater than or equal to 0",
		"⚠ Porcentaxe non válida, debe ser un número maior ca %g%%\n":   "⚠ Invalid percentage, must be a number greater than %g%%\n",
		"Mes (1-12): ": "Month (1-12): ",
		"⚠ Mes non válido, debe estar entre 1 e 12": "⚠ Invalid month, must be between 1 and 12",
		"Ano: ":                          "Year: ",
		"⚠ Ano non válido":               "⚠ Invalid year",
		"Investimentos:":                 "Investments:",
		"  [%d] %s — %s\n":               "  [%d] %s — %s\n",
		"Selecciona (1-%d): ":            "Select (1-%d): ",
		"⚠ Selección non válida":         "⚠ Invalid selection",
		"Tipo de investimento:\n":        "Investment type:\n",
		"  [1] Acción\n":                 "  [1] Stock\n",
		"  [2] Índice\n":                 "  [2] Index\n",
		"  [3] Copy-trading\n":           "  [3] Copy trading\n",
		"  [4] Fondo\n":                  "  [4] Fund\n",
		"> ":                             "> ",
		"⚠ Tipo non válido, escolle 1-4": "⚠ Invalid type, choose 1-4",
		"Tipo de transacción:":           "Transaction type:",
		"  [1] Compra":                   "  [1] Buy",
		"  [2] Venda":                    "  [2] Sell",
		"⚠ Tipo non válido, escolle 1 (compra) ou 2 (venda)":                                 "⚠ Invalid type, choose 1 (buy) or 2 (sell)",
		"⚠ A transacción non pode ser anterior á data do activo (%02d/%d). Tenta de novo.\n": "⚠ The transaction cannot be earlier than the asset's date (%02d/%d). Try again.\n",
	})
}
