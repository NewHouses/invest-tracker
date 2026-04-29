package deletetransaction

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
	DeleteTransaction(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Eliminar transacción ---\n")

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
		fmt.Fprintln(w, "Este activo non ten transaccións extras rexistradas (a compra inicial non é unha transacción).")
		return nil
	}

	chosenTx, err := promptTransactionSelection(r, w, chosenAsset, txs)
	if err != nil {
		return err
	}

	if err := repo.DeleteTransaction(chosenTx.ID); err != nil {
		return fmt.Errorf("eliminando transacción: %w", err)
	}

	fmt.Fprintf(w, "✓ Transacción #%d eliminada (%02d/%d — %.2f USD)\n",
		chosenTx.ID, chosenTx.Month, chosenTx.Year, chosenTx.AmountUSD)
	return nil
}

func promptTransactionSelection(r *bufio.Reader, w io.Writer, asset domain.Asset, txs []domain.Transaction) (domain.Transaction, error) {
	fmt.Fprintf(w, "Transaccións de %s — %s:\n", asset.Type.Display(), asset.Name)
	for i, tx := range txs {
		fmt.Fprintf(w, "  [%d] %02d/%d — %.2f USD\n", i+1, tx.Month, tx.Year, tx.AmountUSD)
	}
	for {
		fmt.Fprintf(w, "Selecciona a transacción a eliminar (1-%d): ", len(txs))
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
