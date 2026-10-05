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

type Distribution struct {
	Items     []DistributionItem `json:"items"`
	Total     float64            `json:"total"`
	HasAssets bool               `json:"hasAssets"`
}

type DistributionItem struct {
	Asset domain.Asset `json:"asset"`
	Label string       `json:"label"`
	Value float64      `json:"value"`
	Pct   float64      `json:"pct"`

	colorIndex int
}

type LineChart struct {
	Title     string             `json:"title"`
	Months    []domain.YearMonth `json:"months"`
	Series    []charts.Series    `json:"series"`
	HasAssets bool               `json:"hasAssets"`
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

func BuildDistribution(repo Repo) (Distribution, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return Distribution{}, fmt.Errorf("listando activos: %w", err)
	}
	dist := Distribution{HasAssets: len(assets) > 0}
	for i, a := range assets {
		sum, err := repo.MonthlySummary(a.ID, 9999, 12)
		if err != nil {
			return Distribution{}, fmt.Errorf("calculando lifetime de %s: %w", a.Name, err)
		}
		if sum.TotalInvestedUpTo <= 0 {
			continue
		}
		dist.Total += sum.TotalInvestedUpTo
		dist.Items = append(dist.Items, DistributionItem{
			Asset:      a,
			Label:      fmt.Sprintf("%s — %s", a.Type.Display(), a.Name),
			Value:      sum.TotalInvestedUpTo,
			colorIndex: i,
		})
	}
	if dist.Total > 0 {
		for i := range dist.Items {
			dist.Items[i].Pct = dist.Items[i].Value / dist.Total * 100
		}
	}
	return dist, nil
}

func BuildAssetEvolution(repo Repo, asset domain.Asset) (LineChart, error) {
	months, err := repo.MonthsWithResultsForAsset(asset.ID)
	if err != nil {
		return LineChart{}, fmt.Errorf("obtendo meses: %w", err)
	}
	chart := LineChart{Title: fmt.Sprintf("%s — %s", asset.Type.Display(), asset.Name), Months: months, HasAssets: true}
	if len(months) == 0 {
		return chart, nil
	}
	results := make([]float64, len(months))
	aportes := make([]float64, len(months))
	for i, ym := range months {
		sum, err := repo.MonthlySummary(asset.ID, ym.Year, ym.Month)
		if err != nil {
			return LineChart{}, fmt.Errorf("calculando resumo: %w", err)
		}
		if sum.HasResult {
			results[i] = sum.Result
		} else {
			results[i] = math.NaN()
		}
		aportes[i] = sum.TotalInvestedUpTo
	}
	chart.Series = []charts.Series{
		{Label: "Resultado", Values: results},
		{Label: "Aporte acumulado", Values: aportes},
	}
	return chart, nil
}

func BuildTypeAggregated(repo Repo, typ domain.AssetType) (LineChart, error) {
	ofType, err := assetsOfType(repo, typ)
	if err != nil {
		return LineChart{}, err
	}
	chart := LineChart{Title: fmt.Sprintf("Tipo: %s", typ.Display()), HasAssets: len(ofType) > 0}
	if len(ofType) == 0 {
		return chart, nil
	}
	months, err := unionMonths(repo, ofType)
	if err != nil {
		return LineChart{}, err
	}
	chart.Months = months
	if len(months) == 0 {
		return chart, nil
	}
	results := make([]float64, len(months))
	aportes := make([]float64, len(months))
	for i, ym := range months {
		var resSum, aporteSum float64
		for _, a := range ofType {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return LineChart{}, fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				resSum += sum.Result
			}
			aporteSum += sum.TotalInvestedUpTo
		}
		results[i] = resSum
		aportes[i] = aporteSum
	}
	chart.Series = []charts.Series{
		{Label: "Resultado agregado", Values: results},
		{Label: "Aporte acumulado", Values: aportes},
	}
	return chart, nil
}

func BuildAssetsOfType(repo Repo, typ domain.AssetType) (LineChart, error) {
	ofType, err := assetsOfType(repo, typ)
	if err != nil {
		return LineChart{}, err
	}
	chart := LineChart{Title: fmt.Sprintf("Activos de tipo %s", typ.Display()), HasAssets: len(ofType) > 0}
	if len(ofType) == 0 {
		return chart, nil
	}
	months, err := unionMonths(repo, ofType)
	if err != nil {
		return LineChart{}, err
	}
	chart.Months = months
	if len(months) == 0 {
		return chart, nil
	}
	return buildAssetSeries(repo, chart, ofType, months)
}

func BuildAllAssets(repo Repo) (LineChart, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return LineChart{}, fmt.Errorf("listando activos: %w", err)
	}
	chart := LineChart{Title: "Todos os ativos", HasAssets: len(assets) > 0}
	if len(assets) == 0 {
		return chart, nil
	}
	months, err := repo.MonthsWithResults()
	if err != nil {
		return LineChart{}, fmt.Errorf("obtendo meses: %w", err)
	}
	chart.Months = months
	if len(months) == 0 {
		return chart, nil
	}
	return buildAssetSeries(repo, chart, assets, months)
}

func buildAssetSeries(repo Repo, chart LineChart, assets []domain.Asset, months []domain.YearMonth) (LineChart, error) {
	for _, a := range assets {
		values := make([]float64, len(months))
		any := false
		for i, ym := range months {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return LineChart{}, err
			}
			if sum.HasResult {
				values[i] = sum.Result
				any = true
			} else {
				values[i] = math.NaN()
			}
		}
		if any {
			chart.Series = append(chart.Series, charts.Series{Label: a.Name, Values: values})
		}
	}
	return chart, nil
}

