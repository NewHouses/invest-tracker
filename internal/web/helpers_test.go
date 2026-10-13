package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// doJSON executa unha petición contra o handler completo (con middlewares),
// dirixida a localhost como o faría o navegador do propio equipo.
// body pode ser nil, un string (enviado tal cal) ou calquera valor (en JSON).
func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return doJSONLang(t, h, "", method, path, body)
}

// doJSONLang é coma doJSON pero enviando a cabeceira Accept-Language, como fai
// a web co idioma que escolleu o usuario.
func doJSONLang(t *testing.T, h http.Handler, acceptLanguage, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	switch b := body.(type) {
	case nil:
	case string:
		buf.WriteString(b)
	default:
		if err := json.NewEncoder(&buf).Encode(b); err != nil {
			t.Fatalf("codificando o corpo: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Host = "localhost:8080"
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// decodeBody decodifica a resposta JSON en dst e falla o test se non se pode.
func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), dst); err != nil {
		t.Fatalf("resposta non é JSON válido (%v): %s", err, rr.Body.String())
	}
}

// A web xa non ten contrasinal: a API responde sen ningunha sesión.
func TestAPIWithoutLogin(t *testing.T) {
	s := testServer(t)
	s.handleAPI("GET /api/proba", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	h := s.Handler()

	if rr := doJSON(t, h, http.MethodGet, "/api/proba", nil); rr.Code != http.StatusOK {
		t.Fatalf("sen sesión: código=%d corpo=%s, esperabamos 200", rr.Code, rr.Body.String())
	}
	for _, path := range []string{"/api/auth/status", "/api/auth/login"} {
		if rr := doJSON(t, h, http.MethodGet, path, nil); rr.Code != http.StatusNotFound {
			t.Fatalf("%s: código=%d, esperabamos 404 (xa non hai autenticación)", path, rr.Code)
		}
	}
}
