package web

import (
	"net/http"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/portfolio"
	"invest-tracker/internal/viewcharts"
	"invest-tracker/internal/viewtotalhistory"
)

// registerDashboardRoutes rexistra as rutas do panel principal da API.
func (s *Server) registerDashboardRoutes() {
	s.handleAPI("GET /api/dashboard", s.dashboard)
	s.handleAPI("GET /api/portfolio", s.portfolio)
}

type dashboardResponse struct {
	AssetCount   int                     `json:"assetCount"`
	KPIs         dashboardKPIs           `json:"kpis"`
	Evolution    apiLineChart            `json:"evolution"`
	Distribution viewcharts.Distribution `json:"distribution"`
	Portfolio    portfolio.Portfolio     `json:"portfolio"`
}

type dashboardKPIs struct {
	TotalInvested   float64 `json:"totalInvested"`
	CurrentValue    float64 `json:"currentValue"`
	HasCurrentValue bool    `json:"hasCurrentValue"`
	TotalGain       float64 `json:"totalGain"`
	HasTotalGain    bool    `json:"hasTotalGain"`
	TotalDividends  float64 `json:"totalDividends"`
	AvgIndexPct     float64 `json:"avgIndexPct"`
	HasAverages     bool    `json:"hasAverages"`
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	history, err := viewtotalhistory.Build(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	evolution, err := viewcharts.BuildTotal(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	distribution, err := viewcharts.BuildDistribution(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	now := s.now()
	period := domain.YearMonth{Year: now.Year(), Month: int(now.Month())}
	port, err := portfolio.Build(s.store, period)
	if err != nil {
		s.internalError(w, r, err)
		return
	}

	resp := dashboardResponse{
		AssetCount: history.AssetCount,
		KPIs: dashboardKPIs{
			TotalInvested:   history.LifetimeAporte,
			CurrentValue:    history.CurrentValue,
			HasCurrentValue: history.HasCurrentValue,
			TotalGain:       history.TotalGain,
			HasTotalGain:    history.HasTotalGain,
			TotalDividends:  history.TotalDividends,
			AvgIndexPct:     history.AvgIndexPct,
			HasAverages:     history.HasAverages,
		},
		Evolution:    lineChartDTO(evolution),
		Distribution: distributionDTO(distribution),
		Portfolio:    port,
	}
	writeJSON(w, http.StatusOK, resp)
}
