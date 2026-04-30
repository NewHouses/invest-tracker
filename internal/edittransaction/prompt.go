package edittransaction

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error)
	UpdateTransaction(domain.Transaction) error
}

const (
	fieldAmount = 1
	fieldType   = 2
	fieldDate   = 3
)

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Editar unha transacción ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa opción 1.")
		return nil
	}

	chosenAsset, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}

	txs, err := repo.ListTransactionsByAsset(chosenAsset.ID)
	if err != nil {
		return fmt.Errorf("listando transaccións: %w", err)
	}
	if len(txs) == 0 {
		fmt.Fprintln(w, "Este activo non ten transaccións extras rexistradas.")
		return nil
	}

	chosenTx, err := promptTransactionSelection(r, w, chosenAsset, txs)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Transacción seleccionada: %s\n", describeTx(chosenTx))

	field, err := promptFieldChoice(r, w)
	if err != nil {
		return err
	}

	updated := chosenTx
	switch field {
	case fieldAmount:
		newAmount, err := prompts.Amount(r, w)
		if err != nil {
			return err
		}
		// Conserva o signo actual (compra/venda).
		if chosenTx.AmountUSD < 0 {
			updated.AmountUSD = -newAmount
		} else {
			updated.AmountUSD = newAmount
		}
	case fieldType:
		isVenda, err := prompts.SelectTransactionType(r, w)
		if err != nil {
			return err
		}
		// Aplica o novo signo á cantidade absoluta actual.
		abs := chosenTx.AmountUSD
		if abs < 0 {
			abs = -abs
		}
		if isVenda {
			updated.AmountUSD = -abs
		} else {
			updated.AmountUSD = abs
		}
	case fieldDate:
		m, y, err := prompts.DateNotBefore(r, w, chosenAsset.Month, chosenAsset.Year)
		if err != nil {
			return err
		}
		updated.Month = m
		updated.Year = y
	}

	if err := repo.UpdateTransaction(updated); err != nil {
		return fmt.Errorf("actualizando transacción: %w", err)
	}

	fmt.Fprintf(w, "✓ Transacción #%d actualizada: %s\n", updated.ID, describeTx(updated))
	return nil
}

func promptTransactionSelection(r *bufio.Reader, w io.Writer, asset domain.Asset, txs []domain.Transaction) (domain.Transaction, error) {
	fmt.Fprintf(w, "Transaccións de %s — %s:\n", asset.Type.Display(), asset.Name)
	for i, tx := range txs {
		fmt.Fprintf(w, "  [%d] %s\n", i+1, describeTx(tx))
	}
	for {
		fmt.Fprintf(w, "Selecciona a transacción a editar (1-%d): ", len(txs))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.Transaction{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(txs) {
			return txs[idx-1], nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida")
	}
}

func promptFieldChoice(r *bufio.Reader, w io.Writer) (int, error) {
	fmt.Fprintln(w, "Que campo queres editar?")
	fmt.Fprintln(w, "  [1] Cantidade (USD)")
	fmt.Fprintln(w, "  [2] Tipo (Compra/Venda)")
	fmt.Fprintln(w, "  [3] Data (mes/ano)")
	for {
		fmt.Fprint(w, "> ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		v, perr := strconv.Atoi(line)
		if perr == nil && v >= 1 && v <= 3 {
			return v, nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida, escolle 1-3")
	}
}

// describeTx formatea unha transacción como "MM/YYYY — TIPO X.XX USD".
func describeTx(tx domain.Transaction) string {
	amount := tx.AmountUSD
	label := "COMPRA"
	if amount < 0 {
		amount = -amount
		label = "VENDA"
	}
	return fmt.Sprintf("%02d/%d — %s %.2f USD", tx.Month, tx.Year, label, amount)
}
