package addtxmonth

import (
	"bufio"
	"io"
	"slices"
	"strconv"
	"strings"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Engadir transaccións do mes ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
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

	isCreated := func(a domain.Asset) bool {
		return a.CreatedBy(domain.YearMonth{Year: year, Month: month})
	}
	if !slices.ContainsFunc(assets, isCreated) {
		i18n.Fprintf(w, "Non hai activos creados en %02d/%d ou antes.\n", month, year)
		return nil
	}

	i18n.Fprintf(w, "\nAporte por activo en %02d/%d (deixa en branco para saltar):\n", month, year)

	var saved, skipped int
	var totalAmount float64
	for _, a := range assets {
		if !isCreated(a) {
			i18n.Fprintf(w, "  %s — %s: omitido (creado en %02d/%d)\n",
				a.Type.Display(), a.Name, a.Month, a.Year)
			continue
		}
		amount, skip, err := promptOptionalAmount(r, w,
			i18n.Sprintf("  %s — %s: ", a.Type.Display(), a.Name))
		if err != nil {
			return err
		}
		if skip {
			skipped++
			continue
		}
		tx := domain.Transaction{
			AssetID:   a.ID,
			AmountUSD: amount,
			Month:     month,
			Year:      year,
		}
		id, err := repo.InsertTransaction(tx)
		if err != nil {
			return i18n.Errorf("gardando transacción de %s: %w", a.Name, err)
		}
		i18n.Fprintf(w, "    ✓ #%d %.2f USD\n", id, amount)
		saved++
		totalAmount += amount
	}

	i18n.Fprintf(w, "\n✓ %d transacción(s) engadidas (%.2f USD total), %d saltado(s).\n",
		saved, totalAmount, skipped)
	return nil
}

// promptOptionalAmount permite cantidade > 0 ou liña baleira para saltar.
func promptOptionalAmount(r *bufio.Reader, w io.Writer, label string) (float64, bool, error) {
	for {
		i18n.Fprint(w, label)
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, false, err
		}
		if line == "" {
			return 0, true, nil
		}
		normalized := strings.ReplaceAll(line, ",", ".")
		v, perr := strconv.ParseFloat(normalized, 64)
		if perr == nil && domain.ValidAmount(v) {
			return v, false, nil
		}
		i18n.Fprintln(w, "  ⚠ Valor non válido (debe ser > 0; baleiro para saltar)")
	}
}
