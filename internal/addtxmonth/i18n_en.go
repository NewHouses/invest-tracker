package addtxmonth

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"\n--- Engadir transaccións do mes ---\n": "\n--- Add the month's transactions ---\n",
		"listando activos: %w":                    "listing assets: %w",
		"Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.": "There are no assets yet. Add one first with the 'Add asset' operation.",
		"Non hai activos creados en %02d/%d ou antes.\n":                            "There are no assets created in %02d/%d or earlier.\n",
		"\nAporte por activo en %02d/%d (deixa en branco para saltar):\n":           "\nContribution by asset in %02d/%d (leave blank to skip):\n",
		"  %s — %s: omitido (creado en %02d/%d)\n":                                  "  %s — %s: skipped (created in %02d/%d)\n",
		"  %s — %s: ":                    "  %s — %s: ",
		"gardando transacción de %s: %w": "saving transaction for %s: %w",
		"    ✓ #%d %.2f USD\n":           "    ✓ #%d %.2f USD\n",
		"\n✓ %d transacción(s) engadidas (%.2f USD total), %d saltado(s).\n": "\n✓ %d transaction(s) added (%.2f USD total), %d skipped.\n",
		"  ⚠ Valor non válido (debe ser > 0; baleiro para saltar)":           "  ⚠ Invalid value (must be > 0; blank to skip)",
	})
}