func BuildAllTypes(repo Repo) (LineChart, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return LineChart{}, fmt.Errorf("listando activos: %w", err)
	}
	chart := LineChart{Title: "Resultado por tipo", HasAssets: len(assets) > 0}
	if len(assets) == 0 {
		return chart, nil
	}
	months, err := repo.MonthsWithResults()
	if err != nil {
		return LineChart{}, fmt.Errorf("obtendo meses: %w", err)
	}
	chart.Months = months
	if len(months) == 0 {
		return chart, nil
	}
	allTypes := []domain.AssetType{domain.Accion, domain.Indice, domain.CopyTrading, domain.Fondo}
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
					return LineChart{}, err
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
		if any {
			chart.Series = append(chart.Series, charts.Series{Label: t.Display(), Values: values})
		}
	}
	return chart, nil
}

func BuildTotal(repo Repo) (LineChart, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return LineChart{}, fmt.Errorf("listando activos: %w", err)
	}
	chart := LineChart{Title: "Resultado total da carteira", HasAssets: len(assets) > 0}
	if len(assets) == 0 {
		return chart, nil
	}
	months, err := repo.MonthsWithResults()
	if err != nil {
		return LineChart{}, fmt.Errorf("obtendo meses: %w", err)
	}
	chart.Months = months
	if len(months) == 0 {
		return chart, nil
	}
	values := make([]float64, len(months))
	aportes := make([]float64, len(months))
	var cumDiv float64
	for i, ym := range months {
		var resSum, aporteSum float64
		for _, a := range assets {
			sum, err := repo.MonthlySummary(a.ID, ym.Year, ym.Month)
			if err != nil {
				return LineChart{}, fmt.Errorf("calculando resumo de %s: %w", a.Name, err)
			}
			if sum.HasResult {
				resSum += sum.Result
			}
			aporteSum += sum.TotalInvestedUpTo
		}
		div, err := repo.SumDividends(ym.Year, ym.Month)
		if err != nil {
			return LineChart{}, fmt.Errorf("sumando dividendos: %w", err)
		}
		cumDiv += div
		values[i] = resSum + cumDiv
		aportes[i] = aporteSum
	}
	chart.Series = []charts.Series{
		{Label: "Resultado + dividendos acum.", Values: values},
		{Label: "Aporte acumulado", Values: aportes},
	}
	return chart, nil
}

func chart1Distribution(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Distribución de aportes ---")
	dist, err := BuildDistribution(repo)
	if err != nil {
		return err
	}
	if !dist.HasAssets {
		fmt.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}
	if dist.Total <= 0 {
		fmt.Fprintln(w, "Aínda non hai aportes rexistrados.")
		return nil
	}
	items := make([]charts.BarItem, 0, len(dist.Items))
	for _, item := range dist.Items {
		items = append(items, charts.BarItem{Label: item.Label, Value: item.Value, Color: charts.PaletteFG(item.colorIndex)})
	}
	fmt.Fprintf(w, "\nTotal aportado: %s\n\n", money.USD(dist.Total))
	charts.RenderBars(w, items, dist.Total)
	return nil
}

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
	chart, err := BuildAssetEvolution(repo, chosen)
	if err != nil {
		return err
	}
	if len(chart.Months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para este activo.")
		return nil
	}
	charts.RenderLine(w, chart.Series, chart.Months, chart.Title)
	return nil
}

func chart3TypeAggregated(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dun tipo (agregada) ---")
	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}
	chart, err := BuildTypeAggregated(repo, typ)
	if err != nil {
		return err
	}
	if !chart.HasAssets {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}
	if len(chart.Months) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n", typ.Display())
		return nil
	}
	charts.RenderLine(w, chart.Series, chart.Months, chart.Title)
	return nil
}

func chart4AssetsOfType(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dos activos dun tipo ---")
	typ, err := prompts.SelectAssetType(r, w)
	if err != nil {
		return err
	}
	chart, err := BuildAssetsOfType(repo, typ)
	if err != nil {
		return err
	}
	if !chart.HasAssets {
		fmt.Fprintf(w, "Non hai activos de tipo %s.\n", typ.Display())
		return nil
	}
	if len(chart.Months) == 0 || len(chart.Series) == 0 {
		fmt.Fprintf(w, "Aínda non hai resultados rexistrados para activos de tipo %s.\n", typ.Display())
		return nil
	}
	charts.RenderLine(w, chart.Series, chart.Months, chart.Title)
	return nil
}

func chart5AllTypes(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución dos tipos ---")
	chart, err := BuildAllTypes(repo)
	if err != nil {
		return err
	}
	if !chart.HasAssets {
		fmt.Fprintln(w, "Aínda non hai activos.")
		return nil
	}
	if len(chart.Months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados.")
		return nil
	}
	if len(chart.Series) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados para ningún tipo.")
		return nil
	}
	charts.RenderLine(w, chart.Series, chart.Months, chart.Title)
	return nil
}

func chart6Total(w io.Writer, repo Repo) error {
	fmt.Fprintln(w, "\n--- Evolución do resultado total ---")
	chart, err := BuildTotal(repo)
	if err != nil {
		return err
	}
	if !chart.HasAssets {
		fmt.Fprintln(w, "Aínda non hai activos.")
		return nil
	}
	if len(chart.Months) == 0 {
		fmt.Fprintln(w, "Aínda non hai resultados rexistrados.")
		return nil
	}
	charts.RenderLine(w, chart.Series[:1], chart.Months, chart.Title)
	return nil
}

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
