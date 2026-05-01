package viewassetgeneral

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

const sep = "==================================================================================="

type rowEntry struct {
	year, month       int
	totalInvestedUpTo float64
	investedInMonth   float64
	holding           float64
	result            float64
	gain              float64
	gainPct           float64
	hasMetrics        bool
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Resultado xeral dun activo ---\n")

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
			year:              ym.Year,
			month:             ym.Month,
			totalInvestedUpTo: sum.TotalInvestedUpTo,
			investedInMonth:   sum.InvestedInMonth,
			holding:           sum.EstimatedHolding,
			result:            sum.Result,
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

	// Lifetime totals: usamos a última fila (resultado máis recente).
	lastRow := rows[len(rows)-1]
	// Usamos un sumario "lifetime" para o invested total (mes futuro lonxano).
	lifetimeSum, err := repo.MonthlySummary(chosen.ID, 9999, 12)
	if err != nil {
		return fmt.Errorf("calculando lifetime: %w", err)
	}
	lifetimeInvested := lifetimeSum.TotalInvestedUpTo
	lifetimeGain := lastRow.result - lifetimeInvested
	hasLifetime := lifetimeInvested > 0 && lastRow.result > 0
	var lifetimePct float64
	if hasLifetime {
		lifetimePct = lifetimeGain / lifetimeInvested * 100
	}

	renderTable(w, chosen, rows, lifetimeInvested, lifetimeGain, lifetimePct, hasLifetime,
		nValid, sumPct, sumGain)
	return nil
}

func renderTable(w io.Writer, asset domain.Asset, rows []rowEntry,
	lifetimeInvested, lifetimeGain, lifetimePct float64, hasLifetime bool,
	nValid int, sumPct, sumGain float64) {

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Resultado xeral · %s — %s · %d mes(es) con resultado\n",
		asset.Type.Display(), asset.Name, len(rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Total investido\t%.2f USD\n", lifetimeInvested)
	if hasLifetime {
		fmt.Fprintf(twH, "  Total Ganhanzas/Perdas\t%+.2f USD\n", lifetimeGain)
		fmt.Fprintf(twH, "  Total Índice\t%+.2f%%\n", lifetimePct)
	} else {
		fmt.Fprintln(twH, "  Total Ganhanzas/Perdas\t— USD")
		fmt.Fprintln(twH, "  Total Índice\t— %")
	}
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice medio mensual\t%+.2f%%\n", sumPct/float64(nValid))
		fmt.Fprintf(twH, "  Ganhanza media mensual\t%+.2f USD\n", sumGain/float64(nValid))
	} else {
		fmt.Fprintln(twH, "  Índice medio mensual\t— %")
		fmt.Fprintln(twH, "  Ganhanza media mensual\t— USD")
	}
	twH.Flush()
	fmt.Fprintln(w, sep)

	// Táboa A: investimentos por mes
	fmt.Fprintln(w, "  Investimentos por mes:")
	twA := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twA, "  Ano\tMes\tInvestido total\tEste mes\tNo activo\t")
	for _, row := range rows {
		fmt.Fprintf(twA, "  %d\t%d\t%.2f\t%.2f\t%.2f\t\n",
			row.year, row.month,
			row.totalInvestedUpTo, row.investedInMonth, row.holding,
		)
	}
	twA.Flush()
	fmt.Fprintln(w, sep)

	// Táboa B: resultados e ganhanzas por mes
	fmt.Fprintln(w, "  Resultados e ganhanzas por mes:")
	twB := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twB, "  Ano\tMes\tResultado\tG/P USD\t%\t")
	for _, row := range rows {
		var gainStr, pctStr string
		if row.hasMetrics {
			gainStr = fmt.Sprintf("%+.2f", row.gain)
			pctStr = fmt.Sprintf("%+.2f%%", row.gainPct)
		} else {
			gainStr = "—"
			pctStr = "—"
		}
		fmt.Fprintf(twB, "  %d\t%d\t%.2f\t%s\t%s\t\n",
			row.year, row.month, row.result, gainStr, pctStr,
		)
	}
	twB.Flush()
	fmt.Fprintln(w, sep)
}
