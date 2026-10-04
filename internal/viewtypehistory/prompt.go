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
	fmt.Fprint(w, "\n--- Reporte histórico dun tipo ---\n")

	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	var ofType []domain.Asset
	for _, a := range assets {
		if a.Type == typ {
			ofType = append(ofType, a)
		}
	}
	if len(ofType) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}

	monthsSet := make(map[domain.YearMonth]bool)
	for _, a := range ofType {
		ms, err := repo.MonthsWithResultsForAsset(a.ID)
		if err != nil {
			return fmt.Errorf("obtendo meses de %s: %w", a.Name, err)
		}
		for _, ym := range ms {
			monthsSet[ym] = true
		}
	}
	if len(monthsSet) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n",
			typ.Display())
		return nil
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

	rows := make([]rowEntry, 0, len(months))
	var sumPct, sumGain float64
	var nValid int

	for _, ym := range months {
		var aporte, holding, result float64
		for _, a := range ofType {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			// Só incluímos os activos que reportan resultado neste mes,
			// para que holding e result inclúan o mesmo conxunto e a métrica
			// G/P sexa coherente.
			if !sum.HasResult {
				continue
			}
			aporte += sum.InvestedInMonth
			holding += sum.EstimatedHolding
			result += sum.Result
		}
		row := rowEntry{
			year:    ym.Year,
			month:   ym.Month,
			aporte:  aporte,
			holding: holding,
			result:  result,
		}
		if holding > 0 {
			row.gain = result - holding
			row.gainPct = row.gain / holding * 100
			row.hasMetrics = true
			sumPct += row.gainPct
			sumGain += row.gain
			nValid++
		}
		rows = append(rows, row)
	}

	// Totais lifetime: sumamos por activo o seu invested total e o seu último
	// resultado coñecido. Isto reflicte o "valor actual" da carteira do tipo.
	var lifetimeInvested, lifetimeResult float64
	var hasAnyResult bool
	for _, a := range ofType {
		lifeSum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		lifetimeInvested += lifeSum.TotalInvestedUpTo

		ms, err := repo.MonthsWithResultsForAsset(a.ID)
		if err != nil {
			return fmt.Errorf("obtendo meses de %s: %w", a.Name, err)
		}
		if len(ms) == 0 {
			continue
		}
		last := ms[len(ms)-1]
		lastSum, err := repo.MonthlySummary(a.ID, last.Year, last.Month)
		if err != nil {
			return fmt.Errorf("calculando último resumo de %s: %w", a.Name, err)
		}
		if lastSum.HasResult {
			lifetimeResult += lastSum.Result
			hasAnyResult = true
		}
	}
	lifetimeGain := lifetimeResult - lifetimeInvested
	hasLifetime := lifetimeInvested > 0 && hasAnyResult

	renderReport(w, typ, len(ofType), rows, lifetimeInvested, lifetimeGain, hasLifetime,
		nValid, sumPct, sumGain)
	return nil
}

func renderReport(w io.Writer, typ domain.AssetType, nAssets int, rows []rowEntry,
	lifetimeInvested, lifetimeGain float64, hasLifetime bool,
	nValid int, sumPct, sumGain float64) {

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Tipo: %s · %d activo(s) · %d mes(es) con resultado\n",
		typ.Display(), nAssets, len(rows))
	fmt.Fprintln(w, sep)

	twH := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twH, "  Total Aportado\t%s\n", money.USD(lifetimeInvested))
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice Medio Mensual\t%+.2f%%\n", sumPct/float64(nValid))
		fmt.Fprintf(twH, "  Gañanzas/Perdas Medias Mensuais\t%s\n", money.SignedUSD(sumGain/float64(nValid)))
	} else {
		fmt.Fprintln(twH, "  Índice Medio Mensual\t—")
		fmt.Fprintln(twH, "  Gañanzas/Perdas Medias Mensuais\t—")
	}
	if hasLifetime {
		fmt.Fprintf(twH, "  Total Gañanzas/Perdas\t%s\n", money.SignedUSD(lifetimeGain))
	} else {
		fmt.Fprintln(twH, "  Total Gañanzas/Perdas\t—")
	}
	twH.Flush()
	fmt.Fprintln(w, sep)

	var tbuf bytes.Buffer
	twT := tabwriter.NewWriter(&tbuf, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tNo activo\tÍndice\tG/P\tResultado\t")
	for _, row := range rows {
		var idxStr, gainStr string
		if row.hasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.gainPct)
			gainStr = money.SignedUSD(row.gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		fmt.Fprintf(twT, "  %d\t%d\t%s\t%s\t%s\t%s\t%s\t\n",
			row.year, row.month,
			money.USD(row.aporte), money.USD(row.holding),
			idxStr, gainStr, money.USD(row.result))
	}
	twT.Flush()
	writeColoredRows(w, tbuf.String(), rows)
	fmt.Fprintln(w, sep)
}

// writeColoredRows imprime as liñas xa formatadas: a primeira (cabeceira) sen
// cor, e cada fila de datos envolvida no código ANSI segundo o seu G/P.
func writeColoredRows(w io.Writer, formatted string, rows []rowEntry) {
	lines := strings.Split(strings.TrimRight(formatted, "\n"), "\n")
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(w, lines[0])
	for i, line := range lines[1:] {
		if i < len(rows) && rows[i].hasMetrics {
			fmt.Fprintln(w, colors.ForGain(rows[i].gain)+line+colors.Reset)
		} else {
			fmt.Fprintln(w, line)
		}
	}
}
