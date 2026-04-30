package addresultbatch

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

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir resultados mensuais en serie ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa opción 1.")
		return nil
	}

	chosen, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Sobre %s — %s (creado en %02d/%d)\n",
		chosen.Type.Display(), chosen.Name, chosen.Month, chosen.Year)

	fmt.Fprintln(w, "Introduce o mes inicial:")
	month, year, err := prompts.DateNotBefore(r, w, chosen.Month, chosen.Year)
	if err != nil {
		return err
	}

	saved, skipped := 0, 0
	curMonth, curYear := month, year

	for {
		sum, err := repo.MonthlySummary(chosen.ID, curYear, curMonth)
		if err != nil {
			return fmt.Errorf("calculando resumo: %w", err)
		}

		fmt.Fprintf(w, "\n--- %02d/%d ---\n", curMonth, curYear)
		if sum.EstimatedHolding <= 0 {
			fmt.Fprintln(w, "  ↷ Sen capital no activo neste mes. Saltado.")
			skipped++
		} else {
			fmt.Fprintf(w, "  No activo: %.2f USD\n", sum.EstimatedHolding)
			if sum.HasResult {
				fmt.Fprintf(w, "  Xa hai un resultado rexistrado: %.2f USD. Baleiro mantenno.\n", sum.Result)
			}

			result, skip, err := promptOptionalResult(r, w)
			if err != nil {
				return err
			}
			if skip {
				fmt.Fprintln(w, "  ↷ Saltado.")
				skipped++
			} else {
				mr := domain.MonthlyResult{
					AssetID:   chosen.ID,
					ResultUSD: result,
					Month:     curMonth,
					Year:      curYear,
				}
				id, err := repo.InsertMonthlyResult(mr)
				if err != nil {
					return fmt.Errorf("gardando resultado: %w", err)
				}
				gain := result - sum.EstimatedHolding
				pct := gain / sum.EstimatedHolding * 100
				fmt.Fprintf(w, "  ✓ Gardado #%d — Ganhanzas/Perdas: %+.2f USD (%+.2f%%)\n",
					id, gain, pct)
				saved++
			}
		}

		cont, err := promptContinue(r, w)
		if err != nil {
			return err
		}
		if !cont {
			break
		}
		curMonth, curYear = nextMonth(curMonth, curYear)
	}

	fmt.Fprintf(w, "\n✓ Engadíronse %d resultado(s) e saltáronse %d sobre %s — %s.\n",
		saved, skipped, chosen.Type.Display(), chosen.Name)
	return nil
}

func promptOptionalResult(r *bufio.Reader, w io.Writer) (float64, bool, error) {
	for {
		fmt.Fprint(w, "  Resultado (USD, baleiro = saltar): ")
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
		fmt.Fprintln(w, "  ⚠ Resultado non válido (debe ser > 0, ou baleiro para saltar)")
	}
}

func promptContinue(r *bufio.Reader, w io.Writer) (bool, error) {
	for {
		fmt.Fprint(w, "  Continuar co seguinte mes? [S/n]: ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return false, err
		}
		low := strings.ToLower(line)
		if low == "" || low == "s" || low == "si" || low == "sí" || low == "y" || low == "yes" {
			return true, nil
		}
		if low == "n" || low == "no" || low == "non" {
			return false, nil
		}
		fmt.Fprintln(w, "  ⚠ Resposta non válida, escolle S ou N")
	}
}

func nextMonth(m, y int) (int, int) {
	if m == 12 {
		return 1, y + 1
	}
	return m + 1, y
}
