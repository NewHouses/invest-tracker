package web

import (
	"net/http"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
)

func TestAPIDividends(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	if rr := doJSON(t, h, http.MethodGet, "/api/dividends", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}

	rr := doJSON(t, h, http.MethodGet, "/api/dividends", nil, cookie)
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("lista baleira inesperada: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, h, http.MethodPost, "/api/dividends", map[string]any{"amountUsd": 12.5, "month": 7, "year": 2026}, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("crear dividendo: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var d domain.Dividend
	decodeBody(t, rr, &d)
	if d.ID == 0 || d.AmountUSD != 12.5 || d.Month != 7 || d.Year != 2026 {
		t.Fatalf("dividendo inesperado: %#v", d)
	}
	list, err := s.store.ListDividends()
	if err != nil || len(list) != 1 || list[0].ID != d.ID {
		t.Fatalf("dividendo non persistido: %#v err=%v", list, err)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/dividends", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("lista: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got []domain.Dividend
	decodeBody(t, rr, &got)
	if len(got) != 1 || got[0].ID != d.ID {
		t.Fatalf("lista inesperada: %#v", got)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/dividends", map[string]any{"amountUsd": 0, "month": 13, "year": 2200}, cookie)
	assertFieldErrors(t, rr, "amountUsd", "month", "year")

	if rr = doJSON(t, h, http.MethodDelete, "/api/dividends/9999", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("delete inexistente: %d", rr.Code)
	}
	rr = doJSON(t, h, http.MethodDelete, "/api/dividends/"+itoa(d.ID), nil, cookie)
	if rr.Code != http.StatusNoContent || rr.Body.Len() != 0 {
		t.Fatalf("delete: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if list, _ := s.store.ListDividends(); len(list) != 0 {
		t.Fatalf("dividendo non borrado: %#v", list)
	}
}
