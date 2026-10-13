package repartoaporte

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
}

const sep = "==================================================================="

type TypeSelection struct {
	Type   domain.AssetType `json:"type"`
	Assets []domain.Asset   `json:"assets"`
	// Percents é a porcentaxe do total que leva cada activo, na mesma orde
	// ca Assets. Se ningún tipo a indica, úsase o reparto a partes iguais de
	// EqualPercents.
	Percents []float64 `json:"percents,omitempty"`
}

type Allocation struct {
	Total float64          `json:"total"`
	Types []TypeAllocation `json:"types"`
}

type TypeAllocation struct {
	Type    domain.AssetType  `json:"type"`
	Label   string            `json:"label"`
	Percent float64           `json:"percent"`
	Amount  float64           `json:"amount"`
	Assets  []AssetAllocation `json:"assets"`
}

type AssetAllocation struct {
	Asset   domain.Asset `json:"asset"`
	Percent float64      `json:"percent"`
	Amount  float64      `json:"amount"`
	// New indica un activo que aínda non se creou (vai con ID 0): só se usa
	// para calcular o reparto da primeira vez que se inclúe.
	New bool `json:"new"`
}

// roundingPerPercent é o que pode desviarse da porcentaxe real cada
// porcentaxe escrita con dúas decimais (33,33 en vez de 33,333...).
const roundingPerPercent = 0.005

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Repartir aporte mensual ---\n")

	total, err := prompts.Amount(r, w)
	if err != nil {
		return err
	}

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}

	var selection []TypeSelection
	addNew := true
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos creados: o reparto farase entre activos novos.")
	} else {
		selectedTypes, err := promptSelectTypes(r, w, AvailableTypes(assets))
		if err != nil {
			return err
		}
		for _, t := range selectedTypes {
			ofType := assetsByType(assets, t)
			fmt.Fprintf(w, "\n%s:\n", t.Display())
			selectedAssets, err := promptSelectAssets(r, w, ofType)
			if err != nil {
				return err
			}
			selection = append(selection, TypeSelection{Type: t, Assets: selectedAssets})
		}
		addNew, err = askYesNo(r, w, i18n.T("Queres engadir activos que aínda non creaches? [s/N]: "))
		if err != nil {
			return err
		}
	}
	if addNew {
		if selection, err = promptNewAssets(r, w, selection); err != nil {
			return err
		}
	}

	if err := promptPercents(r, w, selection); err != nil {
		return err
	}
	allocation, err := Allocate(total, selection, i18n.Current())
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

// Allocate reparte total entre os activos seleccionados segundo a
// porcentaxe de cada un (ou a partes iguais, se non se indica ningunha). Os
// activos que aínda non se crearon van con ID 0. Os importes van en centavos
// exactos e suman o total. As etiquetas e os erros van no idioma lang.
func Allocate(total float64, selection []TypeSelection, lang i18n.Lang) (Allocation, error) {
	if total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return Allocation{}, lang.Errorf("a cantidade total debe ser maior ca 0")
	}
	if len(selection) == 0 {
		return Allocation{}, lang.Errorf("debes seleccionar polo menos un tipo")
	}
	for _, sel := range selection {
		if len(sel.Assets) == 0 {
			return Allocation{}, lang.Errorf("o tipo %s non ten activos seleccionados", sel.Type.Label(lang))
		}
	}
	percents, err := selectionPercents(selection, lang)
	if err != nil {
		return Allocation{}, err
	}

	var flat []float64
	sum := 0.0
	for _, ps := range percents {
		flat = append(flat, ps...)
		for _, p := range ps {
			sum += p
		}
	}
	totalCents := int64(math.Round(total * 100))
	cents := splitCents(totalCents, flat)

	alloc := Allocation{Total: float64(totalCents) / 100, Types: make([]TypeAllocation, 0, len(selection))}
	k := 0
	for i, sel := range selection {
		typeAlloc := TypeAllocation{
			Type:   sel.Type,
			Label:  sel.Type.Label(lang),
			Assets: make([]AssetAllocation, 0, len(sel.Assets)),
		}
		var typeCents int64
		for j, a := range sel.Assets {
			// A porcentaxe efectiva: se as indicadas suman 99,99 por
			// redondeo, repártese coma se sumasen 100.
			percent := percents[i][j] * 100 / sum
			typeAlloc.Assets = append(typeAlloc.Assets, AssetAllocation{Asset: a, Percent: roundPercent(percent), Amount: float64(cents[k]) / 100, New: a.ID == 0})
			typeAlloc.Percent += percent
			typeCents += cents[k]
			k++
		}
		typeAlloc.Percent = roundPercent(typeAlloc.Percent)
		typeAlloc.Amount = float64(typeCents) / 100
		alloc.Types = append(alloc.Types, typeAlloc)
	}
	return alloc, nil
}

