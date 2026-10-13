package deletedividend

import (
	"bufio"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListDividends() ([]domain.Dividend, error)
	DeleteDividend(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Eliminar dividendo mensual ---\n")

	dividends, err := repo.ListDividends()
	if err != nil {
		return i18n.Errorf("listando dividendos: %w", err)
	}
	if len(dividends) == 0 {
		i18n.Fprintln(w, "Aínda non hai dividendos rexistrados. Engade un primeiro coa operación 'Engadir dividendo'.")
		return nil
	}

	chosen, err := promptDividendSelection(r, w, dividends)
	if err != nil {
		return err
	}

	if err := repo.DeleteDividend(chosen.ID); err != nil {
		return i18n.Errorf("eliminando dividendo: %w", err)
	}

	i18n.Fprintf(w, "✓ Dividendo #%d eliminado (%02d/%d — %.2f USD)\n",
		chosen.ID, chosen.Month, chosen.Year, chosen.AmountUSD)
	return nil
}

func promptDividendSelection(r *bufio.Reader, w io.Writer, dividends []domain.Dividend) (domain.Dividend, error) {
	i18n.Fprintln(w, "Dividendos:")
	for i, d := range dividends {
		i18n.Fprintf(w, "  [%d] %02d/%d — %.2f USD\n", i+1, d.Month, d.Year, d.AmountUSD)
	}
	for {
		i18n.Fprintf(w, "Selecciona o dividendo a eliminar (1-%d): ", len(dividends))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.Dividend{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(dividends) {
			return dividends[idx-1], nil
		}
		i18n.Fprintln(w, "⚠ Selección non válida")
	}
}
