package addresultprorata

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
	fmt.Fprint(w, "\n--- Engadir resultados proporcionais por tipo ---\n")

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

	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	var eligible []entry
	var excluded []entry
	var totalHolding float64
	for _, a := range ofType {
		sum, err := repo.MonthlySummary(a.ID, year, month)
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
	if len(eligible) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s con capital en %02d/%d.\n",
			typ.Display(), month, year)
		return nil
	}

	fmt.Fprintf(w, "\nActivos de tipo %s en %02d/%d:\n", typ.Display(), month, year)
	for _, e := range eligible {
		share := e.holding / totalHolding * 100
		fmt.Fprintf(w, "  - %s · no activo: %.2f USD (%.2f%%)\n",
			e.asset.Name, e.holding, share)
	}
	for _, e := range excluded {
		fmt.Fprintf(w, "  - %s · no activo: %.2f USD (excluído: non conta para a repartición)\n",
			e.asset.Name, e.holding)
	}
	fmt.Fprintf(w, "Suma de holdings: %.2f USD\n\n", totalHolding)

	gain, err := promptSignedAmount(r, w,
		"Ganhanza/Perda total do tipo (USD, negativo = perda): ")
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "\nDistribución proporcional:\n")
	saved, skipped := 0, 0
	for _, e := range eligible {
		share := e.holding / totalHolding
		gainI := share * gain
		resultI := e.holding + gainI
		if resultI <= 0 {
			fmt.Fprintf(w, "  ⚠ %s: resultado calculado <= 0 (%.2f USD). Saltado.\n",
				e.asset.Name, resultI)
			skipped++
			continue
		}

		mr := domain.MonthlyResult{
			AssetID:   e.asset.ID,
			ResultUSD: resultI,
			Month:     month,
			Year:      year,
		}
		id, err := repo.InsertMonthlyResult(mr)
		if err != nil {
			return fmt.Errorf("gardando resultado para %s: %w", e.asset.Name, err)
		}
		pct := gainI / e.holding * 100
		fmt.Fprintf(w, "  ✓ #%d %s: %.2f → %.2f USD (%+.2f USD, %+.2f%%)\n",
			id, e.asset.Name, e.holding, resultI, gainI, pct)
		saved++
	}

	fmt.Fprintf(w, "\n✓ Engadíronse %d resultado(s) e saltáronse %d en %02d/%d.\n",
		saved, skipped, month, year)
	return nil
}

// promptSignedAmount admite negativos e cero (a diferenza de prompts.Amount).
func promptSignedAmount(r *bufio.Reader, w io.Writer, label string) (float64, error) {
	for {
		fmt.Fprint(w, label)
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Valor non válido (usa . ou , como separador decimal)")
	}
}
