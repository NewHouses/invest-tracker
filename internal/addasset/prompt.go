package addasset

import (
	"bufio"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Saver interface {
	InsertAsset(domain.Asset) (int64, error)
}

func Run(r *bufio.Reader, w io.Writer, saver Saver) error {
	i18n.Fprint(w, "\n--- Engadir novo activo ---\n")

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
		return i18n.Errorf("gardando activo: %w", err)
	}

	i18n.Fprintf(w, "✓ Activo gardado #%d: %s — %s — %.2f USD — %02d/%d\n",
		id, typ.Display(), name, amount, month, year)
	return nil
}

func promptName(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		i18n.Fprint(w, "Nome: ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return "", err
		}
		if line != "" {
			return line, nil
		}
		i18n.Fprintln(w, "⚠ O nome non pode estar baleiro")
	}
}
