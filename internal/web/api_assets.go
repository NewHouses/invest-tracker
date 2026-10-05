package web

import (
	"database/sql"
	"errors"
	"net/http"

	"invest-tracker/internal/domain"
)

type assetCreateRequest struct {
	Type      domain.AssetType `json:"type"`
	Name      string           `json:"name"`
	AmountUSD float64          `json:"amountUsd"`
	Month     int              `json:"month"`
	Year      int              `json:"year"`
}

type assetUpdateRequest struct {
	Name      string  `json:"name"`
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

// registerAssetRoutes rexistra as rutas de ativos da API.
func (s *Server) registerAssetRoutes() {
	s.handleAPI("GET /api/assets", s.listAssets)
	s.handleAPI("GET /api/assets/{id}", s.getAsset)
	s.handleAPI("POST /api/assets", s.createAsset)
	s.handleAPI("PUT /api/assets/{id}", s.updateAsset)
	s.handleAPI("DELETE /api/assets/{id}", s.deleteAsset)
}

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := s.store.ListAssets()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, nonNil(assets))
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, asset)
}

func (s *Server) createAsset(w http.ResponseWriter, r *http.Request) {
	var req assetCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	if !req.Type.Valid() {
		fields["type"] = "o tipo de ativo non é válido"
	}
	name := validateName(fields, "name", req.Name)
	validateAmount(fields, "amountUsd", req.AmountUSD)
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	asset := domain.Asset{Type: req.Type, Name: name, AmountUSD: req.AmountUSD, Month: req.Month, Year: req.Year}
	id, err := s.store.InsertAsset(asset)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	asset.ID = id
	writeJSON(w, http.StatusCreated, asset)
}

func (s *Server) updateAsset(w http.ResponseWriter, r *http.Request) {
	current, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	var req assetUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	name := validateName(fields, "name", req.Name)
	validateAmount(fields, "amountUsd", req.AmountUSD)
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	updated := domain.Asset{ID: current.ID, Type: current.Type, Name: name, AmountUSD: req.AmountUSD, Month: req.Month, Year: req.Year}
	if err := s.store.UpdateAsset(updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "ativo non atopado")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteAsset(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.DeleteAsset(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "ativo non atopado")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
