package deletetransaction

import (
	"bufio"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error)
	DeleteTransaction(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Eliminar transacción ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	chosenAsset, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}

	txs, err := repo.ListTransactionsByAsset(chosenAsset.ID)
	if err != nil {
		return i18n.Errorf("listando transaccións: %w", err)
	}
	if len(txs) == 0 {
		i18n.Fprintln(w, "Este activo non ten transaccións extras rexistradas (a compra inicial non é unha transacción).")
		return nil
	}

	chosenTx, err := promptTransactionSelection(r, w, chosenAsset, txs)
	if err != nil {
		return err
	}

	if err := repo.DeleteTransaction(chosenTx.ID); err != nil {
		return i18n.Errorf("eliminando transacción: %w", err)
	}

	i18n.Fprintf(w, "✓ Transacción #%d eliminada (%02d/%d — %.2f USD)\n",
		chosenTx.ID, chosenTx.Month, chosenTx.Year, chosenTx.AmountUSD)
	return nil
}

func promptTransactionSelection(r *bufio.Reader, w io.Writer, asset domain.Asset, txs []domain.Transaction) (domain.Transaction, error) {
	i18n.Fprintf(w, "Transaccións de %s — %s:\n", asset.Type.Display(), asset.Name)
	for i, tx := range txs {
		i18n.Fprintf(w, "  [%d] %02d/%d — %.2f USD\n", i+1, tx.Month, tx.Year, tx.AmountUSD)
	}
	for {
		i18n.Fprintf(w, "Selecciona a transacción a eliminar (1-%d): ", len(txs))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.Transaction{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(txs) {
			return txs[idx-1], nil
		}
		i18n.Fprintln(w, "⚠ Selección non válida")
	}
}
