package deletedividend

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListDividends() ([]domain.Dividend, error)
	DeleteDividend(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Eliminar dividendo mensual ---\n")

	dividends, err := repo.ListDividends()
	if err != nil {
		return fmt.Errorf("listando dividendos: %w", err)
	}
	if len(dividends) == 0 {
		fmt.Fprintln(w, "Aínda non hai dividendos rexistrados. Engade un primeiro coa operación 'Engadir dividendo'.")
		return nil
	}

	chosen, err := promptDividendSelection(r, w, dividends)
	if err != nil {
		return err
	}

	if err := repo.DeleteDividend(chosen.ID); err != nil {
		return fmt.Errorf("eliminando dividendo: %w", err)
	}

	fmt.Fprintf(w, "✓ Dividendo #%d eliminado (%02d/%d — %.2f USD)\n",
		chosen.ID, chosen.Month, chosen.Year, chosen.AmountUSD)
	return nil
}

func promptDividendSelection(r *bufio.Reader, w io.Writer, dividends []domain.Dividend) (domain.Dividend, error) {
	fmt.Fprintln(w, "Dividendos:")
	for i, d := range dividends {
		fmt.Fprintf(w, "  [%d] %02d/%d — %.2f USD\n", i+1, d.Month, d.Year, d.AmountUSD)
	}
	for {
		fmt.Fprintf(w, "Selecciona o dividendo a eliminar (1-%d): ", len(dividends))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.Dividend{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(dividends) {
			return dividends[idx-1], nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida")
	}
}
