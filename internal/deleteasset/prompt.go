package deleteasset

import (
	"bufio"
	"fmt"
	"io"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/prompts"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	DeleteAsset(id int64) error
}

func Run(r *bufio.Reader, w io.Writer, repo Repo) error {
	fmt.Fprint(w, "\n--- Eliminar activo ---\n")

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

	if err := repo.DeleteAsset(chosen.ID); err != nil {
		return fmt.Errorf("eliminando activo: %w", err)
	}

	fmt.Fprintf(w, "✓ Activo #%d eliminado: %s — %s (e todas as súas transaccións e resultados mensuais)\n",
		chosen.ID, chosen.Type.Display(), chosen.Name)
	return nil
}
