package editasset

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	UpdateAsset(domain.Asset) error
}

const (
	fieldName   = 1
	fieldDate   = 2
	fieldAmount = 3
)

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Editar activo ---\n")

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

	fmt.Fprintf(w, "Activo seleccionado: %s — %s (%.2f USD, %02d/%d)\n",
		chosen.Type.Display(), chosen.Name, chosen.AmountUSD, chosen.Month, chosen.Year)

	field, err := promptFieldChoice(r, w)
	if err != nil {
		return err
	}

	updated := chosen
	switch field {
	case fieldName:
		name, err := promptNewName(r, w)
		if err != nil {
			return err
		}
		updated.Name = name
	case fieldDate:
		m, err := prompts.Month(r, w)
		if err != nil {
			return err
		}
		y, err := prompts.Year(r, w)
		if err != nil {
			return err
		}
		updated.Month = m
		updated.Year = y
	case fieldAmount:
		amount, err := prompts.Amount(r, w)
		if err != nil {
			return err
		}
		updated.AmountUSD = amount
	}

	if err := repo.UpdateAsset(updated); err != nil {
		return fmt.Errorf("actualizando activo: %w", err)
	}

	fmt.Fprintf(w, "✓ Activo #%d actualizado: %s — %s (%.2f USD, %02d/%d)\n",
		updated.ID, updated.Type.Display(), updated.Name, updated.AmountUSD,
		updated.Month, updated.Year)
	return nil
}

func promptFieldChoice(r *bufio.Reader, w io.Writer) (int, error) {
	fmt.Fprintln(w, "Que campo queres editar?")
	fmt.Fprintln(w, "  [1] Nome")
	fmt.Fprintln(w, "  [2] Data (mes/ano)")
	fmt.Fprintln(w, "  [3] Cantidade (USD)")
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
		fmt.Fprintln(w, "⚠ Selección non válida, escolle 1-3")
	}
}

func promptNewName(r *bufio.Reader, w io.Writer) (string, error) {
	for {
		fmt.Fprint(w, "Novo nome: ")
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
