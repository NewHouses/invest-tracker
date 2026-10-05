package web

import (
	"net/http"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewcharts"
	"invest-tracker/internal/viewtotalhistory"
)

// registerDashboardRoutes rexistra as rutas do panel principal da API.
func (s *Server) registerDashboardRoutes() {
	s.handleAPI("GET /api/dashboard", s.dashboard)
}

type dashboardResponse struct {
	AssetCount   int                     `json:"assetCount"`
	KPIs         dashboardKPIs           `json:"kpis"`
	Evolution    apiLineChart            `json:"evolution"`
	Distribution viewcharts.Distribution `json:"distribution"`
	Pending      dashboardPending        `json:"pending"`
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

type dashboardPending struct {
	Period domain.YearMonth           `json:"period"`
	Items  []closemonth.EligibleAsset `json:"items"`
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
	pending, err := closemonth.Eligible(s.store, period.Year, period.Month)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	withoutResult := make([]closemonth.EligibleAsset, 0, len(pending))
	for _, item := range pending {
		if !item.HasResult {
			withoutResult = append(withoutResult, item)
		}
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
		Pending: dashboardPending{
			Period: period,
			Items:  eligibleAssetsDTO(withoutResult),
		},
	}
	writeJSON(w, http.StatusOK, resp)
}
