package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtotalreport"
)

func seedSkippedMonth(t *testing.T, s *Server) (domain.Asset, domain.Asset) {
	t.Helper()
	a := domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Month: 1, Year: 2026}
	b := domain.Asset{Type: domain.Accion, Name: "B", AmountUSD: 500, Month: 1, Year: 2026}
	id, err := s.store.InsertAsset(a)
	if err != nil {
		t.Fatalf("InsertAsset A: %v", err)
	}
	a.ID = id
	id, err = s.store.InsertAsset(b)
	if err != nil {
		t.Fatalf("InsertAsset B: %v", err)
	}
	b.ID = id
	for _, r := range []domain.MonthlyResult{
		{AssetID: a.ID, ResultUSD: 1000, Month: 1, Year: 2026},
		{AssetID: b.ID, ResultUSD: 500, Month: 1, Year: 2026},
		{AssetID: b.ID, ResultUSD: 500, Month: 2, Year: 2026},
		{AssetID: a.ID, ResultUSD: 1100, Month: 3, Year: 2026},
		{AssetID: b.ID, ResultUSD: 510, Month: 3, Year: 2026},
	} {
		if _, err := s.store.InsertMonthlyResult(r); err != nil {
			t.Fatalf("InsertMonthlyResult: %v", err)
		}
	}
	return a, b
}

func compactJSON(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, b); err != nil {
		t.Fatalf("compactando JSON: %v\n%s", err, string(b))
	}
	return out.Bytes()
}

func TestReportsRequireSession(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/reports/total/history", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("código=%d, esperabamos 401", rr.Code)
	}
}

func TestReportsHappyAndParity(t *testing.T) {
	s := testServer(t)
	a, _ := seedSkippedMonth(t, s)
	h := s.Handler()
	cookie := authCookie(t, s)

	cases := []string{
		"/api/reports/asset/" + jsonID(a.ID) + "/month?year=2026&month=3",
		"/api/reports/type/accion/month?year=2026&month=3",
		"/api/reports/asset/" + jsonID(a.ID) + "/history",
		"/api/reports/type/accion/history",
	}
	for _, path := range cases {
		rr := doJSON(t, h, http.MethodGet, path, nil, cookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: código=%d corpo=%s", path, rr.Code, rr.Body.String())
		}
	}

	rr := doJSON(t, h, http.MethodGet, "/api/reports/total/month?year=2026&month=3", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("total/month: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var month viewtotalreport.Report
	decodeBody(t, rr, &month)
	if !month.HasMetrics || month.GainNoDiv != 110 || month.PctNoDiv < 7.33 || month.PctNoDiv > 7.34 {
		t.Fatalf("métricas total/month inesperadas: %#v", month)
	}
	expectedMonth, err := viewtotalreport.Build(s.store, 2026, 3)
	if err != nil {
		t.Fatalf("Build total/month: %v", err)
	}
	expectedMonthJSON, _ := json.Marshal(expectedMonth)
	if !bytes.Equal(compactJSON(t, rr.Body.Bytes()), expectedMonthJSON) {
		t.Fatalf("JSON total/month non coincide\napi=%s\nexp=%s", compactJSON(t, rr.Body.Bytes()), expectedMonthJSON)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/reports/total/history", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("total/history: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var history viewtotalhistory.History
	decodeBody(t, rr, &history)
	if !history.HasAverages || history.AvgIndexPct < 2.44 || history.AvgIndexPct > 2.45 {
		t.Fatalf("media histórica inesperada: %#v", history)
	}
	expectedHistory, err := viewtotalhistory.Build(s.store)
	if err != nil {
		t.Fatalf("Build total/history: %v", err)
	}
	expectedHistoryJSON, _ := json.Marshal(totalHistoryDTO(expectedHistory))
	if !bytes.Equal(compactJSON(t, rr.Body.Bytes()), expectedHistoryJSON) {
		t.Fatalf("JSON total/history non coincide\napi=%s\nexp=%s", compactJSON(t, rr.Body.Bytes()), expectedHistoryJSON)
	}
}

func TestReportsValidationNotFoundAndEmptySlices(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)

	if rr := doJSON(t, h, http.MethodGet, "/api/reports/asset/999/month?year=2026&month=3", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("activo descoñecido: código=%d", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/reports/type/mal/month?year=2026&month=3", nil, cookie); rr.Code != http.StatusBadRequest {
		t.Fatalf("tipo inválido: código=%d", rr.Code)
	}
	rr := doJSON(t, h, http.MethodGet, "/api/reports/total/month?year=2026", nil, cookie)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("query inválida: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var apiErr apiError
	decodeBody(t, rr, &apiErr)
	if apiErr.Fields["month"] == "" {
		t.Fatalf("faltou fields.month: %#v", apiErr)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/reports/total/history", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("empty total/history: código=%d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"rows":[]`)) {
		t.Fatalf("rows debe ser []: %s", rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/reports/type/accion/month?year=2026&month=3", nil, cookie)
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"assets":[]`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"active":[]`)) {
		t.Fatalf("assets/active deben ser []: %s", rr.Body.String())
	}
}

func jsonID(id int64) string {
	return strconv.FormatInt(id, 10)
}
