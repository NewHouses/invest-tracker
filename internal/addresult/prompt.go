package addresult

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

type eligibleEntry struct {
	asset   domain.Asset
	holding float64
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir resultado mensual ---\n")

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
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	var eligible []eligibleEntry
	for _, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, year, month)
		if err != nil {
			return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
		}
		if sum.EstimatedHolding > 0 {
			eligible = append(eligible, eligibleEntry{asset: a, holding: sum.EstimatedHolding})
		}
	}

	if len(eligible) == 0 {
		fmt.Fprintf(w, "Non hai activos con capital investido en %02d/%d.\n", month, year)
		return nil
	}

	chosen, err := promptEligibleSelection(r, w, eligible)
	if err != nil {
		return err
	}

	result, err := promptResult(r, w)
	if err != nil {
		return err
	}

	mr := domain.MonthlyResult{
		AssetID:   chosen.asset.ID,
		ResultUSD: result,
		Month:     month,
		Year:      year,
	}
	id, err := repo.InsertMonthlyResult(mr)
	if err != nil {
		return fmt.Errorf("gardando resultado: %w", err)
	}

	fmt.Fprintf(w, "✓ Resultado gardado #%d sobre %s — %s: %.2f USD — %02d/%d\n",
		id, chosen.asset.Type.Display(), chosen.asset.Name, result, month, year)
	fmt.Fprintf(w, "No activo: %.2f USD\n", chosen.holding)
	gain := result - chosen.holding
	if chosen.holding > 0 {
		pct := gain / chosen.holding * 100
		fmt.Fprintf(w, "Ganhanzas/Perdas: %+.2f USD (%+.2f%%)\n", gain, pct)
	} else {
		fmt.Fprintf(w, "Ganhanzas/Perdas: %+.2f USD (n/a%%)\n", gain)
	}
	return nil
}

func promptEligibleSelection(r *bufio.Reader, w io.Writer, eligible []eligibleEntry) (eligibleEntry, error) {
	fmt.Fprintln(w, "Investimentos:")
	for i, e := range eligible {
		fmt.Fprintf(w, "  [%d] %s — %s (no activo: %.2f USD)\n",
			i+1, e.asset.Type.Display(), e.asset.Name, e.holding)
	}
	for {
		fmt.Fprintf(w, "Selecciona (1-%d): ", len(eligible))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return eligibleEntry{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(eligible) {
			return eligible[idx-1], nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida")
	}
}

func promptResult(r *bufio.Reader, w io.Writer) (float64, error) {
	for {
		fmt.Fprint(w, "Resultado (USD): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && v > 0 {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Resultado non válido, debe ser un número maior ca 0")
	}
}
