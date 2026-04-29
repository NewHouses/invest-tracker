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
	ListInvestments() ([]domain.Investment, error)
	InsertMonthlyResult(domain.MonthlyResult) (int64, error)
	TotalInvested(investmentID int64) (float64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir resultado mensual ---\n")

	invs, err := repo.ListInvestments()
	if err != nil {
		return fmt.Errorf("listando investimentos: %w", err)
	}
	if len(invs) == 0 {
		fmt.Fprintln(w, "Aínda non hai investimentos. Engade un primeiro coa opción 1.")
		return nil
	}

	chosen, err := promptSelection(r, w, invs)
	if err != nil {
		return err
	}

	total, err := repo.TotalInvested(chosen.ID)
	if err != nil {
		return fmt.Errorf("calculando total investido: %w", err)
	}
	fmt.Fprintf(w, "Sobre %s — %s (investido: %.2f USD)\n",
		chosen.Type.Display(), chosen.Name, total)

	result, err := promptResult(r, w)
	if err != nil {
		return err
	}
	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	mr := domain.MonthlyResult{
		InvestmentID: chosen.ID,
		ResultUSD:    result,
		Month:        month,
		Year:         year,
	}
	id, err := repo.InsertMonthlyResult(mr)
	if err != nil {
		return fmt.Errorf("gardando resultado: %w", err)
	}

	fmt.Fprintf(w, "✓ Resultado gardado #%d sobre %s — %s: %.2f USD — %02d/%d\n",
		id, chosen.Type.Display(), chosen.Name, result, month, year)
	fmt.Fprintf(w, "Investido total: %.2f USD\n", total)
	gain := result - total
	if total > 0 {
		pct := gain / total * 100
		fmt.Fprintf(w, "Ganhanzas/Perdas: %+.2f USD (%+.2f%%)\n", gain, pct)
	} else {
		fmt.Fprintf(w, "Ganhanzas/Perdas: %+.2f USD (n/a%%)\n", gain)
	}
	return nil
}

func promptSelection(r *bufio.Reader, w io.Writer, invs []domain.Investment) (domain.Investment, error) {
	fmt.Fprintln(w, "Investimentos:")
	for i, inv := range invs {
		fmt.Fprintf(w, "  [%d] %s — %s\n", i+1, inv.Type.Display(), inv.Name)
	}
	for {
		fmt.Fprintf(w, "Selecciona (1-%d): ", len(invs))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.Investment{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(invs) {
			return invs[idx-1], nil
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
