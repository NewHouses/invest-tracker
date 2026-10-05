package repartoaporte

import (
	"bufio"
	"fmt"
	"io"
	"math"
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

type TypeSelection struct {
	Type   domain.AssetType `json:"type"`
	Assets []domain.Asset   `json:"assets"`
}

type Allocation struct {
	Total float64          `json:"total"`
	Types []TypeAllocation `json:"types"`
}

type TypeAllocation struct {
	Type   domain.AssetType  `json:"type"`
	Label  string            `json:"label"`
	Amount float64           `json:"amount"`
	Assets []AssetAllocation `json:"assets"`
}

type AssetAllocation struct {
	Asset  domain.Asset `json:"asset"`
	Amount float64      `json:"amount"`
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

	availTypes := AvailableTypes(assets)

	selectedTypes, err := promptSelectTypes(r, w, availTypes)
	if err != nil {
		return err
	}

	amountPerType := total / float64(len(selectedTypes))
	selection := make([]TypeSelection, 0, len(selectedTypes))
	for _, t := range selectedTypes {
		ofType := assetsByType(assets, t)
		fmt.Fprintf(w, "\n→ %.2f USD a %s\n", amountPerType, t.Display())
		selectedAssets, err := promptSelectAssets(r, w, ofType)
		if err != nil {
			return err
		}
		selection = append(selection, TypeSelection{Type: t, Assets: selectedAssets})
	}

	allocation, err := Allocate(total, selection)
	if err != nil {
		return err
	}
	renderReport(w, allocation)
	return nil
}

func AvailableTypes(assets []domain.Asset) []domain.AssetType {
	present := make(map[domain.AssetType]bool)
	for _, a := range assets {
		present[a.Type] = true
	}
	allTypes := []domain.AssetType{domain.Accion, domain.Indice, domain.CopyTrading, domain.Fondo}
	out := make([]domain.AssetType, 0, len(allTypes))
	for _, t := range allTypes {
		if present[t] {
			out = append(out, t)
		}
	}
	return out
}

func Allocate(total float64, selection []TypeSelection) (Allocation, error) {
	if total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return Allocation{}, fmt.Errorf("a cantidade total debe ser maior ca 0")
	}
	if len(selection) == 0 {
		return Allocation{}, fmt.Errorf("debes seleccionar polo menos un tipo")
	}

	amountPerType := total / float64(len(selection))
	alloc := Allocation{Total: total, Types: make([]TypeAllocation, 0, len(selection))}
	for _, sel := range selection {
		if len(sel.Assets) == 0 {
			return Allocation{}, fmt.Errorf("o tipo %s non ten activos seleccionados", sel.Type.Display())
		}
		amountPerAsset := amountPerType / float64(len(sel.Assets))
		typeAlloc := TypeAllocation{
			Type:   sel.Type,
			Label:  sel.Type.Display(),
			Amount: amountPerType,
			Assets: make([]AssetAllocation, 0, len(sel.Assets)),
		}
		for _, a := range sel.Assets {
			typeAlloc.Assets = append(typeAlloc.Assets, AssetAllocation{Asset: a, Amount: amountPerAsset})
		}
		alloc.Types = append(alloc.Types, typeAlloc)
	}
	return alloc, nil
}

func assetsByType(assets []domain.Asset, typ domain.AssetType) []domain.Asset {
	var out []domain.Asset
	for _, a := range assets {
		if a.Type == typ {
			out = append(out, a)
		}
	}
	return out
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

func renderReport(w io.Writer, allocation Allocation) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Reparto de aporte mensual: %.2f USD\n", allocation.Total)
	fmt.Fprintln(w, sep)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range allocation.Types {
		fmt.Fprintf(tw, "  %s\t%.2f USD\n", a.Label, a.Amount)
		for _, asset := range a.Assets {
			fmt.Fprintf(tw, "    %s\t%.2f USD\n", asset.Asset.Name, asset.Amount)
		}
	}
	tw.Flush()
	fmt.Fprintln(w, sep)
}
