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
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	MonthsWithResultsForAsset(assetID int64) ([]domain.YearMonth, error)
}

const sep = "==================================================================="

type rowEntry struct {
	year, month int
	aporte      float64
	holding     float64
	result      float64
	gain        float64
	gainPct     float64
	hasMetrics  bool
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Reporte histórico dun activo ---\n")

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

	months, err := repo.MonthsWithResultsForAsset(chosen.ID)
	if err != nil {
		return fmt.Errorf("obtendo meses con resultados: %w", err)
	}
	if len(months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para este activo. Engade un coa operación 'Engadir resultado'.")
		return nil
	}

	rows := make([]rowEntry, 0, len(months))
	var sumPct, sumGain float64
	var nValid int

	for _, ym := range months {
		sum, err := repo.MonthlySummary(chosen.ID, ym.Year, ym.Month)
		if err != nil {
			return fmt.Errorf("calculando resumo: %w", err)
		}
		row := rowEntry{
			year:    ym.Year,
			month:   ym.Month,
			aporte:  sum.InvestedInMonth,
			holding: sum.EstimatedHolding,
			result:  sum.Result,
		}
		if sum.HasResult && sum.EstimatedHolding > 0 {
			row.gain = sum.Result - sum.EstimatedHolding
			row.gainPct = row.gain / sum.EstimatedHolding * 100
			row.hasMetrics = true
			sumPct += row.gainPct
			sumGain += row.gain
			nValid++
		}
		rows = append(rows, row)
	}

	last := rows[len(rows)-1]
	lifetimeSum, err := repo.MonthlySummary(chosen.ID, 9999, 12)
	if err != nil {
		return fmt.Errorf("calculando lifetime: %w", err)
	}
	lifetimeInvested := lifetimeSum.TotalInvestedUpTo
	lifetimeGain := last.result - lifetimeInvested
	hasLifetime := lifetimeInvested > 0 && last.result > 0

	renderReport(w, chosen, rows, lifetimeInvested, lifetimeGain, hasLifetime,
		nValid, sumPct, sumGain)
	return nil
}

func renderReport(w io.Writer, asset domain.Asset, rows []rowEntry,
	lifetimeInvested, lifetimeGain float64, hasLifetime bool,
	nValid int, sumPct, sumGain float64) {

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %s — %s · %d mes(es) con resultado\n",
		asset.Type.Display(), asset.Name, len(rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Total Aportado\t%.2f USD\n", lifetimeInvested)
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice Medio Mensual\t%+.2f%%\n", sumPct/float64(nValid))
		fmt.Fprintf(twH, "  Gañanzas/Perdas Medias Mensuais\t%+.2f USD\n", sumGain/float64(nValid))
	} else {
		fmt.Fprintln(twH, "  Índice Medio Mensual\t— %")
		fmt.Fprintln(twH, "  Gañanzas/Perdas Medias Mensuais\t— USD")
	}
	if hasLifetime {
		fmt.Fprintf(twH, "  Total Gañanzas/Perdas\t%+.2f USD\n", lifetimeGain)
	} else {
		fmt.Fprintln(twH, "  Total Gañanzas/Perdas\t— USD")
	}
	twH.Flush()
	fmt.Fprintln(w, sep)

	twT := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tNo activo\tÍndice\tG/P USD\tResultado\t")
	for _, row := range rows {
		var idxStr, gainStr string
		if row.hasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.gainPct)
			gainStr = fmt.Sprintf("%+.2f", row.gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		fmt.Fprintf(twT, "  %d\t%d\t%.2f\t%.2f\t%s\t%s\t%.2f\t\n",
			row.year, row.month, row.aporte, row.holding, idxStr, gainStr, row.result)
	}
	twT.Flush()
	fmt.Fprintln(w, sep)
}
