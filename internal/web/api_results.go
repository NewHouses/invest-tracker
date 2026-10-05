package web

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
)

type resultRequest struct {
	AssetID   int64   `json:"assetId"`
	ResultUSD float64 `json:"resultUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

type closeMonthRequest struct {
	Month int                     `json:"month"`
	Year  int                     `json:"year"`
	Items []closeMonthRequestItem `json:"items"`
}

type closeMonthRequestItem struct {
	AssetID   int64   `json:"assetId"`
	ResultUSD float64 `json:"resultUsd"`
}

type eligibleResultsResponse struct {
	Period domain.YearMonth           `json:"period"`
	Items  []closemonth.EligibleAsset `json:"items"`
}

type deletedResponse struct {
	Deleted int64 `json:"deleted"`
}

// registerResultRoutes rexistra as rutas de resultados da API.
func (s *Server) registerResultRoutes() {
	s.handleAPI("GET /api/results/eligible", s.listEligibleResults)
	s.handleAPI("POST /api/results", s.createResult)
	s.handleAPI("POST /api/results/close-month", s.closeMonthResults)
	s.handleAPI("GET /api/assets/{id}/results", s.listAssetResults)
	s.handleAPI("DELETE /api/results/{id}", s.deleteResult)
	s.handleAPI("DELETE /api/results", s.deleteResultsByMonth)
}

func (s *Server) listEligibleResults(w http.ResponseWriter, r *http.Request) {
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	items, err := closemonth.Eligible(s.store, period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, eligibleResultsResponse{Period: period, Items: nonNil(items)})
}

func (s *Server) createResult(w http.ResponseWriter, r *http.Request) {
	var req resultRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	validateAmount(fields, "resultUsd", req.ResultUSD)
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	_, assetOK, failed := s.bodyAsset(w, r, fields, "assetId", req.AssetID)
	if failed {
		return
	}
	if assetOK && (domain.YearMonth{Year: req.Year, Month: req.Month}).Valid() {
		if ok, err := s.resultAssetEligible(req.AssetID, req.Year, req.Month); err != nil {
			s.internalError(w, r, err)
			return
		} else if !ok {
			fields["assetId"] = "o ativo non ten capital investido nese mes"
		}
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	result := domain.MonthlyResult{AssetID: req.AssetID, ResultUSD: req.ResultUSD, Year: req.Year, Month: req.Month}
	id, err := s.store.InsertMonthlyResult(result)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	result.ID = id
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) closeMonthResults(w http.ResponseWriter, r *http.Request) {
	var req closeMonthRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := map[string]string{}
	validateYearMonth(fields, "year", "month", req.Year, req.Month)
	eligible := map[int64]bool{}
	periodValid := domain.YearMonth{Year: req.Year, Month: req.Month}.Valid()
	if periodValid {
		items, err := closemonth.Eligible(s.store, req.Year, req.Month)
		if err != nil {
			s.internalError(w, r, err)
			return
		}
		for _, item := range items {
			eligible[item.Asset.ID] = true
		}
	}
	seen := map[int64]bool{}
	results := make([]domain.MonthlyResult, 0, len(req.Items))
	for i, item := range req.Items {
		prefix := fmt.Sprintf("items[%d].", i)
		validateAmount(fields, prefix+"resultUsd", item.ResultUSD)
		if _, ok, failed := s.bodyAsset(w, r, fields, prefix+"assetId", item.AssetID); failed {
			return
		} else if ok && periodValid && !eligible[item.AssetID] {
			fields[prefix+"assetId"] = "o ativo non ten capital investido nese mes"
		}
		if seen[item.AssetID] {
			fields[prefix+"assetId"] = "o ativo está repetido"
		}
		seen[item.AssetID] = true
		results = append(results, domain.MonthlyResult{AssetID: item.AssetID, ResultUSD: item.ResultUSD, Year: req.Year, Month: req.Month})
	}
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	ids, err := s.store.InsertMonthlyResults(results)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, idsResponse{IDs: nonNil(ids)})
}

func (s *Server) listAssetResults(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	results, err := s.store.ListMonthlyResultsByAsset(asset.ID)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, nonNil(results))
}

func (s *Server) deleteResult(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.DeleteMonthlyResult(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "resultado non atopado")
			return
		}
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) deleteResultsByMonth(w http.ResponseWriter, r *http.Request) {
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	deleted, err := s.store.DeleteMonthlyResultsByMonth(period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, deletedResponse{Deleted: deleted})
}

func (s *Server) resultAssetEligible(assetID int64, year, month int) (bool, error) {
	items, err := closemonth.Eligible(s.store, year, month)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.Asset.ID == assetID {
			return true, nil
		}
	}
	return false, nil
}
