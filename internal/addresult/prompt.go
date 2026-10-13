package addresult

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	InsertMonthlyResult(domain.MonthlyResult) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Engadir resultado mensual ---\n")

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
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	eligible, err := closemonth.Eligible(repo, year, month)
	if err != nil {
		return err
	}

	if len(eligible) == 0 {
		i18n.Fprintf(w, "Non hai activos con capital investido en %02d/%d.\n", month, year)
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
		AssetID:   chosen.Asset.ID,
		ResultUSD: result,
		Month:     month,
		Year:      year,
	}
	id, err := repo.InsertMonthlyResult(mr)
	if err != nil {
		return i18n.Errorf("gardando resultado: %w", err)
	}

	i18n.Fprintf(w, "✓ Resultado gardado #%d sobre %s — %s: %.2f USD — %02d/%d\n",
		id, chosen.Asset.Type.Display(), chosen.Asset.Name, result, month, year)
	i18n.Fprintf(w, "No activo: %.2f USD\n", chosen.Holding)
	gain := result - chosen.Holding
	if chosen.Holding > 0 {
		pct := gain / chosen.Holding * 100
		i18n.Fprintf(w, "Gañanzas/Perdas: %+.2f USD (%+.2f%%)\n", gain, pct)
	} else {
		i18n.Fprintf(w, "Gañanzas/Perdas: %+.2f USD (n/a%%)\n", gain)
	}
	return nil
}

func promptEligibleSelection(r *bufio.Reader, w io.Writer, eligible []closemonth.EligibleAsset) (closemonth.EligibleAsset, error) {
	i18n.Fprintln(w, "Investimentos:")
	for i, e := range eligible {
		i18n.Fprintf(w, "  [%d] %s — %s (no activo: %.2f USD)\n",
			i+1, e.Asset.Type.Display(), e.Asset.Name, e.Holding)
	}
	for {
		i18n.Fprintf(w, "Selecciona (1-%d): ", len(eligible))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return closemonth.EligibleAsset{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(eligible) {
			return eligible[idx-1], nil
		}
		i18n.Fprintln(w, "⚠ Selección non válida")
	}
}

func promptResult(r *bufio.Reader, w io.Writer) (float64, error) {
	for {
		i18n.Fprint(w, "Resultado (USD): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && domain.ValidAmount(v) {
			return v, nil
		}
		i18n.Fprintln(w, "⚠ Resultado non válido, debe ser un número maior ca 0")
	}
}
