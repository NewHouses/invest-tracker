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

type Averages struct {
	Months      int     `json:"months"`
	PctNoDiv    float64 `json:"pctNoDiv"`
	GainNoDiv   float64 `json:"gainNoDiv"`
	PctWithDiv  float64 `json:"pctWithDiv"`
	GainWithDiv float64 `json:"gainWithDiv"`
}

type Report struct {
	Period                    domain.YearMonth `json:"period"`
	TotalAssets               int              `json:"totalAssets"`
	AssetsActive              int              `json:"assetsActive"`
	AssetsWithResult          int              `json:"assetsWithResult"`
	Partial                   bool             `json:"partial"`
	TotalInvested             float64          `json:"totalInvested"`
	InvestedInMonth           float64          `json:"investedInMonth"`
	InvestedPlusPrevDividends float64          `json:"investedPlusPrevDividends"`
	Dividends                 float64          `json:"dividends"`
	DividendsPrev             float64          `json:"dividendsPrev"`
	HoldingNoDiv              float64          `json:"holdingNoDiv"`
	HoldingWithDiv            float64          `json:"holdingWithDiv"`
	BaseNoDiv                 float64          `json:"baseNoDiv"`
	BaseWithDiv               float64          `json:"baseWithDiv"`
	ResultNoDiv               float64          `json:"resultNoDiv"`
	ResultWithDiv             float64          `json:"resultWithDiv"`
	GainNoDiv                 float64          `json:"gainNoDiv"`
	GainWithDiv               float64          `json:"gainWithDiv"`
	PctNoDiv                  float64          `json:"pctNoDiv"`
	PctWithDiv                float64          `json:"pctWithDiv"`
	HasResults                bool             `json:"hasResults"`
	HasMetrics                bool             `json:"hasMetrics"`
	HasPctWithDiv             bool             `json:"hasPctWithDiv"`
	Averages                  Averages         `json:"averages"`
}

type monthAgg struct {
	year, month       int
	totalInvested     float64
	investedInMonth   float64
	holding           float64
	holdingWithResult float64
	resultSum         float64
	dividends         float64
	dividendsPrev     float64
	assetsActive      int
	assetsWithResult  int
	totalAssetsInPool int
}

type monthMetrics struct {
	HoldingNoDiv   float64
	HoldingWithDiv float64
	BaseNoDiv      float64
	BaseWithDiv    float64
	ResultNoDiv    float64
	ResultWithDiv  float64
	GainNoDiv      float64
	GainWithDiv    float64
	PctNoDiv       float64
	PctWithDiv     float64
	HasMetrics     bool
}

func Build(repo Repo, year, month int) (Report, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return Report{}, fmt.Errorf("listando activos: %w", err)
	}
	return buildFromAssets(repo, assets, year, month)
}

