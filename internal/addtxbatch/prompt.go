package addtxbatch

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	InsertTransaction(domain.Transaction) (int64, error)
}

const (
	modePerTx         = 1
	modeFixedMonth    = 2
	modeAutoIncrement = 3
)

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Engadir varias transaccións ---\n")

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
	i18n.Fprintf(w, "Sobre %s — %s (creado en %02d/%d)\n",
		chosen.Type.Display(), chosen.Name, chosen.Month, chosen.Year)

	mode, err := promptMode(r, w)
	if err != nil {
		return err
	}

	var fixedMonth, fixedYear int
	if mode == modeFixedMonth || mode == modeAutoIncrement {
		i18n.Fprintln(w, "Introduce a data de inicio:")
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
			i18n.Fprintf(w, "\n--- Transacción #%d ---\n", saved+1)
			txMonth, txYear, err = prompts.DateNotBefore(r, w, chosen.Month, chosen.Year)
			if err != nil {
				return err
			}
		case modeFixedMonth:
			txMonth, txYear = fixedMonth, fixedYear
			i18n.Fprintf(w, "\n--- Transacción #%d en %02d/%d ---\n", saved+1, txMonth, txYear)
		case modeAutoIncrement:
			txMonth, txYear = curMonth, curYear
			i18n.Fprintf(w, "\n--- Transacción #%d en %02d/%d ---\n", saved+1, txMonth, txYear)
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
			return i18n.Errorf("gardando transacción: %w", err)
		}
		i18n.Fprintf(w, "✓ Gardada #%d: %s %.2f USD — %02d/%d\n",
			id, i18n.T(typeLabel), amount, txMonth, txYear)
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

	i18n.Fprintf(w, "\n✓ Engadíronse %d transacción(s) sobre %s — %s.\n",
		saved, chosen.Type.Display(), chosen.Name)
	return nil
}

func promptMode(r *bufio.Reader, w io.Writer) (int, error) {
	i18n.Fprintln(w, "Modo de entrada:")
	i18n.Fprintln(w, "  [1] Establecer mes/ano en cada transacción")
	i18n.Fprintln(w, "  [2] Engadir todas no mesmo mes")
	i18n.Fprintln(w, "  [3] Unha por mes (auto-incrementa) dende un mes inicial")
	for {
		i18n.Fprint(w, "> ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		v, perr := strconv.Atoi(line)
		if perr == nil && v >= 1 && v <= 3 {
			return v, nil
		}
		i18n.Fprintln(w, "⚠ Modo non válido, escolle 1-3")
	}
}

func promptContinue(r *bufio.Reader, w io.Writer) (bool, error) {
	for {
		i18n.Fprint(w, "Engadir outra transacción? [S/n]: ")
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
		i18n.Fprintln(w, "⚠ Resposta non válida, escolle S ou N")
	}
}

func nextMonth(m, y int) (int, int) {
	if m == 12 {
		return 1, y + 1
	}
	return m + 1, y
}
