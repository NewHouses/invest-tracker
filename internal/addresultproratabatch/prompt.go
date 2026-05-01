package addresultproratabatch

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

type entry struct {
	asset   domain.Asset
	holding float64
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir resultados proporcionais por tipo en serie ---\n")

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

	fmt.Fprintln(w, "Introduce o mes inicial:")
	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	totalSaved := 0
	monthsSkipped := 0
	curMonth, curYear := month, year

	for {
		var eligible, excluded []entry
		var totalHolding float64
		for _, a := range ofType {
			sum, err := repo.MonthlySummary(a.ID, curYear, curMonth)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if sum.EstimatedHolding > 0 {
				eligible = append(eligible, entry{asset: a, holding: sum.EstimatedHolding})
				totalHolding += sum.EstimatedHolding
			} else {
				excluded = append(excluded, entry{asset: a, holding: sum.EstimatedHolding})
			}
		}

		fmt.Fprintf(w, "\n--- %02d/%d ---\n", curMonth, curYear)
		if len(eligible) == 0 {
			fmt.Fprintln(w, "  ↷ Sen activos elixibles neste mes. Saltado.")
			monthsSkipped++
		} else {
			for _, e := range eligible {
				share := e.holding / totalHolding * 100
				fmt.Fprintf(w, "  - %s · no activo: %.2f USD (%.2f%%)\n",
					e.asset.Name, e.holding, share)
			}
			for _, e := range excluded {
				fmt.Fprintf(w, "  - %s · no activo: %.2f USD (excluído)\n",
					e.asset.Name, e.holding)
			}
			fmt.Fprintf(w, "  Suma de holdings: %.2f USD\n", totalHolding)

			gain, skip, err := promptOptionalSignedAmount(r, w,
				"  Ganhanza/Perda total do tipo (USD, baleiro = saltar): ")
			if err != nil {
				return err
			}
			if skip {
				fmt.Fprintln(w, "  ↷ Saltado.")
				monthsSkipped++
			} else {
				monthSaved := 0
				for _, e := range eligible {
					share := e.holding / totalHolding
					gainI := share * gain
					resultI := e.holding + gainI
					if resultI <= 0 {
						fmt.Fprintf(w, "    ⚠ %s: resultado <= 0 (%.2f). Saltado.\n",
							e.asset.Name, resultI)
						continue
					}
					mr := domain.MonthlyResult{
						AssetID:   e.asset.ID,
						ResultUSD: resultI,
						Month:     curMonth,
						Year:      curYear,
					}
					id, err := repo.InsertMonthlyResult(mr)
					if err != nil {
						return fmt.Errorf("gardando resultado para %s: %w", e.asset.Name, err)
					}
					pct := gainI / e.holding * 100
					fmt.Fprintf(w, "    ✓ #%d %s: %.2f → %.2f USD (%+.2f, %+.2f%%)\n",
						id, e.asset.Name, e.holding, resultI, gainI, pct)
					monthSaved++
				}
				totalSaved += monthSaved
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

	fmt.Fprintf(w, "\n✓ Engadíronse %d resultado(s) e saltáronse %d mes(es).\n",
		totalSaved, monthsSkipped)
	return nil
}

func promptOptionalSignedAmount(r *bufio.Reader, w io.Writer, label string) (float64, bool, error) {
	for {
		fmt.Fprint(w, label)
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, false, err
		}
		if line == "" {
			return 0, true, nil
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil {
			return v, false, nil
		}
		fmt.Fprintln(w, "  ⚠ Valor non válido (usa . ou , como separador decimal)")
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
