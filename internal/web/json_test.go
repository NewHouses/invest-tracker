package web

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON_KeepsStatusAndEncoding(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]string{"nome": "<AAPL & co>"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("código=%d, esperabamos 201", rr.Code)
	}
	if got, want := rr.Body.String(), "{\"nome\":\"\\u003cAAPL \\u0026 co\\u003e\"}\n"; got != want {
		t.Fatalf("corpo=%q, esperabamos %q", got, want)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type=%q", ct)
	}
}

// Un valor que non se pode serializar (p.e. un importe infinito gardado na
// base de datos) non pode acabar nun 200 co corpo baleiro.
func TestWriteJSON_UnencodableValueIsInternalError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusOK, map[string]float64{"valor": math.Inf(1)})
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("código=%d, esperabamos 500", rr.Code)
	}
	var got apiError
	decodeBody(t, rr, &got)
	if got.Error != "erro interno" {
		t.Fatalf("erro=%q, esperabamos \"erro interno\"", got.Error)
	}
}
