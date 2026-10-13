package deleteresult

import (
	"bufio"
	"io"
	"strconv"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	ListMonthlyResultsByAsset(assetID int64) ([]domain.MonthlyResult, error)
	DeleteMonthlyResult(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Eliminar resultado mensual ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	chosenAsset, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}

	results, err := repo.ListMonthlyResultsByAsset(chosenAsset.ID)
	if err != nil {
		return i18n.Errorf("listando resultados: %w", err)
	}
	if len(results) == 0 {
		i18n.Fprintln(w, "Este activo non ten resultados mensuais rexistrados.")
		return nil
	}

	chosenResult, err := promptResultSelection(r, w, chosenAsset, results)
	if err != nil {
		return err
	}

	if err := repo.DeleteMonthlyResult(chosenResult.ID); err != nil {
		return i18n.Errorf("eliminando resultado: %w", err)
	}

	i18n.Fprintf(w, "✓ Resultado #%d eliminado (%02d/%d — %.2f USD)\n",
		chosenResult.ID, chosenResult.Month, chosenResult.Year, chosenResult.ResultUSD)
	return nil
}

func promptResultSelection(r *bufio.Reader, w io.Writer, asset domain.Asset, results []domain.MonthlyResult) (domain.MonthlyResult, error) {
	i18n.Fprintf(w, "Resultados de %s — %s:\n", asset.Type.Display(), asset.Name)
	for i, mr := range results {
		i18n.Fprintf(w, "  [%d] %02d/%d — %.2f USD\n", i+1, mr.Month, mr.Year, mr.ResultUSD)
	}
	for {
		i18n.Fprintf(w, "Selecciona o resultado a eliminar (1-%d): ", len(results))
		line, err := prompts.ReadLine(r)
		if err != nil {
			return domain.MonthlyResult{}, err
		}
		idx, perr := strconv.Atoi(line)
		if perr == nil && idx >= 1 && idx <= len(results) {
			return results[idx-1], nil
		}
		i18n.Fprintln(w, "⚠ Selección non válida")
	}
}
