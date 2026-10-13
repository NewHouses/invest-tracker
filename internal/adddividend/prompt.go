package adddividend

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Saver interface {
	InsertDividend(domain.Dividend) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, saver Saver) error {
	i18n.Fprint(w, "\n--- Engadir dividendo mensual ---\n")

	amount, err := promptDividend(r, w)
	if err != nil {
		return err
	}
	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	d := domain.Dividend{
		AmountUSD: amount,
		Month:     month,
		Year:      year,
	}
	id, err := saver.InsertDividend(d)
	if err != nil {
		return i18n.Errorf("gardando dividendo: %w", err)
	}

	i18n.Fprintf(w, "✓ Dividendo gardado #%d: %.2f USD — %02d/%d\n",
		id, amount, month, year)
	return nil
}

func promptDividend(r *bufio.Reader, w io.Writer) (float64, error) {
	for {
		i18n.Fprint(w, "Dividendo (USD): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && domain.ValidAmount(v) {
			return v, nil
		}
		i18n.Fprintln(w, "⚠ Dividendo non válido, debe ser un número maior ca 0")
	}
}
