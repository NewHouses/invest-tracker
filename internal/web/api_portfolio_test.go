package web

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	portfoliodata "invest-tracker/internal/portfolio"
)

func TestPortfolioRequireSession(t *testing.T) {
	s := testServer(t)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/portfolio", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("código=%d, esperabamos 401", rr.Code)
	}
}

func TestPortfolioEmptyDB(t *testing.T) {
	s := fixedClockServer(t, time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC))
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/portfolio", nil, authCookie(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("código=%d corpo=%s, esperabamos 200", rr.Code, rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"rows":[]`)) {
		t.Fatalf("rows baleiro debe serializarse como []: %s", rr.Body.String())
	}
	var got portfoliodata.Portfolio
	decodeBody(t, rr, &got)
	if got.Period.Year != 2026 || got.Period.Month != 4 || len(got.Rows) != 0 || got.PendingCount != 0 {
		t.Fatalf("carteira baleira inesperada: %#v", got)
	}
}

func TestPortfolioHappyPendingFlags(t *testing.T) {
	s := fixedClockServer(t, time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC))
	seedSkippedMonth(t, s)
	rr := doJSON(t, s.Handler(), http.MethodGet, "/api/portfolio", nil, authCookie(t, s))
	if rr.Code != http.StatusOK {
		t.Fatalf("código=%d corpo=%s, esperabamos 200", rr.Code, rr.Body.String())
	}
	var got portfoliodata.Portfolio
	decodeBody(t, rr, &got)
	if got.Period.Year != 2026 || got.Period.Month != 4 {
		t.Fatalf("período=%#v, esperabamos 04/2026", got.Period)
	}
	if got.PendingCount != 2 || len(got.Rows) != 2 {
		t.Fatalf("pendentes inesperados: %#v", got)
	}
	for _, row := range got.Rows {
		if !row.Pending {
			t.Fatalf("a fila debía estar pendente: %#v", row)
		}
	}
}
