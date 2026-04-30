package addtxbatch

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
	ListAssets() ([]domain.Asset, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

const (
	modePerTx        = 1
	modeFixedMonth   = 2
	modeAutoIncrement = 3
)

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Engadir varias transaccións ---\n")

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
	fmt.Fprintf(w, "Sobre %s — %s (creado en %02d/%d)\n",
		chosen.Type.Display(), chosen.Name, chosen.Month, chosen.Year)

	mode, err := promptMode(r, w)
	if err != nil {
		return err
	}

	var fixedMonth, fixedYear int
	if mode == modeFixedMonth || mode == modeAutoIncrement {
		fmt.Fprintln(w, "Introduce a data de inicio:")
		fixedMonth, fixedYear, err = prompts.DateNotBefore(r, w, chosen.Month, chosen.Year)
		if err != nil {
			return err
		}
	}

	saved := 0
	curMonth, curYear := fixedMonth, fixedYear
	for {
		var txMonth, txYear int
		switch mode {
		case modePerTx:
			fmt.Fprintf(w, "\n--- Transacción #%d ---\n", saved+1)
			txMonth, txYear, err = prompts.DateNotBefore(r, w, chosen.Month, chosen.Year)
			if err != nil {
				return err
			}
		case modeFixedMonth:
			txMonth, txYear = fixedMonth, fixedYear
			fmt.Fprintf(w, "\n--- Transacción #%d en %02d/%d ---\n", saved+1, txMonth, txYear)
		case modeAutoIncrement:
			txMonth, txYear = curMonth, curYear
			fmt.Fprintf(w, "\n--- Transacción #%d en %02d/%d ---\n", saved+1, txMonth, txYear)
		}

		isVenda, err := prompts.SelectTransactionType(r, w)
		if err != nil {
			return err
		}
		amount, err := prompts.Amount(r, w)
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
			Month:     txMonth,
			Year:      txYear,
		}
		id, err := repo.InsertTransaction(tx)
		if err != nil {
			return fmt.Errorf("gardando transacción: %w", err)
		}
		fmt.Fprintf(w, "✓ Gardada #%d: %s %.2f USD — %02d/%d\n",
			id, typeLabel, amount, txMonth, txYear)
		saved++

		if mode == modeAutoIncrement {
			curMonth, curYear = nextMonth(curMonth, curYear)
		}

		cont, err := promptContinue(r, w)
		if err != nil {
			return err
		}
		if !cont {
			break
		}
	}

	fmt.Fprintf(w, "\n✓ Engadíronse %d transacción(s) sobre %s — %s.\n",
		saved, chosen.Type.Display(), chosen.Name)
	return nil
}

func promptMode(r *bufio.Reader, w io.Writer) (int, error) {
	fmt.Fprintln(w, "Modo de entrada:")
	fmt.Fprintln(w, "  [1] Establecer mes/ano en cada transacción")
	fmt.Fprintln(w, "  [2] Engadir todas no mesmo mes")
	fmt.Fprintln(w, "  [3] Unha por mes (auto-incrementa) dende un mes inicial")
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
		fmt.Fprintln(w, "⚠ Modo non válido, escolle 1-3")
	}
}

func promptContinue(r *bufio.Reader, w io.Writer) (bool, error) {
	for {
		fmt.Fprint(w, "Engadir outra transacción? [S/n]: ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return false, err
		}
		low := strings.ToLower(line)
		if low == "" || low == "s" || low == "si" || low == "sí" || low == "y" || low == "yes" {
			return true, nil
		}
		if low == "n" || low == "no" || low == "non" {
			return false, nil
		}
		fmt.Fprintln(w, "⚠ Resposta non válida, escolle S ou N")
	}
}

func nextMonth(m, y int) (int, int) {
	if m == 12 {
		return 1, y + 1
	}
	return m + 1, y
}
