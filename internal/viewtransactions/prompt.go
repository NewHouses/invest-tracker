package viewtransactions

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error)
}

const sep = "============================================================"

type Row struct {
	IsInitial bool    `json:"isInitial"`
	ID        int64   `json:"id"`
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	IsVenda   bool    `json:"isVenda"`
	Amount    float64 `json:"amount"`
}

type Totals struct {
	Compra float64 `json:"compra"`
	Venda  float64 `json:"venda"`
	Neto   float64 `json:"neto"`
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

	rows := BuildRows(chosen, txs)
	renderTable(w, chosen, rows)
	return nil
}

func BuildRows(asset domain.Asset, txs []domain.Transaction) []Row {
	rows := []Row{
		{
			IsInitial: true,
			Year:      asset.Year,
			Month:     asset.Month,
			IsVenda:   false,
			Amount:    asset.AmountUSD,
		},
	}
	for _, tx := range txs {
		amount := tx.AmountUSD
		isVenda := false
		if amount < 0 {
			amount = -amount
			isVenda = true
		}
		rows = append(rows, Row{
			ID:      tx.ID,
			Year:    tx.Year,
			Month:   tx.Month,
			IsVenda: isVenda,
			Amount:  amount,
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Year != rows[j].Year {
			return rows[i].Year < rows[j].Year
		}
		if rows[i].Month != rows[j].Month {
			return rows[i].Month < rows[j].Month
		}
		// Mesma data: a compra inicial primeiro, despois por id.
		if rows[i].IsInitial != rows[j].IsInitial {
			return rows[i].IsInitial
		}
		return rows[i].ID < rows[j].ID
	})
	return rows
}

func ComputeTotals(rows []Row) Totals {
	var totals Totals
	for _, r := range rows {
		if r.IsVenda {
			totals.Venda += r.Amount
		} else {
			totals.Compra += r.Amount
		}
	}
	totals.Neto = totals.Compra - totals.Venda
	return totals
}

func renderTable(w io.Writer, asset domain.Asset, rows []Row) {
	totals := ComputeTotals(rows)

	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Transaccións de %s — %s\n", asset.Type.Display(), asset.Name)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Total: %d entradas (incluíndo a compra inicial)\n", len(rows))
	fmt.Fprintf(w, "  Compras: %s · Vendas: %s · Neto: %s\n",
		money.USD(totals.Compra), money.USD(totals.Venda), money.SignedUSD(totals.Neto))
	fmt.Fprintln(w, sep)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "  ID\tAno\tMes\tCompra/Venda\tCantidade\t")
	for _, row := range rows {
		idStr := fmt.Sprintf("%d", row.ID)
		if row.IsInitial {
			idStr = "—"
		}
		typeLabel := "COMPRA"
		if row.IsVenda {
			typeLabel = "VENDA"
		}
		fmt.Fprintf(tw, "  %s\t%d\t%d\t%s\t%s\t\n",
			idStr, row.Year, row.Month, typeLabel, money.USD(row.Amount),
		)
	}
	tw.Flush()
	fmt.Fprintln(w, sep)
}
