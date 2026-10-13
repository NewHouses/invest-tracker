package web

import (
	"database/sql"
	"errors"
	"net/http"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/i18n"
)

// langOf devolve o idioma da petición: a web envía Accept-Language co idioma
// que escolleu o usuario.
func langOf(r *http.Request) i18n.Lang {
	return i18n.FromAcceptLanguage(r.Header.Get("Accept-Language"))
}

// internalError rexistra o erro no log e responde 500 sen expor detalles.
func (s *Server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	lang := langOf(r)
	s.log.Print(i18n.Sprintf("erro en %s %s: %v", r.Method, r.URL.Path, err))
	writeError(w, http.StatusInternalServerError, lang.T("erro interno"))
}

// loadAsset le o {id} da ruta e devolve o activo. Se non é válido ou non
// existe, xa respondeu (400/404/500) e devolve ok=false.
func (s *Server) loadAsset(w http.ResponseWriter, r *http.Request) (domain.Asset, bool) {
	lang := langOf(r)
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return domain.Asset{}, false
	}
	a, err := s.store.GetAsset(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, lang.T("activo non atopado"))
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
	lang := langOf(r)
	t := domain.AssetType(r.PathValue("type"))
	if !t.Valid() {
		writeError(w, http.StatusBadRequest, lang.T("tipo de activo non válido"))
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