func buildFromAssets(repo Repo, assets []domain.Asset, year, month int) (Report, error) {
	target, err := aggregateMonth(repo, assets, year, month)
	if err != nil {
		return Report{}, err
	}
	report := reportFromAgg(target)
	if report.AssetsActive == 0 {
		return report, nil
	}

	monthsWithResults, err := repo.MonthsWithResultsUpTo(year, month)
	if err != nil {
		return Report{}, fmt.Errorf("obtendo meses con resultados: %w", err)
	}

	var sumPctNoDiv, sumGainNoDiv, sumPctWithDiv, sumGainWithDiv float64
	var nMonths int
	for _, ym := range monthsWithResults {
		agg, err := aggregateMonth(repo, assets, ym.Year, ym.Month)
		if err != nil {
			return Report{}, err
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
	if nMonths > 0 {
		report.Averages = Averages{
			Months:      nMonths,
			PctNoDiv:    sumPctNoDiv / float64(nMonths),
			GainNoDiv:   sumGainNoDiv / float64(nMonths),
			PctWithDiv:  sumPctWithDiv / float64(nMonths),
			GainWithDiv: sumGainWithDiv / float64(nMonths),
		}
	}
	return report, nil
}

func reportFromAgg(target monthAgg) Report {
	cur := computeMetrics(target)
	return Report{
		Period:                    domain.YearMonth{Year: target.year, Month: target.month},
		TotalAssets:               target.totalAssetsInPool,
		AssetsActive:              target.assetsActive,
		AssetsWithResult:          target.assetsWithResult,
		Partial:                   target.assetsWithResult > 0 && target.assetsWithResult < target.assetsActive,
		TotalInvested:             target.totalInvested,
		InvestedInMonth:           target.investedInMonth,
		InvestedPlusPrevDividends: target.investedInMonth + target.dividendsPrev,
		Dividends:                 target.dividends,
		DividendsPrev:             target.dividendsPrev,
		HoldingNoDiv:              cur.HoldingNoDiv,
		HoldingWithDiv:            cur.HoldingWithDiv,
		BaseNoDiv:                 cur.BaseNoDiv,
		BaseWithDiv:               cur.BaseWithDiv,
		ResultNoDiv:               cur.ResultNoDiv,
		ResultWithDiv:             cur.ResultWithDiv,
		GainNoDiv:                 cur.GainNoDiv,
		GainWithDiv:               cur.GainWithDiv,
		PctNoDiv:                  cur.PctNoDiv,
		PctWithDiv:                cur.PctWithDiv,
		HasResults:                target.assetsWithResult > 0,
		HasMetrics:                cur.HasMetrics,
		HasPctWithDiv:             cur.HasMetrics && cur.BaseWithDiv > 0,
	}
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

	report, err := buildFromAssets(repo, assets, year, month)
	if err != nil {
		return err
	}
	if report.AssetsActive == 0 {
		fmt.Fprintf(w, "Non hai activos con capital investido en %02d/%d.\n", month, year)
		return nil
	}

	renderTable(w, report)
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
		if domain.HasHolding(sum.EstimatedHolding) {
			agg.assetsActive++
			agg.holding += sum.EstimatedHolding
		}
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

func renderTable(w io.Writer, report Report) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Informe total · %02d/%d\n", report.Period.Month, report.Period.Year)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Activos: %d (%d activos en %02d/%d)\n",
		report.TotalAssets, report.AssetsActive, report.Period.Month, report.Period.Year)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Total investido ata o mes\t%s\n", money.USD(report.TotalInvested))
	fmt.Fprintf(tw, "  Investido este mes\t%s\n", money.USD(report.InvestedInMonth))
	fmt.Fprintf(tw, "  Investimento + dividendos prev. mes\t%s\n",
		money.USD(report.InvestedPlusPrevDividends))
	fmt.Fprintf(tw, "  No activo (sen div)\t%s\n", money.USD(report.HoldingNoDiv))
	fmt.Fprintf(tw, "  No activo (con div)\t%s\n", money.USD(report.HoldingWithDiv))
	fmt.Fprintf(tw, "  Dividendos este mes\t%s\n", money.USD(report.Dividends))

	if !report.HasResults {
		fmt.Fprintln(tw, "  Resultado (sen div)\t—")
		fmt.Fprintln(tw, "  Resultado total (con div)\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas\t—")
		fmt.Fprintln(tw, "  Gañanzas/Perdas (con div)\t—")
		fmt.Fprintln(tw, "  Índice\t—")
		fmt.Fprintln(tw, "  Índice (con div)\t—")
	} else {
		coverage := ""
		if report.Partial {
			coverage = fmt.Sprintf("  (%d/%d activos con resultado)",
				report.AssetsWithResult, report.AssetsActive)
		}
		fmt.Fprintf(tw, "  Resultado (sen div)\t%s%s\n", money.USD(report.ResultNoDiv), coverage)
		fmt.Fprintf(tw, "  Resultado total (con div)\t%s\n", money.USD(report.ResultWithDiv))
		if report.HasMetrics {
			fmt.Fprintf(tw, "  Gañanzas/Perdas\t%s\n", money.SignedUSD(report.GainNoDiv))
			fmt.Fprintf(tw, "  Gañanzas/Perdas (con div)\t%s\n", money.SignedUSD(report.GainWithDiv))
			fmt.Fprintf(tw, "  Índice\t%+.2f%%\n", report.PctNoDiv)
			if report.HasPctWithDiv {
				fmt.Fprintf(tw, "  Índice (con div)\t%+.2f%%\n", report.PctWithDiv)
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

	if report.Averages.Months == 0 {
		fmt.Fprintln(w, "  Promedios mensuais: sen meses con resultados.")
	} else {
		fmt.Fprintf(w, "  Promedios mensuais (%d mes(es) con resultado ata %02d/%d):\n",
			report.Averages.Months, report.Period.Month, report.Period.Year)
		twAvg := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(twAvg, "    Índice medio mensual (sen div)\t%+.2f%%\n",
			report.Averages.PctNoDiv)
		fmt.Fprintf(twAvg, "    Gañanza media mensual (sen div)\t%s\n",
			money.SignedUSD(report.Averages.GainNoDiv))
		fmt.Fprintf(twAvg, "    Índice medio mensual (con div)\t%+.2f%%\n",
			report.Averages.PctWithDiv)
		fmt.Fprintf(twAvg, "    Gañanza media mensual (con div)\t%s\n",
			money.SignedUSD(report.Averages.GainWithDiv))
		twAvg.Flush()
	}

	fmt.Fprintln(w, sep)
}
