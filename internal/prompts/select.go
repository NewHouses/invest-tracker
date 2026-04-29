package prompts

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
)

func SelectAsset(r *bufio.Reader, w io.Writer, assets []domain.Asset) (domain.Asset, error) {
	fmt.Fprintln(w, "Investimentos:")
	for i, a := range assets {
		fmt.Fprintf(w, "  [%d] %s — %s\n", i+1, a.Type.Display(), a.Name)
	}
	for {
		fmt.Fprintf(w, "Selecciona (1-%d): ", len(assets))
		line, err := ReadLine(r)
		if err != nil {
			return domain.Asset{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(assets) {
			return assets[idx-1], nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida")
	}
}
