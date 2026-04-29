package clearmonth

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/prompts"
)

type Repo interface {
	DeleteMonthlyResultsByMonth(year, month int) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Limpar mes ---\n")

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	n, err := repo.DeleteMonthlyResultsByMonth(year, month)
	if err != nil {
		return fmt.Errorf("limpando mes: %w", err)
	}

	if n == 0 {
		fmt.Fprintf(w, "Non había resultados rexistrados en %02d/%d.\n", month, year)
	} else {
		fmt.Fprintf(w, "✓ Limpado mes %02d/%d: %d resultado(s) eliminado(s).\n", month, year, n)
	}
	return nil
}