// EqualPercents propón o reparto de sempre (o total a partes iguais entre os
// tipos e, dentro de cada tipo, entre os seus activos) coas porcentaxes
// redondeadas a dúas decimais, que é como se amosan e se editan.
func EqualPercents(selection []TypeSelection) [][]float64 {
	out := equalShares(selection)
	for i := range out {
		for j := range out[i] {
			out[i][j] = math.Round(out[i][j]*100) / 100
		}
	}
	return out
}

// equalShares é o reparto a partes iguais sen redondear (100/3 = 33,333...),
// o que se usa cando non se indica ningunha porcentaxe.
func equalShares(selection []TypeSelection) [][]float64 {
	out := make([][]float64, len(selection))
	for i, sel := range selection {
		out[i] = make([]float64, len(sel.Assets))
		for j := range sel.Assets {
			out[i][j] = 100 / float64(len(selection)) / float64(len(sel.Assets))
		}
	}
	return out
}

// ValidPercent indica se p vale como porcentaxe dun activo: de 0 a 100.
func ValidPercent(p float64) bool {
	return !math.IsNaN(p) && p >= 0 && p <= 100
}

// SumsTo100 indica se n porcentaxes que suman sum suman 100. Admite o que se
// perde ao escribir cada unha con dúas decimais: 33,33 + 33,33 + 33,33 =
// 99,99 vale, e o reparto faise coma se sumasen 100.
func SumsTo100(sum float64, n int) bool {
	return math.Abs(sum-100) <= roundingPerPercent*float64(n)+1e-9
}

func roundPercent(p float64) float64 {
	return math.Round(p*1e6) / 1e6
}

// selectionPercents devolve a porcentaxe de cada activo e comproba que todas
// son válidas e suman 100. Se non se indicou ningunha, propón o reparto a
// partes iguais.
func selectionPercents(selection []TypeSelection, lang i18n.Lang) ([][]float64, error) {
	given := false
	for _, sel := range selection {
		given = given || sel.Percents != nil
	}
	if !given {
		return equalShares(selection), nil
	}
	out := make([][]float64, len(selection))
	sum, n := 0.0, 0
	for i, sel := range selection {
		if len(sel.Percents) != len(sel.Assets) {
			return nil, lang.Errorf("indica a porcentaxe de cada activo")
		}
		for j, p := range sel.Percents {
			if !ValidPercent(p) {
				return nil, lang.Errorf("a porcentaxe de %s debe estar entre 0 e 100", sel.Assets[j].Name)
			}
			sum += p
			n++
		}
		out[i] = sel.Percents
	}
	if !SumsTo100(sum, n) {
		return nil, lang.Errorf("as porcentaxes suman %.2f %%; teñen que sumar 100 %%", sum)
	}
	return out, nil
}

