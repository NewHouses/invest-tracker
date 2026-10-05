package viewreport

import (
	"bufio"
	"fmt"
	"io"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
}

const sep = "========================================================="

type Report struct {
	Asset      domain.Asset          `json:"asset"`
	Period     domain.YearMonth      `json:"period"`
	Summary    domain.MonthlySummary `json:"summary"`
	Gain       float64               `json:"gain"`
	GainPct    float64               `json:"gainPct"`
	HasResult  bool                  `json:"hasResult"`
	HasGainPct bool                  `json:"hasGainPct"`
}

func Build(repo Repo, asset domain.Asset, year, month int) (Report, error) {
	summary, err := repo.MonthlySummary(asset.ID, year, month)
	if err != nil {
		return Report{}, fmt.Errorf("calculando informe: %w", err)
	}

	report := Report{
		Asset:      asset,
		Period:     domain.YearMonth{Year: year, Month: month},
		Summary:    summary,
		HasResult:  summary.HasResult,
		HasGainPct: summary.HasResult && summary.EstimatedHolding > 0,
	}
	if report.HasResult {
		report.Gain = summary.Result - summary.EstimatedHolding
		if report.HasGainPct {
			report.GainPct = report.Gain / summary.EstimatedHolding * 100
		}
	}
	return report, nil
}

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

	report, err := Build(repo, chosen, year, month)
	if err != nil {
		return err
	}

	renderTable(w, report)
	return nil
}

func renderTable(w io.Writer, report Report) {
	a := report.Asset
	s := report.Summary
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %s — %s · %02d/%d\n", a.Type.Display(), a.Name, report.Period.Month, report.Period.Year)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Investido ata o mes\t%s\n", money.USD(s.TotalInvestedUpTo))
	fmt.Fprintf(tw, "  Investido este mes\t%s\n", money.USD(s.InvestedInMonth))
	fmt.Fprintf(tw, "  No activo\t%s\n", money.USD(s.EstimatedHolding))
	if report.HasResult {
		fmt.Fprintf(tw, "  Resultado\t%s\n", money.USD(s.Result))
		fmt.Fprintf(tw, "  Gañanzas/Perdas\t%s\n", money.SignedUSD(report.Gain))
		if report.HasGainPct {
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", report.GainPct)
		} else {
			fmt.Fprintln(tw, "  Índice\tn/a")
		}
	} else {
		fmt.Fprintln(tw, "  Resultado\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas\t—")
		fmt.Fprintln(tw, "  Índice\t—")
	}
	tw.Flush()

	fmt.Fprintln(w, sep)
}
