package viewassethistory

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
	AssetReport(assetID int64) (domain.AssetReport, error)
}

const sep = "========================================================="

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Historial dun activo ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa opción 1.")
		return nil
	}

	chosen, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}

	report, err := repo.AssetReport(chosen.ID)
	if err != nil {
		return fmt.Errorf("calculando historial: %w", err)
	}

	if len(report.Rows) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para este activo. Engade un coa opción 3.")
		return nil
	}

	renderReport(w, chosen, report)
	return nil
}

func renderReport(w io.Writer, a domain.Asset, r domain.AssetReport) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %s — %s\n", a.Type.Display(), a.Name)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Total investido\t%.2f USD\n", r.TotalInvested)
	if r.HasTotalGain {
		fmt.Fprintf(tw, "  Total Ganhanzas/Perdas\t%+.2f USD\n", r.TotalGain)
	} else {
		fmt.Fprintln(tw, "  Total Ganhanzas/Perdas\t— USD")
	}
	if r.HasAvgIndex {
		fmt.Fprintf(tw, "  Índice medio mensual\t%+.2f %%\n", r.AvgMonthlyIndexPct)
	} else {
		fmt.Fprintln(tw, "  Índice medio mensual\t— %")
	}
	tw.Flush()

	fmt.Fprintln(w, sep)

	tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "  Ano\tMes\tInvestido ata o mes\tInvestido este mes\tÍndice\tG/P USD\tResultado\t")
	for _, row := range r.Rows {
		var idx string
		if row.HasGainPct {
			idx = fmt.Sprintf("%+.2f%%", row.GainPct)
		} else {
			idx = "n/a"
		}
		fmt.Fprintf(tw, "  %d\t%d\t%.2f\t%.2f\t%s\t%+.2f\t%.2f\t\n",
			row.Year, row.Month, row.TotalInvestedUpTo, row.InvestedInMonth,
			idx, row.Gain, row.Result)
	}
	tw.Flush()

	fmt.Fprintln(w, sep)
}
