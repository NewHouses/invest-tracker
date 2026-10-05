package web

import (
	"net/http"

	"invest-tracker/internal/domain"
	portfoliobuilder "invest-tracker/internal/portfolio"
)

func (s *Server) portfolio(w http.ResponseWriter, r *http.Request) {
	now := s.now()
	period := domain.YearMonth{Year: now.Year(), Month: int(now.Month())}
	port, err := portfoliobuilder.Build(s.store, period)
	if err != nil {
		s.internalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, port)
}
