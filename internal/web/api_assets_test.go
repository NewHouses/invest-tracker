package web

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
)

func TestAPIAssets(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)

	if rr := doJSON(t, h, http.MethodGet, "/api/assets", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}

	rr := doJSON(t, h, http.MethodGet, "/api/assets", nil, cookie)
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("lista baleira inesperada: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	create := map[string]any{"type": "accion", "name": "  ETF Global  ", "amountUsd": 1000.0, "month": 3, "year": 2026}
	rr = doJSON(t, h, http.MethodPost, "/api/assets", create, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("crear: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var asset domain.Asset
	decodeBody(t, rr, &asset)
	if asset.ID == 0 || asset.Type != domain.Accion || asset.Name != "ETF Global" || asset.AmountUSD != 1000 || asset.Month != 3 || asset.Year != 2026 {
		t.Fatalf("activo creado inesperado: %#v", asset)
	}
	stored, err := s.store.GetAsset(asset.ID)
	if err != nil || stored.Name != "ETF Global" {
		t.Fatalf("activo non persistido: %#v err=%v", stored, err)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(asset.ID), nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("get: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var got domain.Asset
	decodeBody(t, rr, &got)
	if got.ID != asset.ID {
		t.Fatalf("get devolveu %#v", got)
	}

	update := map[string]any{"name": "Novo nome", "amountUsd": 1200.0, "month": 4, "year": 2026}
	rr = doJSON(t, h, http.MethodPut, "/api/assets/"+itoa(asset.ID), update, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("put: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &got)
	if got.Type != domain.Accion || got.Name != "Novo nome" || got.AmountUSD != 1200 || got.Month != 4 {
		t.Fatalf("put inesperado: %#v", got)
	}

	bad := map[string]any{"type": "raro", "name": " ", "amountUsd": -1, "month": 13, "year": 2200}
	rr = doJSON(t, h, http.MethodPost, "/api/assets", bad, cookie)
	assertFieldErrors(t, rr, "type", "name", "amountUsd", "month", "year")

	rr = doJSON(t, h, http.MethodPost, "/api/assets", `{"type":"accion","name":"A","amountUsd":1,"month":1,"year":2026,"extra":true}`, cookie)
	if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "campo descoñecido") {
		t.Fatalf("campo descoñecido: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	if rr = doJSON(t, h, http.MethodGet, "/api/assets/9999", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("get inexistente: código=%d", rr.Code)
	}
	if rr = doJSON(t, h, http.MethodPut, "/api/assets/9999", update, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("put inexistente: código=%d", rr.Code)
	}

	txID, err := s.store.InsertTransaction(domain.Transaction{AssetID: asset.ID, AmountUSD: 50, Month: 5, Year: 2026})
	if err != nil {
		t.Fatalf("InsertTransaction: %v", err)
	}
	resID, err := s.store.InsertMonthlyResult(domain.MonthlyResult{AssetID: asset.ID, ResultUSD: 1300, Month: 5, Year: 2026})
	if err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}
	rr = doJSON(t, h, http.MethodDelete, "/api/assets/"+itoa(asset.ID), nil, cookie)
	if rr.Code != http.StatusNoContent || rr.Body.Len() != 0 {
		t.Fatalf("delete: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if _, err := s.store.GetTransaction(txID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("a transacción non se borrou en cascada: %v", err)
	}
	if err := s.store.DeleteMonthlyResult(resID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("o resultado non se borrou en cascada: %v", err)
	}
	if rr = doJSON(t, h, http.MethodDelete, "/api/assets/"+itoa(asset.ID), nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("delete inexistente: código=%d", rr.Code)
	}
}

func itoa(id int64) string {
	return strconvFormatInt(id)
}
