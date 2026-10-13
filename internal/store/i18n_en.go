package store

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"activo id=%d: %w":            "asset id=%d: %w",
		"dividendo id=%d: %w":         "dividend id=%d: %w",
		"resultado mensual id=%d: %w": "monthly result id=%d: %w",
		"transacción id=%d: %w":       "transaction id=%d: %w",
	})
}
