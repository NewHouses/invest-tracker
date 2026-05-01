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

const sep = "==================================================================================================="

type monthAgg struct {
	year, month       int
	totalInvested     float64
	investedInMonth   float64
	resultSum         float64
	resultSumPrev     float64
	dividends         float64
	dividendsPrev     float64
	assetsWithResult  int
	assetsActive      int
}

type monthMetrics struct {
	HoldingNoDiv   float64
	HoldingWithDiv float64
	ResultNoDiv    float64
	ResultWithDiv  float64
	GainNoDiv      float64
	GainWithDiv    float64
	PctNoDiv       float64
	PctWithDiv     float64
	HasMetrics     bool
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Resultado xeral (historial total) ---\n")

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

	rows := make([]struct {
		agg     monthAgg
		metrics monthMetrics
	}, 0, len(months))

	var sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv float64
	var nValid int
	var lifetimeDividends float64

	for _, ym := range months {
		agg, err := aggregateMonth(repo, assets, ym.Year, ym.Month)
		if err != nil {
			return err
		}
		m := computeMetrics(agg)
		rows = append(rows, struct {
			agg     monthAgg
			metrics monthMetrics
		}{agg, m})
		lifetimeDividends += agg.dividends
		if m.HasMetrics {
			sumPctNoDiv += m.PctNoDiv
			sumGainNoDiv += m.GainNoDiv
			sumPctWithDiv += m.PctWithDiv
			sumGainWithDiv += m.GainWithDiv
			nValid++
		}
	}

	// Lifetime totals: invested + último resultado.
	lifetimeInvested := totalLifetimeInvested(repo, assets)
	lastIdx := len(rows) - 1
	lifetimeResult := rows[lastIdx].agg.resultSum
	lifetimeGain := lifetimeResult - lifetimeInvested
	var lifetimePct float64
	hasLifetimeMetrics := lifetimeInvested > 0 && rows[lastIdx].agg.assetsWithResult > 0
	if hasLifetimeMetrics {
		lifetimePct = lifetimeGain / lifetimeInvested * 100
	}

	renderHistory(w, rows, lifetimeInvested, lifetimeGain, lifetimePct, hasLifetimeMetrics,
		nValid, sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv)
	return nil
}

func totalLifetimeInvested(repo Repo, assets []domain.Asset) float64 {
	// Suma do TotalInvestedUpTo "lifetime" — aplicado a un ano/mes futuro.
	var total float64
	for _, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			continue
		}
		total += sum.TotalInvestedUpTo
	}
	return total
}

func prevMonth(y, m int) (int, int) {
	if m == 1 {
		return y - 1, 12
	}
	return y, m - 1
}

func aggregateMonth(repo Repo, assets []domain.Asset, year, month int) (monthAgg, error) {
	agg := monthAgg{year: year, month: month}
	for _, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, year, month)
		if err != nil {
			return agg, fmt.Errorf("calculando %s: %w", a.Name, err)
		}
		agg.totalInvested += sum.TotalInvestedUpTo
		agg.investedInMonth += sum.InvestedInMonth
		if sum.EstimatedHolding > 0 {
			agg.assetsActive++
		}
		if sum.HasResult {
			agg.resultSum += sum.Result
			agg.assetsWithResult++
		}
	}
	py, pm := prevMonth(year, month)
	for _, a := range assets {
		prevSum, err := repo.MonthlySummary(a.ID, py, pm)
		if err != nil {
			return agg, fmt.Errorf("previo de %s: %w", a.Name, err)
		}
		if prevSum.HasResult {
			agg.resultSumPrev += prevSum.Result
		}
	}
	div, err := repo.SumDividends(year, month)
	if err != nil {
		return agg, err
	}
	agg.dividends = div
	divPrev, err := repo.SumDividends(py, pm)
	if err != nil {
		return agg, err
	}
	agg.dividendsPrev = divPrev
	return agg, nil
}

func computeMetrics(a monthAgg) monthMetrics {
	var m monthMetrics
	m.HoldingNoDiv = a.resultSumPrev + a.investedInMonth
	m.HoldingWithDiv = m.HoldingNoDiv + a.dividendsPrev
	m.ResultNoDiv = a.resultSum
	m.ResultWithDiv = a.resultSum + a.dividends

	if m.HoldingNoDiv <= 0 || a.assetsWithResult == 0 {
		return m
	}
	m.GainNoDiv = m.ResultNoDiv - m.HoldingNoDiv
	m.PctNoDiv = m.GainNoDiv / m.HoldingNoDiv * 100
	m.GainWithDiv = m.ResultWithDiv - m.HoldingWithDiv
	if m.HoldingWithDiv > 0 {
		m.PctWithDiv = m.GainWithDiv / m.HoldingWithDiv * 100
	}
	m.HasMetrics = true
	return m
}

