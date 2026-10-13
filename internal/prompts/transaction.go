package prompts

import (
	"bufio"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
)

// SelectTransactionType pide ao usuario que escolla entre Compra (1) ou Venda (2).
// Devolve isVenda = true se é venda, false se é compra. Re-pregunta ata válido.
func SelectTransactionType(r *bufio.Reader, w io.Writer) (bool, error) {
	i18n.Fprintln(w, "Tipo de transacción:")
	i18n.Fprintln(w, "  [1] Compra")
	i18n.Fprintln(w, "  [2] Venda")
	for {
		i18n.Fprint(w, "> ")
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
		i18n.Fprintln(w, "⚠ Tipo non válido, escolle 1 (compra) ou 2 (venda)")
	}
}

// DateNotBefore pide mes e ano e valida que a data non sexa anterior á mínima.
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
		date := domain.YearMonth{Year: year, Month: month}
		minDate := domain.YearMonth{Year: minYear, Month: minMonth}
		if !date.Before(minDate) {
			return month, year, nil
		}
		i18n.Fprintf(w, "⚠ A transacción non pode ser anterior á data do activo (%02d/%d). Tenta de novo.\n",
			minMonth, minYear)
	}
}
