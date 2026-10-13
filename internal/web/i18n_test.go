package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A API responde no idioma que pide a web en Accept-Language: inglés se o pide
// e galego en calquera outro caso.
func TestAPILanguage(t *testing.T) {
	s := testServer(t)
	seedSkippedMonth(t, s)
	h := s.Handler()

	badAsset := map[string]any{"type": "accion", "name": " ", "amountUsd": 0, "month": 13, "year": 2026}
	cases := []struct {
		name, accept, method, path string
		body                       any
		want                       []string
	}{
		{"validación en inglés", "en-GB,en;q=0.9,gl;q=0.8", http.MethodPost, "/api/assets", badAsset,
			[]string{`"error":"invalid data"`, "the name cannot be empty", "the amount must be a number greater than 0", "the month must be between 1 and 12"}},
		{"validación en galego", "gl-ES,gl;q=0.9,en;q=0.8", http.MethodPost, "/api/assets", badAsset,
			[]string{`"error":"datos non válidos"`, "o nome non pode estar baleiro", "o mes debe estar entre 1 e 12"}},
		{"sen cabeceira", "", http.MethodPost, "/api/assets", badAsset, []string{`"error":"datos non válidos"`}},
		{"idioma non admitido", "es-ES,es;q=0.9", http.MethodPost, "/api/assets", badAsset, []string{`"error":"datos non válidos"`}},
		{"non atopado", "en", http.MethodGet, "/api/assets/999", nil, []string{`"error":"asset not found"`}},
		{"identificador", "en", http.MethodGet, "/api/assets/abc", nil, []string{`"error":"invalid identifier"`}},
		{"JSON roto", "en", http.MethodPost, "/api/assets", "{", []string{"invalid JSON"}},
		{"campo descoñecido", "en", http.MethodPost, "/api/assets", `{"nome":"x"}`, []string{"unknown field"}},
		{"tipos en inglés", "en", http.MethodGet, "/api/meta", nil,
			[]string{`"label":"Stock"`, `"label":"Index"`, `"label":"Copy trading"`, `"label":"Fund"`}},
		{"tipos en galego", "gl", http.MethodGet, "/api/meta", nil, []string{`"label":"Acción"`, `"label":"Índice"`, `"label":"Fondo"`}},
		{"gráfica en inglés", "en", http.MethodGet, "/api/charts/assets", nil, []string{`"title":"All assets"`}},
		{"gráfica en galego", "gl", http.MethodGet, "/api/charts/assets", nil, []string{`"title":"Todos os activos"`}},
		{"gráfica dun tipo", "en", http.MethodGet, "/api/charts/type/accion/assets", nil, []string{`"title":"Stock assets"`}},
		{"reparto", "en", http.MethodPost, "/api/tools/allocation", map[string]any{"total": 0, "selection": []any{}},
			[]string{"the total must be a positive amount", "at least one type must be selected"}},
		{"proxección", "en", http.MethodPost, "/api/tools/projection", map[string]any{"years": 0, "mode": "x"},
			[]string{"the start date is required", "the years must be between 1 and", "the mode must be contribution or salary"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := doJSONLang(t, h, tc.accept, tc.method, tc.path, tc.body)
			for _, want := range tc.want {
				if !strings.Contains(rr.Body.String(), want) {
					t.Fatalf("falta %q na resposta (código %d): %s", want, rr.Code, rr.Body.String())
				}
			}
		})
	}
}

func TestHostCheckLanguage(t *testing.T) {
	h := testServer(t).Handler()
	for accept, want := range map[string]string{
		"en": "address not allowed: use the computer's IP or localhost",
		"gl": "enderezo non permitido: usa a IP do equipo ou localhost",
	} {
		for _, path := range []string{"/api/meta", "/"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Host = "malicioso.example:8080"
			req.Header.Set("Accept-Language", accept)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), want) {
				t.Fatalf("%s en %s: código=%d corpo=%q", path, accept, rr.Code, rr.Body.String())
			}
		}
	}
}
