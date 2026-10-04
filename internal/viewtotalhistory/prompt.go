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
		var totalTx, fondos, baseResult float64
		for _, a := range assets {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			// Aporte Mensual = Σ transaccións de tódolos activos − dividendos.
			// Inclúese tamén o investimento de activos sen resultado neste mes.
			totalTx += sum.InvestedInMonth
			// Fondos e Resultado: só dos activos con resultado neste mes para
			// que ámbolos dous lados da G/P inclúan o mesmo conxunto.
			if !sum.HasResult {
				continue
			}
			fondos += sum.EstimatedHolding
			baseResult += sum.Result
		}
		div, err := repo.SumDividends(ym.Year, ym.Month)
		if err != nil {
			return fmt.Errorf("sumando dividendos de %d/%d: %w", ym.Month, ym.Year, err)
		}
		aporte := totalTx - div
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
	fmt.Fprintf(twH, "  Aporte histórico total\t%s\n", money.USD(lifetimeAporte))
	if nValid > 0 {
		fmt.Fprintf(twH, "  Índice Medio\t%+.2f%%\n", sumPct/float64(nValid))
		fmt.Fprintf(twH, "  G/P Media\t%s\n", money.SignedUSD(sumGain/float64(nValid)))
	} else {
		fmt.Fprintln(twH, "  Índice Medio\t—")
		fmt.Fprintln(twH, "  G/P Media\t—")
	}
	if hasLifetime {
		fmt.Fprintf(twH, "  G/P Total\t%s\n", money.SignedUSD(lifetimeGain))
	} else {
		fmt.Fprintln(twH, "  G/P Total\t—")
	}
	fmt.Fprintf(twH, "  Dividendos totais\t%s\n", money.USD(totalDiv))
	twH.Flush()
	fmt.Fprintln(w, sep)

	var tbuf bytes.Buffer
	twT := tabwriter.NewWriter(&tbuf, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(twT, "  Ano\tMes\tAporte Mensual\tFondos\tÍndice\tG/P\tDividendos\tResultado\t")
	for _, row := range rows {
		var idxStr, gainStr string
		if row.hasMetrics {
			idxStr = fmt.Sprintf("%+.2f%%", row.gainPct)
			gainStr = money.SignedUSD(row.gain)
		} else {
			idxStr = "n/a"
			gainStr = "—"
		}
		// Aporte pode ser negativo se os dividendos exceden as transaccións.
		fmt.Fprintf(twT, "  %d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			row.year, row.month,
			money.SignedUSD(row.aporte), money.USD(row.fondos),
			idxStr, gainStr,
			money.USD(row.dividends), money.USD(row.result))
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
