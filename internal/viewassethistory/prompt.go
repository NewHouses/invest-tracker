package viewassethistory

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
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	MonthsWithResultsForAsset(assetID int64) ([]domain.YearMonth, error)
}

const sep = "==================================================================="

type History struct {
	Asset         domain.Asset `json:"asset"`
	Rows          []Row        `json:"rows"`
	TotalInvested float64      `json:"totalInvested"`
	AvgIndexPct   float64      `json:"avgIndexPct"`
	AvgGain       float64      `json:"avgGain"`
	HasAverages   bool         `json:"hasAverages"`
	TotalGain     float64      `json:"totalGain"`
	HasTotalGain  bool         `json:"hasTotalGain"`
}

type Row struct {
	Period     domain.YearMonth `json:"period"`
	Aporte     float64          `json:"aporte"`
	Holding    float64          `json:"holding"`
	Result     float64          `json:"result"`
	Gain       float64          `json:"gain"`
	GainPct    float64          `json:"gainPct"`
	HasMetrics bool             `json:"hasMetrics"`
}

func Current(repo Repo, asset domain.Asset, now domain.YearMonth) (*Row, error) {
	if !asset.CreatedBy(now) {
		return nil, nil
	}
	months, err := repo.MonthsWithResultsForAsset(asset.ID)
	if err != nil {
		return nil, fmt.Errorf("obtendo meses con resultados: %w", err)
	}
	for _, ym := range months {
		if !ym.Before(now) {
			return nil, nil
		}
	}
	sum, err := repo.MonthlySummary(asset.ID, now.Year, now.Month)
	if err != nil {
		return nil, fmt.Errorf("calculando resumo: %w", err)
	}
	if sum.EstimatedHolding <= 0 && sum.InvestedInMonth == 0 {
		return nil, nil
	}
	return &Row{Period: now, Aporte: sum.InvestedInMonth, Holding: sum.EstimatedHolding}, nil
}

func Build(repo Repo, asset domain.Asset) (History, error) {
	months, err := repo.MonthsWithResultsForAsset(asset.ID)
	if err != nil {
		return History{}, fmt.Errorf("obtendo meses con resultados: %w", err)
	}
	if len(months) == 0 {
		return History{Asset: asset}, nil
	}

	history := History{
		Asset: asset,
		Rows:  make([]Row, 0, len(months)),
	}
	var sumPct, sumGain float64
	var nValid int

	for _, ym := range months {
		sum, err := repo.MonthlySummary(asset.ID, ym.Year, ym.Month)
		if err != nil {
			return History{}, fmt.Errorf("calculando resumo: %w", err)
		}
		row := Row{
			Period:  ym,
			Aporte:  sum.InvestedInMonth,
			Holding: sum.EstimatedHolding,
			Result:  sum.Result,
		}
		if sum.HasResult && sum.EstimatedHolding > 0 {
			row.Gain = sum.Result - sum.EstimatedHolding
			row.GainPct = row.Gain / sum.EstimatedHolding * 100
			row.HasMetrics = true
			sumPct += row.GainPct
			sumGain += row.Gain
			nValid++
		}
		history.Rows = append(history.Rows, row)
	}

	last := history.Rows[len(history.Rows)-1]
	lifetimeSum, err := repo.MonthlySummary(asset.ID, 9999, 12)
	if err != nil {
		return History{}, fmt.Errorf("calculando lifetime: %w", err)
	}
	history.TotalInvested = lifetimeSum.TotalInvestedUpTo
	history.TotalGain = last.Result - history.TotalInvested
	history.HasTotalGain = history.TotalInvested > 0 && last.Result > 0
	if nValid > 0 {
		history.AvgIndexPct = sumPct / float64(nValid)
		history.AvgGain = sumGain / float64(nValid)
		history.HasAverages = true
	}

	return history, nil
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

	history, err := Build(repo, chosen)
	if err != nil {
		return err
	}
	if len(history.Rows) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para este activo. Engade un coa operación 'Engadir resultado'.")
		return nil
	}

	renderReport(w, history)
	return nil
}

func renderReport(w io.Writer, history History) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %s — %s · %d mes(es) con resultado\n",
		history.Asset.Type.Display(), history.Asset.Name, len(history.Rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Total Aportado\t%s\n", money.USD(history.TotalInvested))
	if history.HasAverages {
		fmt.Fprintf(twH, "  Índice Medio Mensual\t%+.2f%%\n", history.AvgIndexPct)
		fmt.Fprintf(twH, "  Gañanzas/Perdas Medias Mensuais\t%s\n", money.SignedUSD(history.AvgGain))
	} else {
		fmt.Fprintln(twH, "  Índice Medio Mensual\t—")
		fmt.Fprintln(twH, "  Gañanzas/Perdas Medias Mensuais\t—")
	}
	if history.HasTotalGain {
		fmt.Fprintf(twH, "  Total Gañanzas/Perdas\t%s\n", money.SignedUSD(history.TotalGain))
	} else {
		fmt.Fprintln(twH, "  Total Gañanzas/Perdas\t—")
	}
	twH.Flush()
	fmt.Fprintln(w, sep)

	var tbuf bytes.Buffer
	twT := tabwriter.NewWriter(&tbuf, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tNo activo\tÍndice\tG/P\tResultado\t")
	for _, row := range history.Rows {
		var idxStr, gainStr string
		if row.HasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.GainPct)
			gainStr = money.SignedUSD(row.Gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		fmt.Fprintf(twT, "  %d\t%d\t%s\t%s\t%s\t%s\t%s\t\n",
			row.Period.Year, row.Period.Month,
			money.USD(row.Aporte), money.USD(row.Holding),
			idxStr, gainStr, money.USD(row.Result))
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
