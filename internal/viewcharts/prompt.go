package viewcharts

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"

	"invest-tracker/internal/charts"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/money"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	MonthsWithResults() ([]domain.YearMonth, error)
	MonthsWithResultsForAsset(assetID int64) ([]domain.YearMonth, error)
	SumDividends(year, month int) (float64, error)
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Ver gráficas ---\n")

	for {
		choice, err := promptChartType(r, w)
		if err != nil {
			return err
		}
		if choice == 0 {
			return nil
		}

		var dErr error
		switch choice {
		case 1:
			dErr = chart1Distribution(w, repo)
		case 2:
			dErr = chart2AssetEvolution(r, w, repo)
		case 3:
			dErr = chart3TypeAggregated(r, w, repo)
		case 4:
			dErr = chart4AssetsOfType(r, w, repo)
		case 5:
			dErr = chart5AllTypes(w, repo)
		case 6:
			dErr = chart6Total(w, repo)
		}
		if dErr != nil {
			if errors.Is(dErr, prompts.ErrCancelled) {
				fmt.Fprintln(w, "↷ Cancelado.")
				continue
			}
			return dErr
		}
	}
}

func promptChartType(r *bufio.Reader, w io.Writer) (int, error) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  [1] Distribución de aportes")
	fmt.Fprintln(w, "  [2] Evolución dun activo")
	fmt.Fprintln(w, "  [3] Evolución dun tipo (agregada)")
	fmt.Fprintln(w, "  [4] Evolución dos activos dun tipo")
	fmt.Fprintln(w, "  [5] Evolución dos tipos")
	fmt.Fprintln(w, "  [6] Evolución do resultado total")
	fmt.Fprintln(w, "  [0] Voltar")
	for {
		fmt.Fprint(w, "> ")
		line, err := prompts.ReadLine(r)
		if err != nil {
			return 0, err
		}
		n, perr := strconv.Atoi(line)
		if perr == nil && n >= 0 && n <= 6 {
			return n, nil
		}
		fmt.Fprintln(w, "⚠ Selección non válida")
	}
}

// ---- Chart 1: Distribución de aportes ----

func chart1Distribution(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Distribución de aportes ---")
	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}
	items := make([]charts.BarItem, 0, len(assets))
	var total float64
	for i, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		if sum.TotalInvestedUpTo <= 0 {
			continue
		}
		items = append(items, charts.BarItem{
			Label: fmt.Sprintf("%s — %s", a.Type.Display(), a.Name),
			Value: sum.TotalInvestedUpTo,
			Color: charts.PaletteFG(i),
		})
		total += sum.TotalInvestedUpTo
	}
	if total <= 0 {
		fmt.Fprintln(w, "Aínda non hai aportes rexistrados.")
		return nil
	}
	fmt.Fprintf(w, "\nTotal aportado: %s\n\n", money.USD(total))
	charts.RenderBars(w, items, total)
	return nil
}

// ---- Chart 2: Evolución dun activo ----

func chart2AssetEvolution(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dun activo ---")
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
	months, err := repo.MonthsWithResultsForAsset(chosen.ID)
	if err != nil {
		return fmt.Errorf("obtendo meses: %w", err)
	}
	if len(months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para este activo.")
		return nil
	}
	results := make([]float64, len(months))
	aportes := make([]float64, len(months))
	for i, ym := range months {
		sum, err := repo.MonthlySummary(chosen.ID, ym.Year, ym.Month)
		if err != nil {
			return fmt.Errorf("calculando resumo: %w", err)
		}
		if sum.HasResult {
			results[i] = sum.Result
		} else {
			results[i] = math.NaN()
		}
		aportes[i] = sum.TotalInvestedUpTo
	}
	series := []charts.Series{
		{Label: "Resultado", Values: results},
		{Label: "Aporte acumulado", Values: aportes},
	}
	charts.RenderLine(w, series, months, fmt.Sprintf("%s — %s", chosen.Type.Display(), chosen.Name))
	return nil
}

// ---- Chart 3: Evolución dun tipo (agregada) ----

func chart3TypeAggregated(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dun tipo (agregada) ---")
	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}
	ofType, err := assetsOfType(repo, typ)
	if err != nil {
		return err
	}
	if len(ofType) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}
	months, err := unionMonths(repo, ofType)
	if err != nil {
		return err
	}
	if len(months) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n", typ.Display())
		return nil
	}
	results := make([]float64, len(months))
	aportes := make([]float64, len(months))
	for i, ym := range months {
		var resSum, aporteSum float64
		for _, a := range ofType {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				resSum += sum.Result
			}
			aporteSum += sum.TotalInvestedUpTo
		}
		results[i] = resSum
		aportes[i] = aporteSum
	}
	series := []charts.Series{
		{Label: "Resultado agregado", Values: results},
		{Label: "Aporte acumulado", Values: aportes},
	}
	charts.RenderLine(w, series, months, fmt.Sprintf("Tipo: %s", typ.Display()))
	return nil
}

