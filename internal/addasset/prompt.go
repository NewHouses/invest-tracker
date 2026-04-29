package addasset

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Saver interface {
	InsertInvestment(domain.Investment) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, saver Saver) error {
	fmt.Fprint(w, "\n--- Engadir novo activo ---\n")

	typ, err := promptType(r, w)
	if err != nil {
		return err
	}
	name, err := promptName(r, w)
	if err != nil {
		return err
	}
	amount, err := prompts.Amount(r, w)
	if err != nil {
		return err
	}
	month, err := prompts.Month(r, w)
	if err != nil {
		return err
	}
	year, err := prompts.Year(r, w)
	if err != nil {
		return err
	}

	inv := domain.Investment{
		Type:      typ,
		Name:      name,
		AmountUSD: amount,
		Month:     month,
		Year:      year,
	}
	id, err := saver.InsertInvestment(inv)
	if err != nil {
		return fmt.Errorf("gardando investimento: %w", err)
	}

	fmt.Fprintf(w, "✓ Investimento gardado #%d: %s — %s — %.2f USD — %02d/%d\n",
		id, typ.Display(), name, amount, month, year)
	return nil
}

func promptType(r *bufio.Reader, w io.Writer) (domain.InvestmentType, error) {
	fmt.Fprint(w, "Tipo de investimento:\n")
	fmt.Fprint(w, "  [1] Acción\n")
	fmt.Fprint(w, "  [2] Índice\n")
	fmt.Fprint(w, "  [3] Copy-trading\n")
	fmt.Fprint(w, "  [4] Fondo\n")
	for {
		fmt.Fprint(w, "> ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return "", err
		}
		switch line {
		case "1":
			return domain.Accion, nil
		case "2":
			return domain.Indice, nil
		case "3":
			return domain.CopyTrading, nil
		case "4":
			return domain.Fondo, nil
		}
		fmt.Fprintln(w, "⚠ Tipo non válido, escolle 1-4")
	}
}

func promptName(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "Nome: ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return "", err
		}
		if line != "" {
			return line, nil
		}
		fmt.Fprintln(w, "⚠ O nome non pode estar baleiro")
	}
}
