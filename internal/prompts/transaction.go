package prompts

import (
	"bufio"
	"fmt"
	"io"
)

// SelectTransactionType pide ao usuario que escolla entre Compra (1) ou Venda (2).
// Devolve isVenda = true se é venda, false se é compra. Re-pregunta ata válido.
func SelectTransactionType(r *bufio.Reader, w io.Writer) (bool, error) {
	fmt.Fprintln(w, "Tipo de transacción:")
	fmt.Fprintln(w, "  [1] Compra")
	fmt.Fprintln(w, "  [2] Venda")
	for {
		fmt.Fprint(w, "> ")
		line, err := ReadLine(r)
		if err != nil {
			return false, err
		}
		switch line {
		case "1":
			return false, nil
		case "2":
			return true, nil
		}
		fmt.Fprintln(w, "⚠ Tipo non válido, escolle 1 (compra) ou 2 (venda)")
	}
}

// DateNotBefore pide mes e ano e valida que (year*12+month) >= (minYear*12+minMonth).
// Re-pregunta ambos ante data anterior.
func DateNotBefore(r *bufio.Reader, w io.Writer, minMonth, minYear int) (int, int, error) {
	for {
		month, err := Month(r, w)
		if err != nil {
			return 0, 0, err
		}
		year, err := Year(r, w)
		if err != nil {
			return 0, 0, err
		}
		if year*12+month >= minYear*12+minMonth {
			return month, year, nil
		}
		fmt.Fprintf(w, "⚠ A transacción non pode ser anterior á data do activo (%02d/%d). Tenta de novo.\n",
			minMonth, minYear)
	}
}
