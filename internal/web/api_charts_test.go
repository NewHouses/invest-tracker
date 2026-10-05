package web

import (
	"bytes"
	"net/http"
	"testing"
)

func TestChartsRequireSession(t *testing.T) {
	s := testServer(t)
	for _, path := range []string{"/api/charts/total", "/api/charts/assets"} {
		rr := doJSON(t, s.Handler(), http.MethodGet, path, nil, nil)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s: código=%d, esperabamos 401", path, rr.Code)
		}
	}
}

func TestChartsHappyAndNaNAsNull(t *testing.T) {
	s := testServer(t)
	a, _ := seedSkippedMonth(t, s)
	h := s.Handler()
	cookie := authCookie(t, s)

	paths := []string{
		"/api/charts/distribution",
		"/api/charts/asset/" + jsonID(a.ID),
		"/api/charts/type/accion",
		"/api/charts/type/accion/assets",
		"/api/charts/assets",
		"/api/charts/types",
		"/api/charts/total",
	}
	for _, path := range paths {
		rr := doJSON(t, h, http.MethodGet, path, nil, cookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: código=%d corpo=%s", path, rr.Code, rr.Body.String())
		}
	}

	rr := doJSON(t, h, http.MethodGet, "/api/charts/type/accion/assets", nil, cookie)
	var chart apiLineChart
	decodeBody(t, rr, &chart)
	if len(chart.Months) != 3 || len(chart.Series) != 2 {
		t.Fatalf("gráfica inesperada: %#v", chart)
	}
	if chart.Series[0].Label != "A" || chart.Series[0].Values[1] != nil {
		t.Fatalf("o mes sen resultado de A debe serializarse como null: %#v", chart.Series[0])
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"values":[1000,null,1100]`)) {
		t.Fatalf("NaN debe ser null no JSON: %s", rr.Body.String())
	}

	rr = doJSON(t, h, http.MethodGet, "/api/charts/assets", nil, cookie)
	decodeBody(t, rr, &chart)
	if chart.Title != "Todos os ativos" || len(chart.Series) != 2 || chart.Series[0].Values[1] != nil {
		t.Fatalf("todos os activos debe conservar o gap como null: %#v", chart)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"values":[1000,null,1100]`)) {
		t.Fatalf("NaN debe ser null no JSON de todos os activos: %s", rr.Body.String())
	}
}

func TestChartsValidationNotFoundAndEmptySlices(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	cookie := authCookie(t, s)

	if rr := doJSON(t, h, http.MethodGet, "/api/charts/asset/999", nil, cookie); rr.Code != http.StatusNotFound {
		t.Fatalf("activo descoñecido: código=%d", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/charts/type/invalido", nil, cookie); rr.Code != http.StatusBadRequest {
		t.Fatalf("tipo inválido: código=%d", rr.Code)
	}

	rr := doJSON(t, h, http.MethodGet, "/api/charts/distribution", nil, cookie)
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(`"items":[]`)) {
		t.Fatalf("distribution baleira debe ter items=[]: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/charts/total", nil, cookie)
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(`"months":[]`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"series":[]`)) {
		t.Fatalf("line chart baleira debe ter slices=[]: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/charts/assets", nil, cookie)
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(`"months":[]`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"series":[]`)) {
		t.Fatalf("todos os activos baleiro debe ter slices=[]: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}
