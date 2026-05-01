package viewtypereport

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

type entry struct {
	asset domain.Asset
	sum   domain.MonthlySummary
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Informe mensual por tipo ---\n")

	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}

	var ofType []domain.Asset
	for _, a := range assets {
		if a.Type == typ {
			ofType = append(ofType, a)
		}
	}
	if len(ofType) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}

	fmt.Fprintf(w, "Activos de tipo %s:\n", typ.Display())
	for _, a := range ofType {
		fmt.Fprintf(w, "  - %s\n", a.Name)
	}

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	var active []entry
	for _, a := range ofType {
		sum, err := repo.MonthlySummary(a.ID, year, month)
		if err != nil {
			return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
		}
		if sum.EstimatedHolding > 0 {
			active = append(active, entry{asset: a, sum: sum})
		}
	}

	if len(active) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s con capital investido en %02d/%d.\n",
			typ.Display(), month, year)
		return nil
	}

	renderTable(w, typ, year, month, active)
	return nil
}

func renderTable(w io.Writer, typ domain.AssetType, year, month int, active []entry) {
	var totalInvested, investedInMonth, holding float64
	var resultSum, holdingForResult float64
	var withResult int
	for _, e := range active {
		totalInvested += e.sum.TotalInvestedUpTo
		investedInMonth += e.sum.InvestedInMonth
		holding += e.sum.EstimatedHolding
		if e.sum.HasResult {
			resultSum += e.sum.Result
			holdingForResult += e.sum.EstimatedHolding
			withResult++
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Tipo: %s · %02d/%d\n", typ.Display(), month, year)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Activos incluídos: %d\n", len(active))
	for _, e := range active {
		fmt.Fprintf(w, "    - %s (no activo: %.2f USD)\n", e.asset.Name, e.sum.EstimatedHolding)
	}
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Investido ata o mes\t%.2f USD\n", totalInvested)
	fmt.Fprintf(tw, "  Investido este mes\t%.2f USD\n", investedInMonth)
	fmt.Fprintf(tw, "  No activo\t%.2f USD\n", holding)

	switch {
	case withResult == 0:
		fmt.Fprintln(tw, "  Resultado\t— USD")
		fmt.Fprintln(tw, "  Gañanzas/Perdas\t— USD")
		fmt.Fprintln(tw, "  Índice\t—")
	case withResult == len(active):
		fmt.Fprintf(tw, "  Resultado\t%.2f USD\n", resultSum)
		gain := resultSum - holdingForResult
		fmt.Fprintf(tw, "  Gañanzas/Perdas\t%+.2f USD\n", gain)
		if holdingForResult > 0 {
			pct := gain / holdingForResult * 100
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", pct)
		} else {
			fmt.Fprintln(tw, "  Índice\tn/a")
		}
	default:
		fmt.Fprintf(tw, "  Resultado (parc.)\t%.2f USD  (%d/%d activos)\n", resultSum, withResult, len(active))
		gain := resultSum - holdingForResult
		fmt.Fprintf(tw, "  Gañanzas/Perdas (parc.)\t%+.2f USD\n", gain)
		if holdingForResult > 0 {
			pct := gain / holdingForResult * 100
			fmt.Fprintf(tw, "  Índice (parc.)\t%+.2f%%\n", pct)
		} else {
			fmt.Fprintln(tw, "  Índice (parc.)\tn/a")
		}
	}
	tw.Flush()

	fmt.Fprintln(w, sep)
}
