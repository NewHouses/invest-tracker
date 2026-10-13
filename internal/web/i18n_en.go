package web

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		// Consola do servidor.
		"Control de Investimentos — modo web":                         "Control de Investimentos — web mode",
		"  Local:      %s\n":                                          "  Local:      %s\n",
		"  Rede local: %s\n":                                          "  Network:    %s\n",
		"  Datos:      %s\n":                                          "  Data:       %s\n",
		"Preme Ctrl+C para parar.":                                    "Press Ctrl+C to stop.",
		"Control de Investimentos xa está aberto en %s.\n":            "Control de Investimentos is already open at %s.\n",
		"non se puido abrir o navegador: %v\n":                        "could not open the browser: %v\n",
		"non se pode abrir a base de datos: %v\n":                     "cannot open the database: %v\n",
		"non se pode iniciar o servidor: o porto %s xa está en uso\n": "cannot start the server: port %s is already in use\n",
		"non se pode iniciar o servidor: %v\n":                        "cannot start the server: %v\n",
		"erro no servidor web: %v\n":                                  "web server error: %v\n",
		"erro en %s %s: %v":                                           "error in %s %s: %v",
		"pánico atendendo %s %s: %v":                                  "panic while handling %s %s: %v",
		"Uso: invest-tracker web [opcións]":                           "Usage: invest-tracker web [options]",
		"Opcións:":                                                    "Options:",
		"enderezo no que escoitar":                                    "address to listen on",
		"ruta da base de datos SQLite":                                "path to the SQLite database",
		"non abrir o navegador automaticamente":                       "do not open the browser automatically",

		// Erros xerais da API.
		"erro interno":             "internal error",
		"recurso non atopado":      "resource not found",
		"método non permitido":     "method not allowed",
		"datos non válidos":        "invalid data",
		"identificador non válido": "invalid identifier",
		"enderezo non permitido: usa a IP do equipo ou localhost": "address not allowed: use the computer's IP or localhost",
		"o corpo JSON supera o tamaño máximo de 1 MiB":            "the JSON body exceeds the maximum size of 1 MiB",
		"o corpo JSON non pode estar baleiro":                     "the JSON body cannot be empty",
		"o corpo JSON debe conter un único valor":                 "the JSON body must contain a single value",
		"campo descoñecido %s":                                    "unknown field %s",
		"JSON non válido: %w":                                     "invalid JSON: %w",

		// Recursos non atopados.
		"activo non atopado":      "asset not found",
		"transacción non atopada": "transaction not found",
		"resultado non atopado":   "result not found",
		"dividendo non atopado":   "dividend not found",

		// Validacións.
		"o nome non pode estar baleiro":                                  "the name cannot be empty",
		"o nome non pode superar 100 caracteres":                         "the name cannot exceed 100 characters",
		"o importe debe ser un número maior ca 0":                        "the amount must be a number greater than 0",
		"o ano debe estar entre %d e %d":                                 "the year must be between %d and %d",
		"o mes debe estar entre 1 e 12":                                  "the month must be between 1 and 12",
		"o tipo de activo non é válido":                                  "the asset type is not valid",
		"tipo de activo non válido":                                      "invalid asset type",
		"tipo de activo duplicado":                                       "duplicate asset type",
		"o activo non existe":                                            "the asset does not exist",
		"o activo está repetido":                                         "the asset is repeated",
		"o activo creouse despois dese mes":                              "the asset was created after that month",
		"o activo non ten capital investido nese mes":                    "the asset has no invested capital that month",
		"a transacción non pode ser anterior á data do activo (%02d/%d)": "the transaction cannot be earlier than the asset's date (%02d/%d)",
		"a data non pode ser posterior á primeira transacción ou resultado do activo (%02d/%d)": "the date cannot be later than the asset's first transaction or result (%02d/%d)",
		"o tipo de transacción debe ser compra ou venda":                                        "the transaction type must be buy or sell",
		"o total debe ser unha cantidade positiva":                                              "the total must be a positive amount",
		"debe seleccionarse polo menos un tipo":                                                 "at least one type must be selected",
		"debe seleccionarse polo menos un activo":                                               "at least one asset must be selected",
		"os activos seleccionados deben existir e pertencer ao tipo indicado":                   "the selected assets must exist and belong to the given type",
		"indica a porcentaxe de cada activo":                                                    "enter the percentage for each asset",
		"a porcentaxe debe estar entre 0 e 100":                                                 "the percentage must be between 0 and 100",
		"as porcentaxes teñen que sumar 100 %":                                                  "the percentages must add up to 100%",
		"a data de inicio é obrigatoria":                                                        "the start date is required",
		"o modo debe ser contribution ou salary":                                                "the mode must be contribution or salary",
		"o tipo de regra debe ser fixed ou percent":                                             "the rule type must be fixed or percent",
		"o valor da regra non é válido":                                                         "the rule value is not valid",
		"a periodicidade debe estar entre 0 e %d meses":                                         "the frequency must be between 0 and %d months",
		"a data de inicio da regra non é válida":                                                "the rule's start date is not valid",
	})
}
