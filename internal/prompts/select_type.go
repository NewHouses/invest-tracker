package prompts

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/domain"
)

func SelectAssetType(r *bufio.Reader, w io.Writer) (domain.AssetType, error) {
	fmt.Fprint(w, "Tipo de investimento:\n")
	fmt.Fprint(w, "  [1] Acción\n")
	fmt.Fprint(w, "  [2] Índice\n")
	fmt.Fprint(w, "  [3] Copy-trading\n")
	fmt.Fprint(w, "  [4] Fondo\n")
	for {
		fmt.Fprint(w, "> ")
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
		fmt.Fprintln(w, "⚠ Tipo non válido, escolle 1-4")
	}
}
