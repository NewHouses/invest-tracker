package welcome

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"Ferramenta CLI para o seguimento mensual":                            "Command-line tool for monthly tracking",
		"Benvida! Escolle unha opción:":                                       "Welcome! Choose an option:",
		"(escribe ':q' ou 'cancelar' en calquera prompt para voltar ao menú)": "(type ':q' or 'cancel' at any prompt to go back to the menu)",
		"opción non válida":                                                   "invalid option",

		"Operacións con activos":         "Asset operations",
		"Engadir activo":                 "Add asset",
		"Editar activo":                  "Edit asset",
		"Eliminar activo":                "Delete asset",
		"Operacións con transaccións":    "Transaction operations",
		"Engadir transacción":            "Add transaction",
		"Engadir transaccións en serie":  "Add a series of transactions",
		"Engadir transaccións do mes":    "Add the month's transactions",
		"Repartir aporte mensual":        "Split monthly contribution",
		"Editar transacción":             "Edit transaction",
		"Eliminar transacción":           "Delete transaction",
		"Operacións de resultados":       "Result operations",
		"Engadir resultado":              "Add result",
		"Engadir dividendo":              "Add dividend",
		"Pechar mes":                     "Close month",
		"Eliminar resultado":             "Delete result",
		"Eliminar dividendo":             "Delete dividend",
		"Limpar mes":                     "Clear month",
		"Informes":                       "Reports",
		"Reporte histórico dun activo":   "Asset history report",
		"Reporte histórico dun tipo":     "Asset type history report",
		"Ver transaccións dun activo":    "View an asset's transactions",
		"Ver informe mensual dun activo": "View an asset's monthly report",
		"Ver informe mensual por tipo":   "View monthly report by type",
		"Ver informe mensual total":      "View total monthly report",
		"Reporte histórico completo":     "Full history report",
		"Ver gráficas":                   "View charts",
		"Proxección a 20 anos":           "20-year projection",
		"Saír":                           "Exit",
		"Voltar":                         "Back",
	})
}
