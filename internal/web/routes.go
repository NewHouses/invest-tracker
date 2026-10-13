package web

import "net/http"

func (s *Server) routes() {
	s.registerMetaRoutes()
	s.registerAssetRoutes()
	s.registerTransactionRoutes()
	s.registerResultRoutes()
	s.registerDividendRoutes()
	s.registerReportRoutes()
	s.registerChartRoutes()
	s.registerToolRoutes()
	s.registerDashboardRoutes()
	s.mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, langOf(r).T("recurso non atopado"))
	}))
	s.mux.Handle("/", s.staticHandler())
}

func (s *Server) handleAPI(pattern string, h http.HandlerFunc) {
	s.mux.HandleFunc(pattern, h)
}
