package viewtotalreport

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
	SumDividends(year, month int) (float64, error)
	MonthsWithResultsUpTo(year, month int) ([]domain.YearMonth, error)
}

const sep = "========================================================="

// monthAgg agrega valores en bruto para un (year, month) sobre todos os
// activos pasados.
type monthAgg struct {
	year, month       int
	totalInvested     float64 // suma de TotalInvestedUpTo
	investedInMonth   float64 // suma de InvestedInMonth
	resultSum         float64 // suma de monthly_results para ese mes
	resultSumPrev     float64 // suma de monthly_results para o mes anterior
	dividends         float64 // dividendos do mes
	dividendsPrev     float64 // dividendos do mes anterior
	assetsActive      int     // activos con EstimatedHolding > 0
	assetsWithResult  int     // activos con resultado rexistrado
	totalAssetsInPool int     // tamaño do pool de activos consultados
}

// monthMetrics recolle métricas derivadas dunha agregación. HasMetrics indica
// se hai datos suficientes para computar gañanzas (HoldingNoDiv > 0 e algún
// activo con resultado).
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
	fmt.Fprint(w, "\n--- Informe mensual total ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa opción 1.")
		return nil
	}

	fmt.Fprintln(w, "Activos:")
	for _, a := range assets {
		fmt.Fprintf(w, "  - %s — %s\n", a.Type.Display(), a.Name)
	}

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	target, err := aggregateMonth(repo, assets, year, month)
	if err != nil {
		return err
	}
	if target.assetsActive == 0 {
		fmt.Fprintf(w, "Non hai activos con capital investido en %02d/%d.\n", month, year)
		return nil
	}

	// Promedios sobre meses con resultados ata target (incluído).
	monthsWithResults, err := repo.MonthsWithResultsUpTo(year, month)
	if err != nil {
		return fmt.Errorf("obtendo meses con resultados: %w", err)
	}

	var sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv float64
	var nMonths int
	for _, ym := range monthsWithResults {
		agg, err := aggregateMonth(repo, assets, ym.Year, ym.Month)
		if err != nil {
			return err
		}
		m := computeMetrics(agg)
		if !m.HasMetrics {
			continue
		}
		sumPctNoDiv += m.PctNoDiv
		sumGainNoDiv += m.GainNoDiv
		sumPctWithDiv += m.PctWithDiv
		sumGainWithDiv += m.GainWithDiv
		nMonths++
	}

	renderTable(w, target, nMonths, sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv)
	return nil
}

func prevMonth(y, m int) (int, int) {
	if m == 1 {
		return y - 1, 12
	}
	return y, m - 1
}

func aggregateMonth(repo Repo, assets []domain.Asset, year, month int) (monthAgg, error) {
	agg := monthAgg{year: year, month: month, totalAssetsInPool: len(assets)}

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
			return agg, fmt.Errorf("calculando previo de %s: %w", a.Name, err)
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

func renderTable(w io.Writer, target monthAgg, nAvgMonths int,
	sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv float64) {

	cur := computeMetrics(target)

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Informe total · %02d/%d\n", target.month, target.year)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Activos: %d (%d activos en %02d/%d)\n",
		target.totalAssetsInPool, target.assetsActive, target.month, target.year)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Total investido ata o mes\t%.2f USD\n", target.totalInvested)
	fmt.Fprintf(tw, "  Investido este mes\t%.2f USD\n", target.investedInMonth)
	fmt.Fprintf(tw, "  Investimento + dividendos prev. mes\t%.2f USD\n",
		target.investedInMonth+target.dividendsPrev)
	fmt.Fprintf(tw, "  No activo (sen div)\t%.2f USD\n", cur.HoldingNoDiv)
	fmt.Fprintf(tw, "  No activo (con div)\t%.2f USD\n", cur.HoldingWithDiv)
	fmt.Fprintf(tw, "  Dividendos este mes\t%.2f USD\n", target.dividends)

	if target.assetsWithResult == 0 {
		fmt.Fprintln(tw, "  Resultado (sen div)\t— USD")
		fmt.Fprintln(tw, "  Resultado total (con div)\t— USD")
		fmt.Fprintln(tw, "  Ganhanzas/Perdas\t— USD")
		fmt.Fprintln(tw, "  Ganhanzas/Perdas (con div)\t— USD")
		fmt.Fprintln(tw, "  Índice\t— %")
		fmt.Fprintln(tw, "  Índice (con div)\t— %")
	} else {
		coverage := ""
		if target.assetsWithResult < target.assetsActive {
			coverage = fmt.Sprintf("  (%d/%d activos con resultado)",
				target.assetsWithResult, target.assetsActive)
		}
		fmt.Fprintf(tw, "  Resultado (sen div)\t%.2f USD%s\n", cur.ResultNoDiv, coverage)
		fmt.Fprintf(tw, "  Resultado total (con div)\t%.2f USD\n", cur.ResultWithDiv)
		if cur.HasMetrics {
			fmt.Fprintf(tw, "  Ganhanzas/Perdas\t%+.2f USD\n", cur.GainNoDiv)
			fmt.Fprintf(tw, "  Ganhanzas/Perdas (con div)\t%+.2f USD\n", cur.GainWithDiv)
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", cur.PctNoDiv)
			if cur.HoldingWithDiv > 0 {
				fmt.Fprintf(tw, "  Índice (con div)\t%+.2f%%\n", cur.PctWithDiv)
			} else {
				fmt.Fprintln(tw, "  Índice (con div)\tn/a")
			}
		} else {
			fmt.Fprintln(tw, "  Ganhanzas/Perdas\tn/a")
			fmt.Fprintln(tw, "  Ganhanzas/Perdas (con div)\tn/a")
			fmt.Fprintln(tw, "  Índice\tn/a")
			fmt.Fprintln(tw, "  Índice (con div)\tn/a")
		}
	}
	tw.Flush()

	fmt.Fprintln(w, sep)

	// Promedios
	if nAvgMonths == 0 {
		fmt.Fprintln(w, "  Promedios mensuais: sen meses con resultados.")
	} else {
		fmt.Fprintf(w, "  Promedios mensuais (%d mes(es) con resultado ata %02d/%d):\n",
			nAvgMonths, target.month, target.year)
		twAvg := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(twAvg, "    Índice medio mensual (sen div)\t%+.2f%%\n",
			sumPctNoDiv/float64(nAvgMonths))
		fmt.Fprintf(twAvg, "    Ganhanza media mensual (sen div)\t%+.2f USD\n",
			sumGainNoDiv/float64(nAvgMonths))
		fmt.Fprintf(twAvg, "    Índice medio mensual (con div)\t%+.2f%%\n",
			sumPctWithDiv/float64(nAvgMonths))
		fmt.Fprintf(twAvg, "    Ganhanza media mensual (con div)\t%+.2f USD\n",
			sumGainWithDiv/float64(nAvgMonths))
		twAvg.Flush()
	}

	fmt.Fprintln(w, sep)
}
