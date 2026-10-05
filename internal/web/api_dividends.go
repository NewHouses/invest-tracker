package web

import (
	"database/sql"
	"errors"
	"net/http"

	"invest-tracker/internal/domain"
)

type dividendRequest struct {
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

// registerDividendRoutes rexistra as rutas de dividendos da API.
func (s *Server) registerDividendRoutes() {
	s.handleAPI("GET /api/dividends", s.listDividends)
	s.handleAPI("POST /api/dividends", s.createDividend)
	s.handleAPI("DELETE /api/dividends/{id}", s.deleteDividend)
}

func (s *Server) listDividends(w http.ResponseWriter, r *http.Request) {
	dividends, err := s.store.ListDividends()
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, nonNil(dividends))
}

func (s *Server) createDividend(w http.ResponseWriter, r *http.Request) {
	var req dividendRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	validateAmount(fields, "amountUsd", req.AmountUSD)
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	dividend := domain.Dividend{AmountUSD: req.AmountUSD, Month: req.Month, Year: req.Year}
	id, err := s.store.InsertDividend(dividend)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	dividend.ID = id
	writeJSON(w, http.StatusCreated, dividend)
}

func (s *Server) deleteDividend(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.DeleteDividend(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "dividendo non atopado")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
