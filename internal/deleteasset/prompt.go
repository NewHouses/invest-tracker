package deleteasset

import (
	"bufio"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	DeleteAsset(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	i18n.Fprint(w, "\n--- Eliminar activo ---\n")

	assets, err := repo.ListAssets()
	if err != nil {
		return i18n.Errorf("listando activos: %w", err)
	}
	if len(assets) == 0 {
		i18n.Fprintln(w, "Aínda non hai activos. Engade un primeiro coa operación 'Engadir activo'.")
		return nil
	}

	chosen, err := prompts.SelectAsset(r, w, assets)
	if err != nil {
		return err
	}

	if err := repo.DeleteAsset(chosen.ID); err != nil {
		return i18n.Errorf("eliminando activo: %w", err)
	}

	i18n.Fprintf(w, "✓ Activo #%d eliminado: %s — %s (e todas as súas transaccións e resultados mensuais)\n",
		chosen.ID, chosen.Type.Display(), chosen.Name)
	return nil
}
