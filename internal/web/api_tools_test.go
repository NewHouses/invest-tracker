package web

import (
	"net/http"
	"testing"
	"time"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
)

func authCookieWithServerClock(t *testing.T, s *Server) *http.Cookie {
	t.Helper()
	token := "token-tools-" + t.Name()
	if err := s.store.CreateSession(hashToken(token), s.now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return &http.Cookie{Name: sessionCookieName, Value: token}
}

func TestToolsRequireSession(t *testing.T) {
	s := testServer(t)
	if rr := doJSON(t, s.Handler(), http.MethodGet, "/api/tools/projection/defaults", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("defaults sen sesión: código=%d, esperabamos 401", rr.Code)
	}
	if rr := doJSON(t, s.Handler(), http.MethodPost, "/api/tools/projection", projectionRequest{}, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("projection sen sesión: código=%d, esperabamos 401", rr.Code)
	}
}

func TestAllocationHappyAndValidation(t *testing.T) {
	s := testServer(t)
	a, b := seedSkippedMonth(t, s)
	h := s.Handler()
	cookie := authCookie(t, s)

	body := allocationRequest{Total: 120, Selection: []allocationSelection{{Type: domain.Accion, AssetIDs: []int64{a.ID, b.ID}}}}
	rr := doJSON(t, h, http.MethodPost, "/api/tools/allocation", body, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("allocation: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got repartoaporte.Allocation
	decodeBody(t, rr, &got)
	if got.Total != 120 || len(got.Types) != 1 || got.Types[0].Amount != 120 || len(got.Types[0].Assets) != 2 || got.Types[0].Assets[0].Amount != 60 {
		t.Fatalf("reparto inesperado: %#v", got)
	}

	bad := allocationRequest{Total: -1, Selection: []allocationSelection{{Type: domain.Indice, AssetIDs: []int64{a.ID}}, {Type: "malo"}}}
	rr = doJSON(t, h, http.MethodPost, "/api/tools/allocation", bad, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("allocation inválida: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["total"] == "" || apiErr.Fields["selection[0].assetIds"] == "" || apiErr.Fields["selection[1].type"] == "" || apiErr.Fields["selection[1].assetIds"] == "" {
		t.Fatalf("faltan erros de campos: %#v", apiErr.Fields)
	}

	bad = allocationRequest{Total: 100, Selection: []allocationSelection{{Type: domain.Accion, AssetIDs: []int64{999}}}}
	rr = doJSON(t, h, http.MethodPost, "/api/tools/allocation", bad, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("assetId descoñecido debe ser 400: código=%d", rr.Code)
	}
}

func TestProjectionDefaultsWithoutAssets(t *testing.T) {
	s := fixedClockServer(t, time.Date(2026, 10, 5, 13, 0, 0, 0, time.UTC))
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/tools/projection/defaults", nil, authCookieWithServerClock(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("defaults: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got projectionDefaultsResponse
	decodeBody(t, rr, &got)
	if got.StartFromAssets || got.Start != (domain.YearMonth{Year: 2026, Month: 10}) || got.Years != 20 || got.InvestmentRatePct != 18.8235 {
		t.Fatalf("defaults inesperados: %#v", got)
	}
	if len(got.SalaryRules) != 1 || got.SalaryRules[0].From != (domain.YearMonth{Year: 2027, Month: 9}) || got.SalaryRules[0].Value != 10 {
		t.Fatalf("regras por defecto inesperadas: %#v", got.SalaryRules)
	}
}

func TestProjectionDefaultsWithAssets(t *testing.T) {
	s := fixedClockServer(t, time.Date(2030, 5, 1, 0, 0, 0, 0, time.UTC))
	seedSkippedMonth(t, s)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/tools/projection/defaults", nil, authCookieWithServerClock(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("defaults con activos: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got projectionDefaultsResponse
	decodeBody(t, rr, &got)
	if !got.StartFromAssets || got.Start != (domain.YearMonth{Year: 2026, Month: 1}) || got.SalaryRules[0].From != (domain.YearMonth{Year: 2026, Month: 9}) {
		t.Fatalf("defaults con activos inesperados: %#v", got)
	}
}

func TestProjectionStartRouteRemoved(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/tools/projection/start", nil, authCookie(t, s))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("/start debe ser 404: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}

func TestProjectionValidation(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	start := domain.YearMonth{Year: 2026, Month: 1}

	rr := doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{
		Years:               0,
		InitialInvestment:   -1,
		MonthlyReturnPct:    -100,
		Mode:                domain.ProjectionModeContribution,
		MonthlyContribution: -1,
		Rules: []domain.GrowthRule{{
			Kind:        "mala",
			Value:       1,
			EveryMonths: -1,
			From:        domain.YearMonth{Year: 0, Month: 1},
		}},
	}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validación contribution: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	for _, key := range []string{"start", "years", "initialInvestment", "monthlyReturnPct", "monthlyContribution", "rules[0].kind", "rules[0].everyMonths", "rules[0].from"} {
		if apiErr.Fields[key] == "" {
			t.Fatalf("faltou erro %s en %#v", key, apiErr.Fields)
		}
	}

	rr = doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{Start: &start, Years: 1, MonthlyReturnPct: 0, Mode: domain.ProjectionModeSalary, AnnualSalary: -1, InvestmentRatePct: 0, Rules: []domain.GrowthRule{{Kind: domain.GrowthKindPercent, Value: -100, From: start}}}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validación salary: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &apiErr)
	for _, key := range []string{"annualSalary", "investmentRatePct", "rules[0].value"} {
		if apiErr.Fields[key] == "" {
			t.Fatalf("faltou erro %s en %#v", key, apiErr.Fields)
		}
	}

	tooMany := make([]domain.GrowthRule, domain.MaxGrowthRules+1)
	for i := range tooMany {
		tooMany[i] = domain.GrowthRule{Kind: domain.GrowthKindFixed, From: start}
	}
	rr = doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{Start: &start, Years: 1, Mode: domain.ProjectionModeContribution, Rules: tooMany}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validación regras: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["rules"] == "" {
		t.Fatalf("faltou erro rules: %#v", apiErr.Fields)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/tools/projection", `{"start":{"year":2026,"month":1},"years":1,"mode":"contribution","monthlyReturnPct":0,"extra":1}`, cookie)
	if rr.Code != http.StatusBadRequest || !contains(rr.Body.String(), "campo descoñecido") {
		t.Fatalf("campo descoñecido debe ser 400: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}

func TestProjectionContributionHappyPath(t *testing.T) {
	s := testServer(t)
	start := domain.YearMonth{Year: 2026, Month: 1}
	req := projectionRequest{
		Start:               &start,
		Years:               1,
		InitialInvestment:   1000,
		MonthlyReturnPct:    0,
		Mode:                domain.ProjectionModeContribution,
		MonthlyContribution: 100,
		Rules: []domain.GrowthRule{{
			Kind:        domain.GrowthKindFixed,
			Value:       10,
			EveryMonths: 6,
			From:        start.AddMonths(6),
		}},
	}
	rr := doJSON(t, s.Handler(), http.MethodPost, "/api/tools/projection", req, authCookie(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("projection contribution: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got projectionResponse
	decodeBody(t, rr, &got)
	if len(got.Months) != 12 || got.Input.Rules == nil {
		t.Fatalf("meses ou echo inesperado: %#v", got)
	}
	if got.Months[0].Contribution != 100 || got.Months[5].Contribution != 100 || got.Months[6].Contribution != 110 || got.Months[11].Contribution != 110 {
		t.Fatalf("aportes inesperados: %#v %#v %#v", got.Months[0], got.Months[6], got.Months[11])
	}
	if got.Summary.InitialInvestment != 1000 || got.Summary.FirstContribution != 100 || got.Summary.FinalContribution != 110 || got.Summary.TotalContributions != 1260 || got.Summary.TotalInvested != 2260 || got.Summary.FinalCapital != 2260 || got.Summary.Months != 12 {
		t.Fatalf("resumo inesperado: %#v", got.Summary)
	}
}

func TestProjectionSalaryHappyPath(t *testing.T) {
	s := testServer(t)
	start := domain.YearMonth{Year: 2026, Month: 1}
	req := projectionRequest{
		Start:             &start,
		Years:             1,
		MonthlyReturnPct:  0,
		Mode:              domain.ProjectionModeSalary,
		AnnualSalary:      120000,
		InvestmentRatePct: 10,
	}
	rr := doJSON(t, s.Handler(), http.MethodPost, "/api/tools/projection", req, authCookie(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("projection salary: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got projectionResponse
	decodeBody(t, rr, &got)
	if len(got.Months) != 12 || got.Months[0].AnnualSalary != 120000 || got.Months[0].MonthlySalary != 10000 || got.Months[0].Contribution != 1000 {
		t.Fatalf("primeiro mes inesperado: %#v", got.Months)
	}
	if got.Summary.FinalAnnualSalary != 120000 || got.Summary.TotalContributions != 12000 || got.Summary.TotalInvested != 12000 || got.Summary.FinalCapital != 12000 || got.Summary.FinalContribution != 1000 {
		t.Fatalf("resumo inesperado: %#v", got.Summary)
	}
}