// ---- Chart 4: Evolución dos activos dun tipo ----

func chart4AssetsOfType(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dos activos dun tipo ---")
	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}
	ofType, err := assetsOfType(repo, typ)
	if err != nil {
		return err
	}
	if len(ofType) == 0 {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}
	months, err := unionMonths(repo, ofType)
	if err != nil {
		return err
	}
	if len(months) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n", typ.Display())
		return nil
	}
	series := make([]charts.Series, 0, len(ofType))
	for _, a := range ofType {
		values := make([]float64, len(months))
		any := false
		for i, ym := range months {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return err
			}
			if sum.HasResult {
				values[i] = sum.Result
				any = true
			} else {
				values[i] = math.NaN()
			}
		}
		if !any {
			continue
		}
		series = append(series, charts.Series{Label: a.Name, Values: values})
	}
	if len(series) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n", typ.Display())
		return nil
	}
	charts.RenderLine(w, series, months, fmt.Sprintf("Activos de tipo %s", typ.Display()))
	return nil
}

// ---- Chart 5: Evolución dos tipos ----

func chart5AllTypes(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dos tipos ---")
	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos.")
		return nil
	}
	months, err := repo.MonthsWithResults()
	if err != nil {
		return fmt.Errorf("obtendo meses: %w", err)
	}
	if len(months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados.")
		return nil
	}
	allTypes := []domain.AssetType{
		domain.Accion, domain.Indice, domain.CopyTrading, domain.Fondo,
	}
	series := make([]charts.Series, 0, len(allTypes))
	for _, t := range allTypes {
		var ofType []domain.Asset
		for _, a := range assets {
			if a.Type == t {
				ofType = append(ofType, a)
			}
		}
		if len(ofType) == 0 {
			continue
		}
		values := make([]float64, len(months))
		any := false
		for i, ym := range months {
			var resSum float64
			hasRes := false
			for _, a := range ofType {
				sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
				if err != nil {
					return err
				}
				if sum.HasResult {
					resSum += sum.Result
					hasRes = true
				}
			}
			if hasRes {
				values[i] = resSum
				any = true
			} else {
				values[i] = math.NaN()
			}
		}
		if !any {
			continue
		}
		series = append(series, charts.Series{Label: t.Display(), Values: values})
	}
	if len(series) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para ningún tipo.")
		return nil
	}
	charts.RenderLine(w, series, months, "Resultado por tipo")
	return nil
}

// ---- Chart 6: Evolución do resultado total ----

func chart6Total(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución do resultado total ---")
	assets, err := repo.ListAssets()
	if err != nil {
		return fmt.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		fmt.Fprintln(w, "Aínda non hai activos.")
		return nil
	}
	months, err := repo.MonthsWithResults()
	if err != nil {
		return fmt.Errorf("obtendo meses: %w", err)
	}
	if len(months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados.")
		return nil
	}
	values := make([]float64, len(months))
	var cumDiv float64
	for i, ym := range months {
		var resSum float64
		for _, a := range assets {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				resSum += sum.Result
			}
		}
		div, err := repo.SumDividends(ym.Year, ym.Month)
		if err != nil {
			return fmt.Errorf("sumando dividendos: %w", err)
		}
		cumDiv += div
		values[i] = resSum + cumDiv
	}
	series := []charts.Series{
		{Label: "Resultado + dividendos acum.", Values: values},
	}
	charts.RenderLine(w, series, months, "Resultado total da carteira")
	return nil
}

// ---- helpers ----

func assetsOfType(repo Repo, typ domain.AssetType) ([]domain.Asset, error) {
	all, err := repo.ListAssets()
	if err != nil {
		return nil, fmt.Errorf("listando activos: %w", err)
	}
	var out []domain.Asset
	for _, a := range all {
		if a.Type == typ {
			out = append(out, a)
		}
	}
	return out, nil
}

func unionMonths(repo Repo, assets []domain.Asset) ([]domain.YearMonth, error) {
	set := make(map[domain.YearMonth]bool)
	for _, a := range assets {
		ms, err := repo.MonthsWithResultsForAsset(a.ID)
		if err != nil {
			return nil, fmt.Errorf("obtendo meses de %s: %w", a.Name, err)
		}
		for _, ym := range ms {
			set[ym] = true
		}
	}
	out := make([]domain.YearMonth, 0, len(set))
	for ym := range set {
		out = append(out, ym)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Year != out[j].Year {
			return out[i].Year < out[j].Year
		}
		return out[i].Month < out[j].Month
	})
	return out, nil
}
