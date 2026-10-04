package viewtotalreport

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
	holding           float64 // suma de EstimatedHolding dos activos con capital
	holdingWithResult float64 // suma de EstimatedHolding dos activos con resultado
	resultSum         float64 // suma de monthly_results para ese mes
	dividends         float64 // dividendos do mes
	dividendsPrev     float64 // dividendos do mes anterior
	assetsActive      int     // activos con EstimatedHolding > 0
	assetsWithResult  int     // activos con resultado rexistrado
	totalAssetsInPool int     // tamaño do pool de activos consultados
}

// monthMetrics recolle métricas derivadas dunha agregación. HasMetrics indica
// se hai datos suficientes para computar gañanzas (BaseNoDiv > 0 e algún
// activo con resultado).
type monthMetrics struct {
	HoldingNoDiv   float64 // capital estimado de tódolos activos con capital
	HoldingWithDiv float64
	BaseNoDiv      float64 // capital estimado dos activos con resultado: base da G/P
	BaseWithDiv    float64
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
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
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
			agg.holding += sum.EstimatedHolding
		}
		// EstimatedHolding parte do último resultado coñecido do activo (non
		// só do mes anterior), así que un mes sen resultado non deixa o
		// activo fóra da base. Resultado e base inclúen o mesmo conxunto.
		if sum.HasResult {
			agg.resultSum += sum.Result
			agg.holdingWithResult += sum.EstimatedHolding
			agg.assetsWithResult++
		}
	}

	div, err := repo.SumDividends(year, month)
	if err != nil {
		return agg, err
	}
	agg.dividends = div

	py, pm := prevMonth(year, month)
	divPrev, err := repo.SumDividends(py, pm)
	if err != nil {
		return agg, err
	}
	agg.dividendsPrev = divPrev

	return agg, nil
}

func computeMetrics(a monthAgg) monthMetrics {
	var m monthMetrics
	m.HoldingNoDiv = a.holding
	m.HoldingWithDiv = m.HoldingNoDiv + a.dividendsPrev
	m.BaseNoDiv = a.holdingWithResult
	m.BaseWithDiv = m.BaseNoDiv + a.dividendsPrev
	m.ResultNoDiv = a.resultSum
	m.ResultWithDiv = a.resultSum + a.dividends

	if m.BaseNoDiv <= 0 || a.assetsWithResult == 0 {
		return m
	}
	m.GainNoDiv = m.ResultNoDiv - m.BaseNoDiv
	m.PctNoDiv = m.GainNoDiv / m.BaseNoDiv * 100
	m.GainWithDiv = m.ResultWithDiv - m.BaseWithDiv
	if m.BaseWithDiv > 0 {
		m.PctWithDiv = m.GainWithDiv / m.BaseWithDiv * 100
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
	fmt.Fprintf(tw, "  Total investido ata o mes\t%s\n", money.USD(target.totalInvested))
	fmt.Fprintf(tw, "  Investido este mes\t%s\n", money.USD(target.investedInMonth))
	fmt.Fprintf(tw, "  Investimento + dividendos prev. mes\t%s\n",
		money.USD(target.investedInMonth+target.dividendsPrev))
	fmt.Fprintf(tw, "  No activo (sen div)\t%s\n", money.USD(cur.HoldingNoDiv))
	fmt.Fprintf(tw, "  No activo (con div)\t%s\n", money.USD(cur.HoldingWithDiv))
	fmt.Fprintf(tw, "  Dividendos este mes\t%s\n", money.USD(target.dividends))

	if target.assetsWithResult == 0 {
		fmt.Fprintln(tw, "  Resultado (sen div)\t—")
		fmt.Fprintln(tw, "  Resultado total (con div)\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas (con div)\t—")
		fmt.Fprintln(tw, "  Índice\t—")
		fmt.Fprintln(tw, "  Índice (con div)\t—")
	} else {
		coverage := ""
		if target.assetsWithResult < target.assetsActive {
			coverage = fmt.Sprintf("  (%d/%d activos con resultado)",
				target.assetsWithResult, target.assetsActive)
		}
		fmt.Fprintf(tw, "  Resultado (sen div)\t%s%s\n", money.USD(cur.ResultNoDiv), coverage)
		fmt.Fprintf(tw, "  Resultado total (con div)\t%s\n", money.USD(cur.ResultWithDiv))
		if cur.HasMetrics {
			fmt.Fprintf(tw, "  Gañanzas/Perdas\t%s\n", money.SignedUSD(cur.GainNoDiv))
			fmt.Fprintf(tw, "  Gañanzas/Perdas (con div)\t%s\n", money.SignedUSD(cur.GainWithDiv))
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", cur.PctNoDiv)
			if cur.BaseWithDiv > 0 {
				fmt.Fprintf(tw, "  Índice (con div)\t%+.2f%%\n", cur.PctWithDiv)
			} else {
				fmt.Fprintln(tw, "  Índice (con div)\tn/a")
			}
		} else {
			fmt.Fprintln(tw, "  Gañanzas/Perdas\tn/a")
			fmt.Fprintln(tw, "  Gañanzas/Perdas (con div)\tn/a")
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
		fmt.Fprintf(twAvg, "    Gañanza media mensual (sen div)\t%s\n",
			money.SignedUSD(sumGainNoDiv/float64(nAvgMonths)))
		fmt.Fprintf(twAvg, "    Índice medio mensual (con div)\t%+.2f%%\n",
			sumPctWithDiv/float64(nAvgMonths))
		fmt.Fprintf(twAvg, "    Gañanza media mensual (con div)\t%s\n",
			money.SignedUSD(sumGainWithDiv/float64(nAvgMonths)))
		twAvg.Flush()
	}

	fmt.Fprintln(w, sep)
}
