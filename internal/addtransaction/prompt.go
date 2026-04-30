package addtransaction

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir nova transacción ---\n")

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
	fmt.Fprintf(w, "Sobre %s — %s\n", chosen.Type.Display(), chosen.Name)

	isVenda, err := promptTransactionType(r, w)
	if err != nil {
		return err
	}

	amount, err := prompts.Amount(r, w)
	if err != nil {
		return err
	}
	month, year, err := promptDateNotBefore(r, w, chosen.Month, chosen.Year)
	if err != nil {
		return err
	}

	storedAmount := amount
	typeLabel := "COMPRA"
	if isVenda {
		storedAmount = -amount
		typeLabel = "VENDA"
	}

	tx := domain.Transaction{
		AssetID:   chosen.ID,
		AmountUSD: storedAmount,
		Month:     month,
		Year:      year,
	}
	id, err := repo.InsertTransaction(tx)
	if err != nil {
		return fmt.Errorf("gardando transacción: %w", err)
	}

	fmt.Fprintf(w, "✓ Transacción gardada #%d sobre %s — %s: %s %.2f USD — %02d/%d\n",
		id, chosen.Type.Display(), chosen.Name, typeLabel, amount, month, year)
	return nil
}

// promptDateNotBefore pide mes e ano e valida que (year*12+month) >= o mínimo
// dado pola data do activo. Re-pregunta mes+ano se é anterior.
func promptDateNotBefore(r *bufio.Reader, w io.Writer, minMonth, minYear int) (int, int, error) {
	for {
		month, err := prompts.Month(r, w)
		if err != nil {
			return 0, 0, err
		}
		year, err := prompts.Year(r, w)
		if err != nil {
			return 0, 0, err
		}
		if year*12+month >= minYear*12+minMonth {
			return month, year, nil
		}
		fmt.Fprintf(w, "⚠ A transacción non pode ser anterior á data do activo (%02d/%d). Tenta de novo.\n",
			minMonth, minYear)
	}
}

func promptTransactionType(r *bufio.Reader, w io.Writer) (bool, error) {
	fmt.Fprintln(w, "Tipo de transacción:")
	fmt.Fprintln(w, "  [1] Compra")
	fmt.Fprintln(w, "  [2] Venda")
	for {
		fmt.Fprint(w, "> ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return false, err
		}
		switch line {
		case "1":
			return false, nil
		case "2":
			return true, nil
		}
		fmt.Fprintln(w, "⚠ Tipo non válido, escolle 1 (compra) ou 2 (venda)")
	}
}
