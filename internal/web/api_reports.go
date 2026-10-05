package web

import (
	"net/http"

	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewreport"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
	"invest-tracker/internal/viewtypehistory"
	"invest-tracker/internal/viewtypereport"
)

// registerReportRoutes rexistra as rutas de informes da API.
func (s *Server) registerReportRoutes() {
	s.handleAPI("GET /api/reports/asset/{id}/month", s.reportAssetMonth)
	s.handleAPI("GET /api/reports/type/{type}/month", s.reportTypeMonth)
	s.handleAPI("GET /api/reports/total/month", s.reportTotalMonth)
	s.handleAPI("GET /api/reports/asset/{id}/history", s.reportAssetHistory)
	s.handleAPI("GET /api/reports/type/{type}/history", s.reportTypeHistory)
	s.handleAPI("GET /api/reports/total/history", s.reportTotalHistory)
}

func (s *Server) reportAssetMonth(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	report, err := viewreport.Build(s.store, asset, period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) reportTypeMonth(w http.ResponseWriter, r *http.Request) {
	typ, ok := pathAssetType(w, r)
	if !ok {
		return
	}
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	report, err := viewtypereport.Build(s.store, typ, period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, typeReportDTO(report))
}

func (s *Server) reportTotalMonth(w http.ResponseWriter, r *http.Request) {
	period, fields := queryYearMonth(r)
	if len(fields) > 0 {
		writeFieldErrors(w, fields)
		return
	}
	report, err := viewtotalreport.Build(s.store, period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) reportAssetHistory(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	history, err := viewassethistory.Build(s.store, asset)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, assetHistoryDTO(history))
}

func (s *Server) reportTypeHistory(w http.ResponseWriter, r *http.Request) {
	typ, ok := pathAssetType(w, r)
	if !ok {
		return
	}
	history, err := viewtypehistory.Build(s.store, typ)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, typeHistoryDTO(history))
}

func (s *Server) reportTotalHistory(w http.ResponseWriter, r *http.Request) {
	history, err := viewtotalhistory.Build(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, totalHistoryDTO(history))
}
