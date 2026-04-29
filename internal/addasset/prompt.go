package addasset

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Saver interface {
	InsertAsset(domain.Asset) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, saver Saver) error {
	fmt.Fprint(w, "\n--- Engadir novo activo ---\n")

	typ, err := prompts.SelectAssetType(r, w)
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

	a := domain.Asset{
		Type:      typ,
		Name:      name,
		AmountUSD: amount,
		Month:     month,
		Year:      year,
	}
	id, err := saver.InsertAsset(a)
	if err != nil {
		return fmt.Errorf("gardando activo: %w", err)
	}

	fmt.Fprintf(w, "✓ Activo gardado #%d: %s — %s — %.2f USD — %02d/%d\n",
		id, typ.Display(), name, amount, month, year)
	return nil
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
