package repartoaporte

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
}

const sep = "==================================================================="

type assetAlloc struct {
	name   string
	amount float64
}

type typeAlloc struct {
	label  string
	amount float64
	assets []assetAlloc
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Repartir aporte mensual ---\n")

	total, err := prompts.Amount(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	// Tipos presentes nos activos, en orde estable.
	allTypes := []domain.AssetType{
		domain.Accion, domain.Indice, domain.CopyTrading, domain.Fondo,
	}
	present := make(map[domain.AssetType]bool)
	for _, a := range assets {
		present[a.Type] = true
	}
	var availTypes []domain.AssetType
	for _, t := range allTypes {
		if present[t] {
			availTypes = append(availTypes, t)
		}
	}

	selectedTypes, err := promptSelectTypes(r, w, availTypes)
	if err != nil {
		return err
	}

	amountPerType := total / float64(len(selectedTypes))

	allocs := make([]typeAlloc, 0, len(selectedTypes))
	for _, t := range selectedTypes {
		var ofType []domain.Asset
		for _, a := range assets {
			if a.Type == t {
				ofType = append(ofType, a)
			}
		}
		fmt.Fprintf(w, "\n→ %.2f USD a %s\n", amountPerType, t.Display())
		selectedAssets, err := promptSelectAssets(r, w, ofType)
		if err != nil {
			return err
		}
		amountPerAsset := amountPerType / float64(len(selectedAssets))
		alloc := typeAlloc{label: t.Display(), amount: amountPerType}
		for _, a := range selectedAssets {
			alloc.assets = append(alloc.assets, assetAlloc{name: a.Name, amount: amountPerAsset})
		}
		allocs = append(allocs, alloc)
	}

	renderReport(w, total, allocs)
	return nil
}

func promptSelectTypes(r *bufio.Reader, w io.Writer, types []domain.AssetType) ([]domain.AssetType, error) {
	fmt.Fprintln(w, "\nTipos dispoñibles:")
	for i, t := range types {
		fmt.Fprintf(w, "  [%d] %s\n", i+1, t.Display())
	}
	for {
		fmt.Fprint(w, "Selecciona os tipos (separados por comas, e.g. 1,3): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return nil, err
		}
		idxs, perr := parseIndices(line, len(types))
		if perr != nil || len(idxs) == 0 {
			fmt.Fprintln(w, "⚠ Selección non válida (escribe un ou máis números válidos separados por comas)")
			continue
		}
		out := make([]domain.AssetType, 0, len(idxs))
		for _, i := range idxs {
			out = append(out, types[i-1])
		}
		return out, nil
	}
}

func promptSelectAssets(r *bufio.Reader, w io.Writer, ofType []domain.Asset) ([]domain.Asset, error) {
	for i, a := range ofType {
		fmt.Fprintf(w, "  [%d] %s\n", i+1, a.Name)
	}
	for {
		fmt.Fprint(w, "Selecciona os activos (separados por comas, e.g. 1,2): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return nil, err
		}
		idxs, perr := parseIndices(line, len(ofType))
		if perr != nil || len(idxs) == 0 {
			fmt.Fprintln(w, "⚠ Selección non válida")
			continue
		}
		out := make([]domain.Asset, 0, len(idxs))
		for _, i := range idxs {
			out = append(out, ofType[i-1])
		}
		return out, nil
	}
}

// parseIndices acepta "1,2,3" ou "1 2 3" ou mesturado, devolvendo a lista
// ordenada de entrada e des-duplicada. Erro se algún token non é número
// ou queda fóra do rango [1, max].
func parseIndices(input string, max int) ([]int, error) {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	seen := make(map[int]bool)
	out := make([]int, 0, len(fields))
	for _, f := range fields {
		v, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("non é un número: %s", f)
		}
		if v < 1 || v > max {
			return nil, fmt.Errorf("fóra de rango: %d", v)
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out, nil
}

func renderReport(w io.Writer, total float64, allocs []typeAlloc) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Reparto de aporte mensual: %.2f USD\n", total)
	fmt.Fprintln(w, sep)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range allocs {
		fmt.Fprintf(tw, "  %s\t%.2f USD\n", a.label, a.amount)
		for _, asset := range a.assets {
			fmt.Fprintf(tw, "    %s\t%.2f USD\n", asset.name, asset.amount)
		}
	}
	tw.Flush()
	fmt.Fprintln(w, sep)
}
