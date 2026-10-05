package web

import "net/http"

func (s *Server) routes() {
	s.registerAuthRoutes()
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
		writeError(w, http.StatusNotFound, "recurso non atopado")
	}))
	s.mux.Handle("/", s.staticHandler())
}

func (s *Server) handleAPI(pattern string, h http.HandlerFunc) {
	s.mux.Handle(pattern, s.requireSession(h))
}

func (s *Server) handlePublic(pattern string, h http.HandlerFunc) {
	s.mux.HandleFunc(pattern, h)
}
