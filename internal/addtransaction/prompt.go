package addtransaction

import (
	"bufio"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Engadir nova transacción ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	chosen, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}
	i18n.Fprintf(w, "Sobre %s — %s\n", chosen.Type.Display(), chosen.Name)

	isVenda, err := prompts.SelectTransactionType(r, w)
	if err != nil {
		return err
	}

	amount, err := prompts.Amount(r, w)
	if err != nil {
		return err
	}
	month, year, err := prompts.DateNotBefore(r, w, chosen.Month, chosen.Year)
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
		return i18n.Errorf("gardando transacción: %w", err)
	}

	i18n.Fprintf(w, "✓ Transacción gardada #%d sobre %s — %s: %s %.2f USD — %02d/%d\n",
		id, chosen.Type.Display(), chosen.Name, i18n.T(typeLabel), amount, month, year)
	return nil
}
