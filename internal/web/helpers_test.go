package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// authCookie crea directamente no store unha sesión válida durante unha hora
// e devolve a cookie que a representa.
func authCookie(t *testing.T, s *Server) *http.Cookie {
	t.Helper()
	token := "token-de-proba-" + strings.ReplaceAll(t.Name(), "/", "-")
	if err := s.store.CreateSession(hashToken(token), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return &http.Cookie{Name: sessionCookieName, Value: token}
}

// doJSON executa unha petición contra o handler completo (con middlewares).
// body pode ser nil, un string (enviado tal cal) ou calquera valor (en JSON).
func doJSON(t *testing.T, h http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
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

func TestRequireSession(t *testing.T) {
	s := testServer(t)
	s.handleAPI("GET /api/proba-sesion", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	h := s.Handler()

	if rr := doJSON(t, h, http.MethodGet, "/api/proba-sesion", nil, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen cookie: código=%d, esperabamos 401", rr.Code)
	}
	bad := &http.Cookie{Name: sessionCookieName, Value: "inventado"}
	if rr := doJSON(t, h, http.MethodGet, "/api/proba-sesion", nil, bad); rr.Code != http.StatusUnauthorized {
		t.Fatalf("cookie descoñecida: código=%d, esperabamos 401", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/proba-sesion", nil, authCookie(t, s)); rr.Code != http.StatusOK {
		t.Fatalf("sesión válida: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	expiredToken := "caducado"
	if err := s.store.CreateSession(hashToken(expiredToken), time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	expired := &http.Cookie{Name: sessionCookieName, Value: expiredToken}
	if rr := doJSON(t, h, http.MethodGet, "/api/proba-sesion", nil, expired); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sesión caducada: código=%d, esperabamos 401", rr.Code)
	}
}
