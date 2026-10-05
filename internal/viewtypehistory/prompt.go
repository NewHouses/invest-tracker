package viewtypehistory

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
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
	Type          domain.AssetType `json:"type"`
	AssetCount    int              `json:"assetCount"`
	Rows          []Row            `json:"rows"`
	TotalInvested float64          `json:"totalInvested"`
	AvgIndexPct   float64          `json:"avgIndexPct"`
	AvgGain       float64          `json:"avgGain"`
	HasAverages   bool             `json:"hasAverages"`
	TotalGain     float64          `json:"totalGain"`
	HasTotalGain  bool             `json:"hasTotalGain"`
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

func Build(repo Repo, typ domain.AssetType) (History, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return History{}, fmt.Errorf("listando activos: %w", err)
	}
	var ofType []domain.Asset
	for _, a := range assets {
		if a.Type == typ {
			ofType = append(ofType, a)
		}
	}
	history := History{Type: typ, AssetCount: len(ofType)}
	if len(ofType) == 0 {
		return history, nil
	}

	monthsSet := make(map[domain.YearMonth]bool)
	monthsByAsset := make(map[int64][]domain.YearMonth, len(ofType))
	for _, a := range ofType {
		ms, err := repo.MonthsWithResultsForAsset(a.ID)
		if err != nil {
			return History{}, fmt.Errorf("obtendo meses de %s: %w", a.Name, err)
		}
		monthsByAsset[a.ID] = ms
		for _, ym := range ms {
			monthsSet[ym] = true
		}
	}
	if len(monthsSet) == 0 {
		return history, nil
	}
	months := make([]domain.YearMonth, 0, len(monthsSet))
	for ym := range monthsSet {
		months = append(months, ym)
	}
	sort.Slice(months, func(i, j int) bool {
		if months[i].Year != months[j].Year {
			return months[i].Year < months[j].Year
		}
		return months[i].Month < months[j].Month
	})

	history.Rows = make([]Row, 0, len(months))
	var sumPct, sumGain float64
	var nValid int

	for _, ym := range months {
		var aporte, holding, result float64
		for _, a := range ofType {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return History{}, fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if !sum.HasResult {
				continue
			}
			aporte += sum.InvestedInMonth
			holding += sum.EstimatedHolding
			result += sum.Result
		}
		row := Row{
			Period:  ym,
			Aporte:  aporte,
			Holding: holding,
			Result:  result,
		}
		if holding > 0 {
			row.Gain = result - holding
			row.GainPct = row.Gain / holding * 100
			row.HasMetrics = true
			sumPct += row.GainPct
			sumGain += row.Gain
			nValid++
		}
		history.Rows = append(history.Rows, row)
	}

	var lifetimeResult float64
	var hasAnyResult bool
	for _, a := range ofType {
		lifeSum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return History{}, fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		history.TotalInvested += lifeSum.TotalInvestedUpTo

		ms := monthsByAsset[a.ID]
		if len(ms) == 0 {
			continue
		}
		last := ms[len(ms)-1]
		lastSum, err := repo.MonthlySummary(a.ID, last.Year, last.Month)
		if err != nil {
			return History{}, fmt.Errorf("calculando último resumo de %s: %w", a.Name, err)
		}
		if lastSum.HasResult {
			lifetimeResult += lastSum.Result
			hasAnyResult = true
		}
	}
	history.TotalGain = lifetimeResult - history.TotalInvested
	history.HasTotalGain = history.TotalInvested > 0 && hasAnyResult
	if nValid > 0 {
		history.AvgIndexPct = sumPct / float64(nValid)
		history.AvgGain = sumGain / float64(nValid)
		history.HasAverages = true
	}

	return history, nil
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Reporte histórico dun tipo ---\n")

	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}

	history, err := Build(repo, typ)
	if err != nil {
		return err
	}
	if history.AssetCount == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}
	if len(history.Rows) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n",
			typ.Display())
		return nil
	}

	renderReport(w, history)
	return nil
}

func renderReport(w io.Writer, history History) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Tipo: %s · %d activo(s) · %d mes(es) con resultado\n",
		history.Type.Display(), history.AssetCount, len(history.Rows))
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
