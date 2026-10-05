package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"invest-tracker/internal/domain"
)

func fixedClockServer(t *testing.T, now time.Time) *Server {
	t.Helper()
	return New(testStore(t), testStaticFS(), Options{Logger: log.New(io.Discard, "", 0), Now: func() time.Time { return now }})
}

func TestDashboardRequireSession(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/dashboard", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("código=%d, esperabamos 401", rr.Code)
	}
}

func TestDashboardHappyPendingAndEmptySlices(t *testing.T) {
	s := fixedClockServer(t, time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC))
	a, _ := seedSkippedMonth(t, s)
	h := s.Handler()
	cookie := authCookie(t, s)

	rr := doJSON(t, h, http.MethodGet, "/api/dashboard", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("dashboard: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got dashboardResponse
	decodeBody(t, rr, &got)
	if got.AssetCount != 2 || got.KPIs.TotalInvested != 1500 || !got.KPIs.HasCurrentValue || got.KPIs.TotalGain != 110 || got.KPIs.AvgIndexPct < 2.44 || got.KPIs.AvgIndexPct > 2.45 {
		t.Fatalf("KPIs inesperados: %#v", got.KPIs)
	}
	if got.Portfolio.Period.Year != 2026 || got.Portfolio.Period.Month != 4 || got.Portfolio.PendingCount != 2 || len(got.Portfolio.Rows) != 2 {
		t.Fatalf("carteira inesperada: %#v", got.Portfolio)
	}
	if !got.Portfolio.Rows[0].Pending || !got.Portfolio.Rows[1].Pending {
		t.Fatalf("as dúas filas debían estar pendentes: %#v", got.Portfolio.Rows)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if _, ok := raw["pending"]; ok {
		t.Fatalf("a resposta xa non debe incluír a chave pending: %s", rr.Body.String())
	}
	if len(got.Evolution.Months) != 3 || len(got.Distribution.Items) != 2 {
		t.Fatalf("gráficas inesperadas: %#v", got)
	}

	if _, err := s.store.InsertMonthlyResult(domain.MonthlyResult{AssetID: a.ID, ResultUSD: 1110, Month: 4, Year: 2026}); err != nil {
		t.Fatalf("InsertMonthlyResult abril: %v", err)
	}
	rr = doJSON(t, h, http.MethodGet, "/api/dashboard", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("dashboard 2: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &got)
	if got.Portfolio.PendingCount != 1 || len(got.Portfolio.Rows) != 2 || got.Portfolio.Rows[0].Pending || !got.Portfolio.Rows[1].Pending {
		t.Fatalf("debe quedar só B pendente: %#v", got.Portfolio)
	}

	empty := fixedClockServer(t, time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC))
	rr = doJSON(t, empty.Handler(), http.MethodGet, "/api/dashboard", nil, authCookie(t, empty))
	if rr.Code != http.StatusOK {
		t.Fatalf("dashboard baleiro: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if body := rr.Body.String(); !containsAll(body, []string{`"months":[]`, `"series":[]`, `"items":[]`, `"rows":[]`}) {
		t.Fatalf("slices baleiros deben ser []: %s", body)
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !contains(s, part) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
