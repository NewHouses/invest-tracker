package web

import (
	"net/http"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
)

func TestToolsRequireSession(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/tools/projection/start", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("código=%d, esperabamos 401", rr.Code)
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

func TestProjectionStartAndProjection(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)

	rr := doJSON(t, h, http.MethodGet, "/api/tools/projection/start", nil, cookie)
	if rr.Code != http.StatusOK || rr.Body.String() != "{\"start\":null}\n" {
		t.Fatalf("start baleiro inesperado: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{AnnualSalary: 120000, MonthlyReturnPct: 1}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("sen start debe fallar: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["start"] == "" {
		t.Fatalf("faltou fields.start: %#v", apiErr)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{AnnualSalary: 120000, MonthlyReturnPct: 1, Start: &domain.YearMonth{Year: 2026, Month: 5}}, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("projection con start: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var projected projectionResponse
	decodeBody(t, rr, &projected)
	if projected.StartFromAssets || projected.Start.Year != 2026 || projected.Start.Month != 5 || len(projected.Months) != 240 || projected.Summary.FinalCapital <= 0 {
		t.Fatalf("proxección inesperada: %#v", projected)
	}
	if projected.Rules.InvestmentRate != domain.InvestmentRate || projected.Rules.SalaryRaiseRate != domain.SalaryRaiseRate || projected.Rules.SalaryRaiseMonth != domain.SalaryRaiseMonth || projected.Rules.Years != domain.ProjectionYears {
		t.Fatalf("regras inesperadas: %#v", projected.Rules)
	}

	s2 := testServer(t)
	seedSkippedMonth(t, s2)
	h2 := s2.Handler()
	cookie2 := authCookie(t, s2)
	rr = doJSON(t, h2, http.MethodGet, "/api/tools/projection/start", nil, cookie2)
	if rr.Code != http.StatusOK || rr.Body.String() != "{\"start\":{\"year\":2026,\"month\":1}}\n" {
		t.Fatalf("start con activos inesperado: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h2, http.MethodPost, "/api/tools/projection", projectionRequest{AnnualSalary: 120000, MonthlyReturnPct: 1, Start: &domain.YearMonth{Year: 2030, Month: 1}}, cookie2)
	if rr.Code != http.StatusOK {
		t.Fatalf("projection con activos: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &projected)
	if !projected.StartFromAssets || projected.Start.Year != 2026 || projected.Start.Month != 1 || projected.Input.MonthlyReturnPct != 1 || len(projected.Months) != 240 {
		t.Fatalf("start de activos non se aplicou: %#v", projected)
	}
}

func TestProjectionValidation(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	start := domain.YearMonth{Year: 2026, Month: 1}
	rr := doJSON(t, h, http.MethodPost, "/api/tools/projection", projectionRequest{AnnualSalary: -1, MonthlyReturnPct: -100, Start: &start}, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validación: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["annualSalary"] == "" || apiErr.Fields["monthlyReturnPct"] == "" {
		t.Fatalf("faltan erros: %#v", apiErr.Fields)
	}
}
