package web

import (
	"net/http"
	"strings"
	"testing"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
)

func TestAPIResults(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	assetID := mustAsset(t, s, domain.Asset{Type: domain.Indice, Name: "Índice", AmountUSD: 1000, Month: 1, Year: 2026})
	lateID := mustAsset(t, s, domain.Asset{Type: domain.Fondo, Name: "Tarde", AmountUSD: 700, Month: 3, Year: 2026})

	if rr := doJSON(t, h, http.MethodGet, "/api/results/eligible?year=2026&month=1", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}

	rr := doJSON(t, h, http.MethodGet, "/api/results/eligible?year=2025&month=12", nil, cookie)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"items":[]`) {
		t.Fatalf("eligible baleiro inesperado: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/results/eligible?year=2026&month=1", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("eligible: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var elig eligibleResultsResponse
	decodeBody(t, rr, &elig)
	if len(elig.Items) != 1 || elig.Items[0].Asset.ID != assetID || elig.Items[0].Holding != 1000 {
		t.Fatalf("eligible inesperado: %#v", elig)
	}

	rr = doJSON(t, h, http.MethodPost, "/api/results", map[string]any{"assetId": assetID, "resultUsd": 1100.0, "month": 1, "year": 2026}, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("crear resultado: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var result domain.MonthlyResult
	decodeBody(t, rr, &result)
	if result.ID == 0 || result.AssetID != assetID || result.ResultUSD != 1100 {
		t.Fatalf("resultado inesperado: %#v", result)
	}

	rr = doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(assetID)+"/results", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("lista resultados: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var results []domain.MonthlyResult
	decodeBody(t, rr, &results)
	if len(results) != 1 || results[0].ID != result.ID {
		t.Fatalf("lista resultados inesperada: %#v", results)
	}
	freshID := mustAsset(t, s, domain.Asset{Type: domain.Accion, Name: "Sen resultados", AmountUSD: 10, Month: 1, Year: 2026})
	rr = doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(freshID)+"/results", nil, cookie)
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("lista resultados baleira inesperada: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	rr = doJSON(t, h, http.MethodPost, "/api/results", map[string]any{"assetId": lateID, "resultUsd": 1.0, "month": 2, "year": 2026}, cookie)
	assertFieldErrors(t, rr, "assetId")
	if !strings.Contains(rr.Body.String(), "o ativo non ten capital investido nese mes") {
		t.Fatalf("mensaxe de elixibilidade inesperada: %s", rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodPost, "/api/results", map[string]any{"assetId": int64(999), "resultUsd": 0, "month": 13, "year": 2200}, cookie)
	assertFieldErrors(t, rr, "assetId", "resultUsd", "month", "year")

	badClose := map[string]any{"month": 2, "year": 2026, "items": []map[string]any{{"assetId": assetID, "resultUsd": 1200.0}, {"assetId": lateID, "resultUsd": 800.0}, {"assetId": assetID, "resultUsd": 1300.0}}}
	rr = doJSON(t, h, http.MethodPost, "/api/results/close-month", badClose, cookie)
	assertFieldErrors(t, rr, "items[1].assetId", "items[2].assetId")
	items, err := closemonth.Eligible(s.store, 2026, 2)
	if err != nil {
		t.Fatalf("Eligible: %v", err)
	}
	for _, item := range items {
		if item.Asset.ID == assetID && item.HasResult {
			t.Fatal("o peche inválido gardou resultados")
		}
	}
	goodClose := map[string]any{"month": 2, "year": 2026, "items": []map[string]any{{"assetId": assetID, "resultUsd": 1200.0}}}
	rr = doJSON(t, h, http.MethodPost, "/api/results/close-month", goodClose, cookie)
	if rr.Code != http.StatusCreated {
		t.Fatalf("peche válido: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var ids idsResponse
	decodeBody(t, rr, &ids)
	if len(ids.IDs) != 1 {
		t.Fatalf("ids inesperados: %#v", ids)
	}

	rr = doJSON(t, h, http.MethodDelete, "/api/results?year=2026&month=2", nil, cookie)
	if rr.Code != http.StatusOK {
		t.Fatalf("limpar mes: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	var deleted deletedResponse
	decodeBody(t, rr, &deleted)
	if deleted.Deleted != 1 {
		t.Fatalf("deleted inesperado: %#v", deleted)
	}

	if rr = doJSON(t, h, http.MethodGet, "/api/assets/9999/results", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("lista activo inexistente: %d", rr.Code)
	}
	if rr = doJSON(t, h, http.MethodDelete, "/api/results/9999", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("delete inexistente: %d", rr.Code)
	}
	rr = doJSON(t, h, http.MethodDelete, "/api/results/"+itoa(result.ID), nil, cookie)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}

// Corrixir un resultado (desde "Engadir resultado" ou "Pechar mes") debe
// substituílo: antes quedaban as dúas filas na lista e, ao borrar a visible,
// o mes volvía mostrar o valor vello.
func TestAPIResultsCorrectionReplacesPreviousValue(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)
	assetID := mustAsset(t, s, domain.Asset{Type: domain.Indice, Name: "Índice", AmountUSD: 1000, Month: 1, Year: 2026})

	for _, value := range []float64{1100, 1150} {
		if rr := doJSON(t, h, http.MethodPost, "/api/results", map[string]any{"assetId": assetID, "resultUsd": value, "month": 1, "year": 2026}, cookie); rr.Code != http.StatusCreated {
			t.Fatalf("gardar %.0f: código=%d corpo=%s", value, rr.Code, rr.Body.String())
		}
	}
	close := closeMonthRequest{Month: 1, Year: 2026, Items: []closeMonthRequestItem{{AssetID: assetID, ResultUSD: 1175}}}
	if rr := doJSON(t, h, http.MethodPost, "/api/results/close-month", close, cookie); rr.Code != http.StatusCreated {
		t.Fatalf("pechar mes: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	rr := doJSON(t, h, http.MethodGet, "/api/assets/"+itoa(assetID)+"/results", nil, cookie)
	var results []domain.MonthlyResult
	decodeBody(t, rr, &results)
	if len(results) != 1 || results[0].ResultUSD != 1175 {
		t.Fatalf("lista tras corrixir = %#v, esperabamos un só resultado con 1175", results)
	}

	if rr := doJSON(t, h, http.MethodDelete, "/api/results/"+itoa(results[0].ID), nil, cookie); rr.Code != http.StatusNoContent {
		t.Fatalf("borrar: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/reports/asset/"+itoa(assetID)+"/month?year=2026&month=1", nil, cookie)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"hasResult":false`) {
		t.Fatalf("tras borrar o resultado o mes non debería ter ningún: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}