// splitCents reparte total centavos en proporción ás porcentaxes. Cada parte
// leva os centavos enteiros que lle tocan, e os que sobran van un a un ás
// partes que máis perderon ao redondear (nun empate, á primeira): así a suma
// é exactamente o total e un activo ao 0 % nunca leva nada.
func splitCents(total int64, percents []float64) []int64 {
	out := make([]int64, len(percents))
	sum := 0.0
	for _, p := range percents {
		sum += p
	}
	if sum <= 0 {
		return out
	}
	frac := make([]float64, len(percents))
	var used int64
	for i, p := range percents {
		exact := float64(total) * p / sum
		whole := math.Floor(exact)
		out[i] = int64(whole)
		frac[i] = exact - whole
		used += out[i]
	}
	order := make([]int, len(percents))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return frac[order[a]] > frac[order[b]] })
	for k := 0; used < total; k++ {
		out[order[k%len(order)]]++
		used++
	}
	// Por se o redondeo en coma flotante deu algún centavo de máis.
	for k := len(order) - 1; used > total && k >= 0; k-- {
		if i := order[k]; out[i] > 0 {
			out[i]--
			used--
		}
	}
	return out
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
	i18n.Fprintln(w, "\nTipos dispoñibles:")
	for i, t := range types {
		i18n.Fprintf(w, "  [%d] %s\n", i+1, t.Display())
	}
	for {
		i18n.Fprint(w, "Selecciona os tipos (separados por comas, e.g. 1,3): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return nil, err
		}
		idxs, perr := parseIndices(line, len(types))
		if perr != nil || len(idxs) == 0 {
			i18n.Fprintln(w, "⚠ Selección non válida (escribe un ou máis números válidos separados por comas)")
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
		i18n.Fprintf(w, "  [%d] %s\n", i+1, a.Name)
	}
	for {
		i18n.Fprint(w, "Selecciona os activos (separados por comas, e.g. 1,2): ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return nil, err
		}
		idxs, perr := parseIndices(line, len(ofType))
		if perr != nil || len(idxs) == 0 {
			i18n.Fprintln(w, "⚠ Selección non válida")
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
			return nil, i18n.Errorf("non é un número: %s", f)
		}
		if v < 1 || v > max {
			return nil, i18n.Errorf("fóra de rango: %d", v)
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out, nil
}

// promptPercents amosa o reparto proposto (a partes iguais) e deixa cambiar a
// porcentaxe de cada activo ata que sumen 100. Se se cambian, gárdaas en
// selection.
func promptPercents(r *bufio.Reader, w io.Writer, selection []TypeSelection) error {
	percents := EqualPercents(selection)
	i18n.Fprintln(w, "\nReparto proposto:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, sel := range selection {
		for j, a := range sel.Assets {
			fmt.Fprintf(tw, "  %s — %s\t%6.2f %%\n", sel.Type.Display(), displayName(a), percents[i][j])
		}
	}
	tw.Flush()

	change, err := askYesNo(r, w, i18n.T("Queres cambiar as porcentaxes? [s/N]: "))
	if err != nil || !change {
		// Sen cambios, Allocate fai o reparto a partes iguais exacto.
		return err
	}
	for {
		i18n.Fprintln(w, "Porcentaxe de cada activo (Intro para manter a que ten):")
		sum, n := 0.0, 0
		for i, sel := range selection {
			for j, a := range sel.Assets {
				p, err := promptAssetPercent(r, w, sel.Type, a, percents[i][j])
				if err != nil {
					return err
				}
				percents[i][j] = p
				sum += p
				n++
			}
		}
		if SumsTo100(sum, n) {
			break
		}
		i18n.Fprintf(w, "⚠ As porcentaxes suman %.2f %%; teñen que sumar 100 %%.\n", sum)
	}
	for i := range selection {
		selection[i].Percents = percents[i]
	}
	return nil
}

// askYesNo fai unha pregunta (xa traducida) de resposta si/non; por
// defecto, non.
func askYesNo(r *bufio.Reader, w io.Writer, question string) (bool, error) {
	for {
		fmt.Fprint(w, question)
		line, err := prompts.ReadLine(r)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(line) {
		case "s", "si", "sí", "y", "yes":
			return true, nil
		case "", "n", "non", "no":
			return false, nil
		}
		i18n.Fprintln(w, "⚠ Resposta non válida, escolle S ou N")
	}
}

func promptAssetPercent(r *bufio.Reader, w io.Writer, typ domain.AssetType, a domain.Asset, current float64) (float64, error) {
	for {
		fmt.Fprintf(w, "  %s — %s [%.2f]: ", typ.Display(), displayName(a), current)
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		if line == "" {
			return current, nil
		}
		v, perr := parsePercent(line)
		if perr == nil && ValidPercent(v) {
			return v, nil
		}
		i18n.Fprintln(w, "⚠ Porcentaxe non válida, debe ser un número entre 0 e 100")
	}
}

// promptNewAssets pide activos que aínda non se crearon (tipo e nome) e
// engádeos ao seu tipo dentro da selección: só serven para o cálculo.
func promptNewAssets(r *bufio.Reader, w io.Writer, selection []TypeSelection) ([]TypeSelection, error) {
	for {
		i18n.Fprintln(w, "\nActivo novo (aínda sen crear):")
		typ, err := prompts.SelectAssetType(r, w)
		if err != nil {
			return nil, err
		}
		name, err := promptNewAssetName(r, w)
		if err != nil {
			return nil, err
		}
		selection = addToSelection(selection, domain.Asset{Type: typ, Name: name})
		more, err := askYesNo(r, w, i18n.T("Engadir outro activo novo? [s/N]: "))
		if err != nil {
			return nil, err
		}
		if !more {
			return selection, nil
		}
	}
}

func promptNewAssetName(r *bufio.Reader, w io.Writer) (string, error) {
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

// addToSelection engade un activo ao grupo do seu tipo ou, se o tipo aínda
// non está na selección, nun grupo novo ao final.
func addToSelection(selection []TypeSelection, a domain.Asset) []TypeSelection {
	for i := range selection {
		if selection[i].Type == a.Type {
			selection[i].Assets = append(selection[i].Assets, a)
			return selection
		}
	}
	return append(selection, TypeSelection{Type: a.Type, Assets: []domain.Asset{a}})
}

// displayName é o nome do activo, marcando os que aínda non se crearon.
func displayName(a domain.Asset) string {
	if a.ID == 0 {
		return a.Name + i18n.T(" (novo)")
	}
	return a.Name
}

// parsePercent le unha porcentaxe como "25", "12,5", "12.5" ou "12,5 %".
func parsePercent(s string) (float64, error) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "%"))
	return strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
}

func renderReport(w io.Writer, allocation Allocation) {
	fmt.Fprintln(w)
	i18n.Fprintln(w, sep)
	i18n.Fprintf(w, "  Reparto de aporte mensual: %.2f USD\n", allocation.Total)
	i18n.Fprintln(w, sep)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range allocation.Types {
		fmt.Fprintf(tw, "  %s\t%6.2f %%\t%10.2f USD\n", a.Label, a.Percent, a.Amount)
		for _, asset := range a.Assets {
			fmt.Fprintf(tw, "    %s\t%6.2f %%\t%10.2f USD\n", displayName(asset.Asset), asset.Percent, asset.Amount)
		}
	}
	tw.Flush()
	i18n.Fprintln(w, sep)
}
