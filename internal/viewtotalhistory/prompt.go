package viewtotalhistory

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"invest-tracker/internal/colors"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	SumDividends(year, month int) (float64, error)
	MonthsWithResults() ([]domain.YearMonth, error)
}

const sep = "==================================================================="

type History struct {
	AssetCount      int     `json:"assetCount"`
	Rows            []Row   `json:"rows"`
	LifetimeAporte  float64 `json:"lifetimeAporte"`
	AvgIndexPct     float64 `json:"avgIndexPct"`
	AvgGain         float64 `json:"avgGain"`
	HasAverages     bool    `json:"hasAverages"`
	TotalGain       float64 `json:"totalGain"`
	HasTotalGain    bool    `json:"hasTotalGain"`
	TotalDividends  float64 `json:"totalDividends"`
	CurrentValue    float64 `json:"currentValue"`
	HasCurrentValue bool    `json:"hasCurrentValue"`
}

type Row struct {
	Period     domain.YearMonth `json:"period"`
	Aporte     float64          `json:"aporte"`
	Fondos     float64          `json:"fondos"`
	Dividends  float64          `json:"dividends"`
	Result     float64          `json:"result"`
	Gain       float64          `json:"gain"`
	GainPct    float64          `json:"gainPct"`
	HasMetrics bool             `json:"hasMetrics"`
}

func Build(repo Repo) (History, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return History{}, fmt.Errorf("listando activos: %w", err)
	}
	history := History{AssetCount: len(assets)}
	if len(assets) == 0 {
		return history, nil
	}

	months, err := repo.MonthsWithResults()
	if err != nil {
		return History{}, fmt.Errorf("obtendo meses con resultados: %w", err)
	}
	if len(months) == 0 {
		return history, nil
	}

	history.Rows = make([]Row, 0, len(months))
	var sumPct, sumGain float64
	var nValid int

	for _, ym := range months {
		var totalTx, fondos, baseResult float64
		for _, a := range assets {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return History{}, fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			totalTx += sum.InvestedInMonth
			if !sum.HasResult {
				continue
			}
			fondos += sum.EstimatedHolding
			baseResult += sum.Result
		}
		div, err := repo.SumDividends(ym.Year, ym.Month)
		if err != nil {
			return History{}, fmt.Errorf("sumando dividendos de %d/%d: %w", ym.Month, ym.Year, err)
		}
		aporte := totalTx - div
		result := baseResult + div
		row := Row{
			Period:    ym,
			Aporte:    aporte,
			Fondos:    fondos,
			Dividends: div,
			Result:    result,
		}
		if fondos > 0 {
			row.Gain = result - fondos
			row.GainPct = row.Gain / fondos * 100
			row.HasMetrics = true
			sumPct += row.GainPct
			sumGain += row.Gain
			nValid++
		}
		history.TotalDividends += div
		history.Rows = append(history.Rows, row)
	}

	for _, a := range assets {
		lifeSum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return History{}, fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		history.LifetimeAporte += lifeSum.TotalInvestedUpTo
	}

	var lifetimeLastResult float64
	var hasAnyResult bool
	for _, a := range assets {
		for i := len(months) - 1; i >= 0; i-- {
			sum, err := repo.MonthlySummary(a.ID, months[i].Year, months[i].Month)
			if err != nil {
				return History{}, fmt.Errorf("buscando último resultado de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				lifetimeLastResult += sum.Result
				hasAnyResult = true
				break
			}
		}
	}
	history.CurrentValue = lifetimeLastResult + history.TotalDividends
	history.HasCurrentValue = hasAnyResult
	history.TotalGain = history.CurrentValue - history.LifetimeAporte
	history.HasTotalGain = history.LifetimeAporte > 0 && hasAnyResult
	if nValid > 0 {
		history.AvgIndexPct = sumPct / float64(nValid)
		history.AvgGain = sumGain / float64(nValid)
		history.HasAverages = true
	}

	return history, nil
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Reporte histórico completo ---\n")

	history, err := Build(repo)
	if err != nil {
		return err
	}
	if history.AssetCount == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}
	if len(history.Rows) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados. Engade resultados mensuais coa operación 'Engadir resultado' ou 'Pechar mes'.")
		return nil
	}

	renderReport(w, history)
	return nil
}

func renderReport(w io.Writer, history History) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Reporte histórico completo · %d activo(s) · %d mes(es) con resultado\n",
		history.AssetCount, len(history.Rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Aporte histórico total\t%s\n", money.USD(history.LifetimeAporte))
	if history.HasAverages {
		fmt.Fprintf(twH, "  Índice Medio\t%+.2f%%\n", history.AvgIndexPct)
		fmt.Fprintf(twH, "  G/P Media\t%s\n", money.SignedUSD(history.AvgGain))
	} else {
		fmt.Fprintln(twH, "  Índice Medio\t—")
		fmt.Fprintln(twH, "  G/P Media\t—")
	}
	if history.HasTotalGain {
		fmt.Fprintf(twH, "  G/P Total\t%s\n", money.SignedUSD(history.TotalGain))
	} else {
		fmt.Fprintln(twH, "  G/P Total\t—")
	}
	fmt.Fprintf(twH, "  Dividendos totais\t%s\n", money.USD(history.TotalDividends))
	twH.Flush()
	fmt.Fprintln(w, sep)

	var tbuf bytes.Buffer
	twT := tabwriter.NewWriter(&tbuf, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tFondos\tÍndice\tG/P\tDividendos\tResultado\t")
	for _, row := range history.Rows {
		var idxStr, gainStr string
		if row.HasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.GainPct)
			gainStr = money.SignedUSD(row.Gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		fmt.Fprintf(twT, "  %d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			row.Period.Year, row.Period.Month,
			money.SignedUSD(row.Aporte), money.USD(row.Fondos),
			idxStr, gainStr,
			money.USD(row.Dividends), money.USD(row.Result))
	}
	twT.Flush()
	writeColoredRows(w, tbuf.String(), history.Rows)
	fmt.Fprintln(w, sep)
}

// writeColoredRows imprime as liñas xa formatadas: a primeira (cabeceira) sen
// cor, e cada fila de datos envolvida no código ANSI segundo o seu G/P.
func writeColoredRows(w io.Writer, formatted string, rows []Row) {
	lines := strings.Split(strings.TrimRight(formatted, "\n"), "\n")
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(w, lines[0])
	for i, line := range lines[1:] {
		if i < len(rows) && rows[i].HasMetrics {
			fmt.Fprintln(w, colors.ForGain(rows[i].Gain)+line+colors.Reset)
		} else {
			fmt.Fprintln(w, line)
		}
	}
}
