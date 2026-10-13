package editasset

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Editar activo ---\n": "\n--- Edit asset ---\n",
		"listando activos: %w":      "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.":                                       "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Activo seleccionado: %s — %s (%.2f USD, %02d/%d)\n":                                                              "Selected asset: %s — %s (%.2f USD, %02d/%d)\n",
		"comprobando os movementos do activo: %w":                                                                         "checking the asset movements: %w",
		"⚠ A data non pode ser posterior á primeira transacción ou resultado do activo (%02d/%d). Non se cambiou nada.\n": "⚠ The date cannot be later than the asset's first transaction or result (%02d/%d). Nothing was changed.\n",
		"actualizando activo: %w":                                 "updating asset: %w",
		"✓ Activo #%d actualizado: %s — %s (%.2f USD, %02d/%d)\n": "✓ Asset #%d updated: %s — %s (%.2f USD, %02d/%d)\n",
		"Que campo queres editar?":                                "Which field do you want to edit?",
		"  [1] Nome":                                              "  [1] Name",
		"  [2] Data (mes/ano)":                                    "  [2] Date (month/year)",
		"  [3] Cantidade (USD)":                                   "  [3] Amount (USD)",
		"> ":                                                      "> ",
		"⚠ Selección non válida, escolle 1-3":                     "⚠ Invalid selection, choose 1-3",
		"Novo nome: ":                                             "New name: ",
		"⚠ O nome non pode estar baleiro":                         "⚠ The name cannot be empty",
	})
}
