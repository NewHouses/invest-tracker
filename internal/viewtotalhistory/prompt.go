package viewtotalhistory

import (
	"bufio"
	"fmt"
	"io"
	"text/tabwriter"

	"invest-tracker/internal/domain"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	SumDividends(year, month int) (float64, error)
	MonthsWithResults() ([]domain.YearMonth, error)
}

const sep = "==================================================================="

type rowEntry struct {
	year, month int
	aporte      float64
	fondos      float64
	dividends   float64
	result      float64 // sum of asset results + dividendos do mes
	gain        float64
	gainPct     float64
	hasMetrics  bool
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Reporte histórico completo ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	months, err := repo.MonthsWithResults()
	if err != nil {
		return fmt.Errorf("obtendo meses con resultados: %w", err)
	}
	if len(months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados. Engade resultados mensuais coa operación 'Engadir resultado' ou 'Pechar mes'.")
		return nil
	}

	rows := make([]rowEntry, 0, len(months))
	var sumPct, sumGain, totalDiv float64
	var nValid int

	for _, ym := range months {
		var aporte, fondos, baseResult float64
		for _, a := range assets {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			// Só agregamos os activos que reportan resultado neste mes:
			// así Fondos e Resultado inclúen sempre o mesmo conxunto.
			if !sum.HasResult {
				continue
			}
			aporte += sum.InvestedInMonth
			fondos += sum.EstimatedHolding
			baseResult += sum.Result
		}
		div, err := repo.SumDividends(ym.Year, ym.Month)
		if err != nil {
			return fmt.Errorf("sumando dividendos de %d/%d: %w", ym.Month, ym.Year, err)
		}
		result := baseResult + div
		row := rowEntry{
			year:      ym.Year,
			month:     ym.Month,
			aporte:    aporte,
			fondos:    fondos,
			dividends: div,
			result:    result,
		}
		if fondos > 0 {
			row.gain = result - fondos
			row.gainPct = row.gain / fondos * 100
			row.hasMetrics = true
			sumPct += row.gainPct
			sumGain += row.gain
			nValid++
		}
		totalDiv += div
		rows = append(rows, row)
	}

	// Aporte lifetime: suma de TotalInvestedUpTo a 9999/12 por activo.
	var lifetimeAporte float64
	for _, a := range assets {
		lifeSum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		lifetimeAporte += lifeSum.TotalInvestedUpTo
	}

	// G/P Total: valor actual da carteira (último resultado coñecido por activo
	// + dividendos acumulados) − aporte total.
	var lifetimeLastResult float64
	var hasAnyResult bool
	for _, a := range assets {
		for i := len(months) - 1; i >= 0; i-- {
			sum, err := repo.MonthlySummary(a.ID, months[i].Year, months[i].Month)
			if err != nil {
				return fmt.Errorf("buscando último resultado de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				lifetimeLastResult += sum.Result
				hasAnyResult = true
				break
			}
		}
	}
	lifetimeGain := lifetimeLastResult + totalDiv - lifetimeAporte
	hasLifetime := lifetimeAporte > 0 && hasAnyResult

	renderReport(w, len(assets), rows, lifetimeAporte, lifetimeGain, totalDiv,
		hasLifetime, nValid, sumPct, sumGain)
	return nil
}

func renderReport(w io.Writer, nAssets int, rows []rowEntry,
	lifetimeAporte, lifetimeGain, totalDiv float64, hasLifetime bool,
	nValid int, sumPct, sumGain float64) {

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Reporte histórico completo · %d activo(s) · %d mes(es) con resultado\n",
		nAssets, len(rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Aporte histórico total\t%.2f USD\n", lifetimeAporte)
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice Medio\t%+.2f%%\n", sumPct/float64(nValid))
		fmt.Fprintf(twH, "  G/P Media\t%+.2f USD\n", sumGain/float64(nValid))
	} else {
		fmt.Fprintln(twH, "  Índice Medio\t— %")
		fmt.Fprintln(twH, "  G/P Media\t— USD")
	}
	if hasLifetime {
		fmt.Fprintf(twH, "  G/P Total\t%+.2f USD\n", lifetimeGain)
	} else {
		fmt.Fprintln(twH, "  G/P Total\t— USD")
	}
	fmt.Fprintf(twH, "  Dividendos totais\t%.2f USD\n", totalDiv)
	twH.Flush()
	fmt.Fprintln(w, sep)

	twT := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tFondos\tÍndice\tG/P USD\tDividendos\tResultado\t")
	for _, row := range rows {
		var idxStr, gainStr string
		if row.hasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.gainPct)
			gainStr = fmt.Sprintf("%+.2f", row.gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		fmt.Fprintf(twT, "  %d\t%d\t%.2f\t%.2f\t%s\t%s\t%.2f\t%.2f\t\n",
			row.year, row.month, row.aporte, row.fondos, idxStr, gainStr,
			row.dividends, row.result)
	}
	twT.Flush()
	fmt.Fprintln(w, sep)
}
