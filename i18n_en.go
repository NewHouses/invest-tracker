package main

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"↷ Operación cancelada. Volvendo ao menú.": "↷ Operation cancelled. Back to the menu.",
		"⚠ erro %s: %s\n":                          "⚠ error %s: %s\n",
		"⚠ non se puido gardar o idioma: %v\n":     "⚠ could not save the language: %v\n",
		"✓ Idioma: galego":                         "✓ Language: English",
		"erro lendo entrada: %v\n":                 "error reading input: %v\n",
		"non se pode abrir a base de datos: %v":    "cannot open the database: %v",
		"(Datos: %s)":                              "(Data: %s)",
		"Ata logo!":                                "Goodbye!",

		"engadindo activo":                   "adding asset",
		"editando activo":                    "editing asset",
		"eliminando activo":                  "deleting asset",
		"engadindo transacción":              "adding transaction",
		"engadindo varias transaccións":      "adding several transactions",
		"engadindo transaccións do mes":      "adding the month's transactions",
		"repartindo aporte mensual":          "splitting monthly contribution",
		"editando transacción":               "editing transaction",
		"eliminando transacción":             "deleting transaction",
		"engadindo resultado":                "adding result",
		"engadindo dividendo":                "adding dividend",
		"pechando o mes":                     "closing the month",
		"eliminando resultado":               "deleting result",
		"eliminando dividendo":               "deleting dividend",
		"limpando o mes":                     "clearing the month",
		"xerando historial":                  "building history",
		"xerando historial por tipo":         "building history by type",
		"listando transaccións":              "listing transactions",
		"xerando informe":                    "building report",
		"xerando informe por tipo":           "building report by type",
		"xerando informe total":              "building total report",
		"xerando reporte histórico completo": "building full history report",
		"debuxando gráficas":                 "drawing charts",
		"calculando a proxección":            "calculating the projection",
	})
}
