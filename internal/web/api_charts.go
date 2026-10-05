package web

import (
	"net/http"

	"invest-tracker/internal/viewcharts"
)

// registerChartRoutes rexistra as rutas de gráficas da API.
func (s *Server) registerChartRoutes() {
	s.handleAPI("GET /api/charts/distribution", s.chartDistribution)
	s.handleAPI("GET /api/charts/asset/{id}", s.chartAsset)
	s.handleAPI("GET /api/charts/type/{type}", s.chartTypeAggregated)
	s.handleAPI("GET /api/charts/type/{type}/assets", s.chartTypeAssets)
	s.handleAPI("GET /api/charts/assets", s.chartAssets)
	s.handleAPI("GET /api/charts/types", s.chartTypes)
	s.handleAPI("GET /api/charts/total", s.chartTotal)
}

func (s *Server) chartDistribution(w http.ResponseWriter, r *http.Request) {
	dist, err := viewcharts.BuildDistribution(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, distributionDTO(dist))
}

func (s *Server) chartAsset(w http.ResponseWriter, r *http.Request) {
	asset, ok := s.loadAsset(w, r)
	if !ok {
		return
	}
	chart, err := viewcharts.BuildAssetEvolution(s.store, asset)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}

func (s *Server) chartTypeAggregated(w http.ResponseWriter, r *http.Request) {
	typ, ok := pathAssetType(w, r)
	if !ok {
		return
	}
	chart, err := viewcharts.BuildTypeAggregated(s.store, typ)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}

func (s *Server) chartTypeAssets(w http.ResponseWriter, r *http.Request) {
	typ, ok := pathAssetType(w, r)
	if !ok {
		return
	}
	chart, err := viewcharts.BuildAssetsOfType(s.store, typ)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}

func (s *Server) chartAssets(w http.ResponseWriter, r *http.Request) {
	chart, err := viewcharts.BuildAllAssets(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}

func (s *Server) chartTypes(w http.ResponseWriter, r *http.Request) {
	chart, err := viewcharts.BuildAllTypes(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}

func (s *Server) chartTotal(w http.ResponseWriter, r *http.Request) {
	chart, err := viewcharts.BuildTotal(s.store)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, lineChartDTO(chart))
}
