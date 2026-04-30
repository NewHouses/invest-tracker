package viewtransactions

import (
	"bufio"
	"fmt"
	"io"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error)
}

const sep = "============================================================"

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Transaccións dun activo ---\n")

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

	txs, err := repo.ListTransactionsByAsset(chosen.ID)
	if err != nil {
		return fmt.Errorf("listando transaccións: %w", err)
	}

	if len(txs) == 0 {
		fmt.Fprintf(w, "%s — %s non ten transaccións rexistradas.\n",
			chosen.Type.Display(), chosen.Name)
		fmt.Fprintln(w, "(Nota: a compra inicial do activo non se conta como transacción.)")
		return nil
	}

	renderTable(w, chosen, txs)
	return nil
}

func renderTable(w io.Writer, asset domain.Asset, txs []domain.Transaction) {
	var totalCompra, totalVenda float64
	for _, tx := range txs {
		if tx.AmountUSD >= 0 {
			totalCompra += tx.AmountUSD
		} else {
			totalVenda += -tx.AmountUSD
		}
	}
	neto := totalCompra - totalVenda

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Transaccións de %s — %s\n", asset.Type.Display(), asset.Name)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Total: %d transacción(s)\n", len(txs))
	fmt.Fprintf(w, "  Compras: %.2f USD · Vendas: %.2f USD · Neto: %+.2f USD\n",
		totalCompra, totalVenda, neto)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "  ID\tAno\tMes\tTipo\tCantidade\t")
	for _, tx := range txs {
		amount := tx.AmountUSD
		typeLabel := "COMPRA"
		if amount < 0 {
			amount = -amount
			typeLabel = "VENDA"
		}
		fmt.Fprintf(tw, "  %d\t%d\t%d\t%s\t%.2f USD\t\n",
			tx.ID, tx.Year, tx.Month, typeLabel, amount,
		)
	}
	tw.Flush()
	fmt.Fprintln(w, sep)
}
