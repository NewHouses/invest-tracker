package closemonth

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	InsertMonthlyResult(domain.MonthlyResult) (int64, error)
}

type eligibleAsset struct {
	asset       domain.Asset
	holding     float64
	prev        float64
	hasPrevious bool
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Pechar mes (resultados) ---\n")

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}

	var eligible []eligibleAsset
	for _, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, year, month)
		if err != nil {
			return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
		}
		if sum.EstimatedHolding > 0 {
			eligible = append(eligible, eligibleAsset{
				asset:       a,
				holding:     sum.EstimatedHolding,
				prev:        sum.Result,
				hasPrevious: sum.HasResult,
			})
		}
	}

	if len(eligible) == 0 {
		fmt.Fprintf(w, "Non hai activos con capital investido para %02d/%d.\n", month, year)
		return nil
	}

	fmt.Fprintf(w, "Vanse pedir os resultados para %d activo(s) en %02d/%d.\n", len(eligible), month, year)
	fmt.Fprintln(w, "(Deixa baleiro para saltar un activo.)")

	saved, skipped := 0, 0
	for i, ea := range eligible {
		fmt.Fprintf(w, "\n[%d/%d] %s — %s (no activo: %.2f USD)\n",
			i+1, len(eligible), ea.asset.Type.Display(), ea.asset.Name, ea.holding)
		if ea.hasPrevious {
			fmt.Fprintf(w, "   Xa hai un resultado rexistrado este mes: %.2f USD. Baleiro mantenno.\n", ea.prev)
		}

		result, skip, err := promptOptionalResult(r, w)
		if err != nil {
			return err
		}
		if skip {
			fmt.Fprintln(w, "   ↷ Saltado.")
			skipped++
			continue
		}

		mr := domain.MonthlyResult{
			AssetID:   ea.asset.ID,
			ResultUSD: result,
			Month:     month,
			Year:      year,
		}
		id, err := repo.InsertMonthlyResult(mr)
		if err != nil {
			return fmt.Errorf("gardando resultado para %s: %w", ea.asset.Name, err)
		}

		gain := result - ea.holding
		if ea.holding > 0 {
			pct := gain / ea.holding * 100
			fmt.Fprintf(w, "   ✓ Gardado #%d — Gañanzas/Perdas: %+.2f USD (%+.2f%%)\n", id, gain, pct)
		} else {
			fmt.Fprintf(w, "   ✓ Gardado #%d — Gañanzas/Perdas: %+.2f USD (n/a%%)\n", id, gain)
		}
		saved++
	}

	fmt.Fprintf(w, "\n✓ Pechouse %02d/%d: %d resultado(s) gardado(s), %d saltado(s).\n",
		month, year, saved, skipped)
	return nil
}

func promptOptionalResult(r *bufio.Reader, w io.Writer) (float64, bool, error) {
	for {
		fmt.Fprint(w, "   Resultado (USD, baleiro = saltar): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, false, err
		}
		if line == "" {
			return 0, true, nil
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && v > 0 {
			return v, false, nil
		}
		fmt.Fprintln(w, "   ⚠ Resultado non válido (debe ser > 0, ou baleiro para saltar)")
	}
}
