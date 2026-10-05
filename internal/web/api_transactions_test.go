package web

import (
	"net/http"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
)

func TestAPITransactions(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	assetID := mustAsset(t, s, domain.Asset{Type: domain.Accion, Name: "A", AmountUSD: 1000, Month: 3, Year: 2026})
	futureID := mustAsset(t, s, domain.Asset{Type: domain.Fondo, Name: "F", AmountUSD: 500, Month: 5, Year: 2026})

	if rr := doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(assetID)+"/transactions", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}

	body := map[string]any{"assetId": assetID, "kind": "venda", "amountUsd": 25.0, "month": 4, "year": 2026}
	rr := doJSON(t, h, http.MethodPost, "/api/transactions", body, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("crear venda: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var tx domain.Transaction
	decodeBody(t, rr, &tx)
	if tx.AmountUSD != -25 || tx.AssetID != assetID {
		t.Fatalf("venda inesperada: %#v", tx)
	}
	stored, err := s.store.GetTransaction(tx.ID)
	if err != nil || stored.AmountUSD != -25 {
		t.Fatalf("venda non persistida negativa: %#v err=%v", stored, err)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(assetID)+"/transactions", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("lista: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var list struct {
		Asset domain.Asset `json:"asset"`
		Rows  []struct {
			Amount    float64 `json:"amount"`
			IsVenda   bool    `json:"isVenda"`
			IsInitial bool    `json:"isInitial"`
		} `json:"rows"`
		Totals struct{ Compra, Venda, Neto float64 } `json:"totals"`
	}
	decodeBody(t, rr, &list)
	if len(list.Rows) != 2 || !list.Rows[1].IsVenda || list.Rows[1].Amount != 25 || list.Totals.Venda != 25 || list.Totals.Neto != 975 {
		t.Fatalf("lista inesperada: %#v", list)
	}

	upd := map[string]any{"kind": "compra", "amountUsd": 40.0, "month": 5, "year": 2026}
	rr = doJSON(t, h, http.MethodPut, "/api/transactions/"+itoa(tx.ID), upd, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("put: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	decodeBody(t, rr, &tx)
	if tx.AmountUSD != 40 || tx.Month != 5 {
		t.Fatalf("put inesperado: %#v", tx)
	}

	bad := map[string]any{"assetId": assetID, "kind": "troco", "amountUsd": 0, "month": 2, "year": 2026}
	rr = doJSON(t, h, http.MethodPost, "/api/transactions", bad, cookie)
	assertFieldErrors(t, rr, "kind", "amountUsd", "month")
	if !strings.Contains(rr.Body.String(), "a transacción non pode ser anterior") {
		t.Fatalf("faltou mensaxe de data anterior: %s", rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodPost, "/api/transactions", map[string]any{"assetId": int64(999), "kind": "compra", "amountUsd": 1, "month": 4, "year": 2026}, cookie)
	assertFieldErrors(t, rr, "assetId")

	batch := map[string]any{"assetId": assetID, "items": []map[string]any{{"kind": "compra", "amountUsd": 10.0, "month": 6, "year": 2026}, {"kind": "venda", "amountUsd": 0, "month": 6, "year": 2026}}}
	rr = doJSON(t, h, http.MethodPost, "/api/transactions/batch", batch, cookie)
	assertFieldErrors(t, rr, "items[1].amountUsd")
	txs, err := s.store.ListTransactionsByAsset(assetID)
	if err != nil || len(txs) != 1 {
		t.Fatalf("o lote inválido non foi atómico: len=%d err=%v", len(txs), err)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/transactions/month-assets?year=2026&month=4", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("month-assets: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var ma struct {
		Period   domain.YearMonth `json:"period"`
		Eligible []domain.Asset   `json:"eligible"`
		Omitted  []domain.Asset   `json:"omitted"`
	}
	decodeBody(t, rr, &ma)
	if len(ma.Eligible) != 1 || ma.Eligible[0].ID != assetID || len(ma.Omitted) != 1 || ma.Omitted[0].ID != futureID {
		t.Fatalf("month-assets inesperado: %#v", ma)
	}

	monthBad := map[string]any{"month": 4, "year": 2026, "items": []map[string]any{{"assetId": assetID, "amountUsd": 5.0}, {"assetId": futureID, "amountUsd": 6.0}, {"assetId": assetID, "amountUsd": 7.0}}}
	rr = doJSON(t, h, http.MethodPost, "/api/transactions/month", monthBad, cookie)
	assertFieldErrors(t, rr, "items[1].assetId", "items[2].assetId")
	txs, _ = s.store.ListTransactionsByAsset(assetID)
	if len(txs) != 1 {
		t.Fatalf("o lote mensual inválido non foi atómico: %d", len(txs))
	}
	monthOK := map[string]any{"month": 4, "year": 2026, "items": []map[string]any{{"assetId": assetID, "amountUsd": 5.0}}}
	rr = doJSON(t, h, http.MethodPost, "/api/transactions/month", monthOK, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("lote mensual válido: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var ids idsResponse
	decodeBody(t, rr, &ids)
	if len(ids.IDs) != 1 {
		t.Fatalf("ids inesperados: %#v", ids)
	}

	if rr = doJSON(t, h, http.MethodGet, "/api/assets/9999/transactions", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("lista activo inexistente: %d", rr.Code)
	}
	if rr = doJSON(t, h, http.MethodPut, "/api/transactions/9999", upd, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("put inexistente: %d", rr.Code)
	}
	if rr = doJSON(t, h, http.MethodDelete, "/api/transactions/9999", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("delete inexistente: %d", rr.Code)
	}
	rr = doJSON(t, h, http.MethodDelete, "/api/transactions/"+itoa(tx.ID), nil, cookie)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}

func mustAsset(t *testing.T, s *Server, a domain.Asset) int64 {
	t.Helper()
	id, err := s.store.InsertAsset(a)
	if err != nil {
		t.Fatalf("InsertAsset: %v", err)
	}
	return id
}
