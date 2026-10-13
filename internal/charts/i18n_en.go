package charts

import "invest-tracker/internal/i18n"

func init() {
	i18n.Register(map[string]string{
		"%s · %02d/%d → %02d/%d (%d meses)": "%s · %02d/%d → %02d/%d (%d months)",
	})
}
