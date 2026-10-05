package web

import (
	"database/sql"
	"errors"
	"net/http"

	"invest-tracker/internal/domain"
)

// internalError rexistra o erro no log e responde 500 sen expor detalles.
func (s *Server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Printf("erro en %s %s: %v", r.Method, r.URL.Path, err)
	writeError(w, http.StatusInternalServerError, "erro interno")
}

// loadAsset le o {id} da ruta e devolve o ativo. Se non é válido ou non
// existe, xa respondeu (400/404/500) e devolve ok=false.
func (s *Server) loadAsset(w http.ResponseWriter, r *http.Request) (domain.Asset, bool) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return domain.Asset{}, false
	}
	a, err := s.store.GetAsset(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "ativo non atopado")
			return domain.Asset{}, false
		}
		s.internalError(w, r, err)
		return domain.Asset{}, false
	}
	return a, true
}

// pathAssetType le o {type} da ruta. Se non é un tipo válido responde 400 e
// devolve ok=false.
func pathAssetType(w http.ResponseWriter, r *http.Request) (domain.AssetType, bool) {
	t := domain.AssetType(r.PathValue("type"))
	if !t.Valid() {
		writeError(w, http.StatusBadRequest, "tipo de ativo non válido")
		return "", false
	}
	return t, true
}

// nonNil garante que as listas se serialicen como [] e nunca como null.
func nonNil[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
