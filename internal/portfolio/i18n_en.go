package portfolio

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"listando activos: %w":           "listing assets: %w",
		"construíndo carteira de %s: %w": "building portfolio for %s: %w",
	})
}
