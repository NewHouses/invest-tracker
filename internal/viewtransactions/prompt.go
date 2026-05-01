package viewtransactions

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error)
}

const sep = "============================================================"

type displayRow struct {
	isInitial bool
	id        int64
	year      int
	month     int
	isVenda   bool
	amount    float64 // sempre positivo
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Transaccións dun activo ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
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

	rows := buildRows(chosen, txs)
	renderTable(w, chosen, rows)
	return nil
}

func buildRows(asset domain.Asset, txs []domain.Transaction) []displayRow {
	rows := []displayRow{
		// Compra inicial: o activo créase mercando por primeira vez.
		{
			isInitial: true,
			year:      asset.Year,
			month:     asset.Month,
			isVenda:   false,
			amount:    asset.AmountUSD,
		},
	}
	for _, tx := range txs {
		amount := tx.AmountUSD
		isVenda := false
		if amount < 0 {
			amount = -amount
			isVenda = true
		}
		rows = append(rows, displayRow{
			id:      tx.ID,
			year:    tx.Year,
			month:   tx.Month,
			isVenda: isVenda,
			amount:  amount,
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].year != rows[j].year {
			return rows[i].year < rows[j].year
		}
		if rows[i].month != rows[j].month {
			return rows[i].month < rows[j].month
		}
		// Mesma data: a compra inicial primeiro, despois por id.
		if rows[i].isInitial != rows[j].isInitial {
			return rows[i].isInitial
		}
		return rows[i].id < rows[j].id
	})
	return rows
}

func renderTable(w io.Writer, asset domain.Asset, rows []displayRow) {
	var totalCompra, totalVenda float64
	for _, r := range rows {
		if r.isVenda {
			totalVenda += r.amount
		} else {
			totalCompra += r.amount
		}
	}
	neto := totalCompra - totalVenda

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Transaccións de %s — %s\n", asset.Type.Display(), asset.Name)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Total: %d entradas (incluíndo a compra inicial)\n", len(rows))
	fmt.Fprintf(w, "  Compras: %.2f USD · Vendas: %.2f USD · Neto: %+.2f USD\n",
		totalCompra, totalVenda, neto)
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "  ID\tAno\tMes\tCompra/Venda\tCantidade\t")
	for _, row := range rows {
		idStr := fmt.Sprintf("%d", row.id)
		if row.isInitial {
			idStr = "—"
		}
		typeLabel := "COMPRA"
		if row.isVenda {
			typeLabel = "VENDA"
		}
		fmt.Fprintf(tw, "  %s\t%d\t%d\t%s\t%.2f USD\t\n",
			idStr, row.year, row.month, typeLabel, row.amount,
		)
	}
	tw.Flush()
	fmt.Fprintln(w, sep)
}