func renderHistory(w io.Writer,
	rows []struct {
		agg     monthAgg
		metrics monthMetrics
	},
	lifetimeInvested, lifetimeGain, lifetimePct float64,
	hasLifetimeMetrics bool,
	nValid int,
	sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv float64,
) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Resultado xeral · %d mes(es) con resultado\n", len(rows))
	fmt.Fprintln(w, sep)

	// Cabeceira: totais e medias
	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Total investido\t%.2f USD\n", lifetimeInvested)
	if hasLifetimeMetrics {
		fmt.Fprintf(twH, "  Total Gañanzas/Perdas\t%+.2f USD\n", lifetimeGain)
		fmt.Fprintf(twH, "  Total Índice\t%+.2f%%\n", lifetimePct)
	} else {
		fmt.Fprintln(twH, "  Total Gañanzas/Perdas\t— USD")
		fmt.Fprintln(twH, "  Total Índice\t— %")
	}
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice medio mensual (sen div)\t%+.2f%%\n", sumPctNoDiv/float64(nValid))
		fmt.Fprintf(twH, "  Gañanza media mensual (sen div)\t%+.2f USD\n", sumGainNoDiv/float64(nValid))
		fmt.Fprintf(twH, "  Índice medio mensual (con div)\t%+.2f%%\n", sumPctWithDiv/float64(nValid))
		fmt.Fprintf(twH, "  Gañanza media mensual (con div)\t%+.2f USD\n", sumGainWithDiv/float64(nValid))
	} else {
		fmt.Fprintln(twH, "  Índice medio mensual (sen div)\t— %")
		fmt.Fprintln(twH, "  Gañanza media mensual (sen div)\t— USD")
		fmt.Fprintln(twH, "  Índice medio mensual (con div)\t— %")
		fmt.Fprintln(twH, "  Gañanza media mensual (con div)\t— USD")
	}
	twH.Flush()
	fmt.Fprintln(w, sep)

	// Táboa A: investimentos por mes
	fmt.Fprintln(w, "  Investimentos por mes:")
	twA := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twA, "  Ano\tMes\tInvestido total\tEste mes\t+Div prev\tNo activo s/d\tNo activo c/d\t")
	for _, row := range rows {
		fmt.Fprintf(twA, "  %d\t%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t\n",
			row.agg.year, row.agg.month,
			row.agg.totalInvested,
			row.agg.investedInMonth,
			row.agg.investedInMonth+row.agg.dividendsPrev,
			row.metrics.HoldingNoDiv,
			row.metrics.HoldingWithDiv,
		)
	}
	twA.Flush()
	fmt.Fprintln(w, sep)

	// Táboa B: resultados, gañanzas e índices por mes
	fmt.Fprintln(w, "  Resultados e gañanzas por mes:")
	twB := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twB, "  Ano\tMes\tDiv\tResult s/d\tResult c/d\tG/P s/d\tG/P c/d\t% s/d\t% c/d\t")
	for _, row := range rows {
		var gainNoDivStr, gainWithDivStr, pctNoDivStr, pctWithDivStr string
		if row.metrics.HasMetrics {
			gainNoDivStr = fmt.Sprintf("%+.2f", row.metrics.GainNoDiv)
			gainWithDivStr = fmt.Sprintf("%+.2f", row.metrics.GainWithDiv)
			pctNoDivStr = fmt.Sprintf("%+.2f%%", row.metrics.PctNoDiv)
			if row.metrics.HoldingWithDiv > 0 {
				pctWithDivStr = fmt.Sprintf("%+.2f%%", row.metrics.PctWithDiv)
			} else {
				pctWithDivStr = "n/a"
			}
		} else {
			gainNoDivStr = "—"
			gainWithDivStr = "—"
			pctNoDivStr = "—"
			pctWithDivStr = "—"
		}
		fmt.Fprintf(twB, "  %d\t%d\t%.2f\t%.2f\t%.2f\t%s\t%s\t%s\t%s\t\n",
			row.agg.year, row.agg.month,
			row.agg.dividends,
			row.metrics.ResultNoDiv,
			row.metrics.ResultWithDiv,
			gainNoDivStr, gainWithDivStr,
			pctNoDivStr, pctWithDivStr,
		)
	}
	twB.Flush()
	fmt.Fprintln(w, sep)
}
