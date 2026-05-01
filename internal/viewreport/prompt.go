package viewreport

import (
	"bufio"
	"fmt"
	"io"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
}

const sep = "========================================================="

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Informe mensual ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	chosen, err := prompts.SelectAsset(r, w, assets)
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

	summary, err := repo.MonthlySummary(chosen.ID, year, month)
	if err != nil {
		return fmt.Errorf("calculando informe: %w", err)
	}

	renderTable(w, chosen, year, month, summary)
	return nil
}

func renderTable(w io.Writer, a domain.Asset, year, month int, s domain.MonthlySummary) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %s — %s · %02d/%d\n", a.Type.Display(), a.Name, month, year)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Investido ata o mes\t%.2f USD\n", s.TotalInvestedUpTo)
	fmt.Fprintf(tw, "  Investido este mes\t%.2f USD\n", s.InvestedInMonth)
	fmt.Fprintf(tw, "  No activo\t%.2f USD\n", s.EstimatedHolding)
	if s.HasResult {
		fmt.Fprintf(tw, "  Resultado\t%.2f USD\n", s.Result)
		gain := s.Result - s.EstimatedHolding
		fmt.Fprintf(tw, "  Ganhanzas/Perdas\t%+.2f USD\n", gain)
		if s.EstimatedHolding > 0 {
			pct := gain / s.EstimatedHolding * 100
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", pct)
		} else {
			fmt.Fprintln(tw, "  Índice\tn/a")
		}
	} else {
		fmt.Fprintln(tw, "  Resultado\t— USD")
		fmt.Fprintln(tw, "  Ganhanzas/Perdas\t— USD")
		fmt.Fprintln(tw, "  Índice\t—")
	}
	tw.Flush()

	fmt.Fprintln(w, sep)
}
