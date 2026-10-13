package prompts

import (
	"bufio"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
)

func SelectAsset(r *bufio.Reader, w io.Writer, assets []domain.Asset) (domain.Asset, error) {
	i18n.Fprintln(w, "Investimentos:")
	for i, a := range assets {
		i18n.Fprintf(w, "  [%d] %s — %s\n", i+1, a.Type.Display(), a.Name)
	}
	for {
		i18n.Fprintf(w, "Selecciona (1-%d): ", len(assets))
		line, err := ReadLine(r)
		if err != nil {
			return domain.Asset{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(assets) {
			return assets[idx-1], nil
		}
		i18n.Fprintln(w, "⚠ Selección non válida")
	}
}
