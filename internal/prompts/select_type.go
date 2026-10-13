package prompts

import (
	"bufio"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
)

func SelectAssetType(r *bufio.Reader, w io.Writer) (domain.AssetType, error) {
	i18n.Fprint(w, "Tipo de investimento:\n")
	i18n.Fprint(w, "  [1] Acción\n")
	i18n.Fprint(w, "  [2] Índice\n")
	i18n.Fprint(w, "  [3] Copy-trading\n")
	i18n.Fprint(w, "  [4] Fondo\n")
	for {
		i18n.Fprint(w, "> ")
		line, err := ReadLine(r)
		if err != nil {
			return "", err
		}
		switch line {
		case "1":
			return domain.Accion, nil
		case "2":
			return domain.Indice, nil
		case "3":
			return domain.CopyTrading, nil
		case "4":
			return domain.Fondo, nil
		}
		i18n.Fprintln(w, "⚠ Tipo non válido, escolle 1-4")
	}
}
