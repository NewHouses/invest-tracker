package addasset

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir novo activo ---\n":                      "\n--- Add new asset ---\n",
		"gardando activo: %w":                                  "saving asset: %w",
		"✓ Activo gardado #%d: %s — %s — %.2f USD — %02d/%d\n": "✓ Asset saved #%d: %s — %s — %.2f USD — %02d/%d\n",
		"Nome: ":                          "Name: ",
		"⚠ O nome non pode estar baleiro": "⚠ The name cannot be empty",
	})
}
