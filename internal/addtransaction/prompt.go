package addtransaction

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListInvestments() ([]domain.Investment, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir nova transacción ---\n")

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
	fmt.Fprintf(w, "Sobre %s — %s\n", chosen.Type.Display(), chosen.Name)

	amount, err := prompts.Amount(r, w)
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

	tx := domain.Transaction{
		InvestmentID: chosen.ID,
		AmountUSD:    amount,
		Month:        month,
		Year:         year,
	}
	id, err := repo.InsertTransaction(tx)
	if err != nil {
		return fmt.Errorf("gardando transacción: %w", err)
	}

	fmt.Fprintf(w, "✓ Transacción gardada #%d sobre %s — %s: %.2f USD — %02d/%d\n",
		id, chosen.Type.Display(), chosen.Name, amount, month, year)
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
