package web

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"invest-tracker/internal/store"
)

func testAuthServer(t *testing.T, now *time.Time) *Server {
	t.Helper()
	s := New(testStore(t), testStaticFS(), Options{Logger: log.New(io.Discard, "", 0), Now: func() time.Time { return *now }})
	s.pbkdf2Iter = 1000
	return s
}

func localJSON(t *testing.T, h http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			buf.WriteString(s)
		} else if err := jsonEncode(&buf, body); err != nil {
			t.Fatalf("codificando JSON: %v", err)
		}
	}
	req := httptest.NewRequest(method, "http://localhost:8080"+path, &buf)
	req.RemoteAddr = "127.0.0.1:5555"
	req.Host = "localhost:8080"
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

func jsonEncode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func TestAuthStatusLocalAndRemoteFreshInstall(t *testing.T) {
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	s := testAuthServer(t, &now)
	h := s.Handler()

	rr := localJSON(t, h, http.MethodGet, "/api/auth/status", nil, nil)
	var got struct{ SetupRequired, Authenticated, SetupAllowed bool }
	decodeBody(t, rr, &got)
	if rr.Code != http.StatusOK || !got.SetupRequired || got.Authenticated || !got.SetupAllowed {
		t.Fatalf("status local inesperado: código=%d %#v", rr.Code, got)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	req.RemoteAddr = "192.168.1.20:5555"
	req.Host = "192.168.1.20:8080"
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	got = struct{ SetupRequired, Authenticated, SetupAllowed bool }{}
	decodeBody(t, rr, &got)
	if rr.Code != http.StatusOK || !got.SetupRequired || got.Authenticated || got.SetupAllowed {
		t.Fatalf("status remoto inesperado: código=%d %#v", rr.Code, got)
	}
}

func TestAuthSetupLocalityAndCookie(t *testing.T) {
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	s := testAuthServer(t, &now)
	h := s.Handler()

	rr := doJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "contrasinal"}, nil)
	if rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "só se pode crear") {
		t.Fatalf("setup remoto: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "http://evil.example:8080/api/auth/setup", strings.NewReader(`{"password":"contrasinal"}`))
	req.RemoteAddr = "127.0.0.1:5555"
	req.Host = "evil.example:8080"
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("setup DNS rebinding: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	rr = localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": ""}, nil)
	if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), `"password"`) {
		t.Fatalf("setup baleiro: código=%d corpo=%s", rr.Code, rr.Body.String())
	}

	// Sen requisitos de lonxitude: un só carácter é válido.
	rr = localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "x"}, nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("setup local: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Secure || cookies[0].MaxAge != int((30*24*time.Hour)/time.Second) {
		t.Fatalf("cookie inesperada: %#v", cookies)
	}

	rr = localJSON(t, h, http.MethodGet, "/api/auth/status", nil, cookies[0])
	var st struct{ SetupRequired, Authenticated, SetupAllowed bool }
	decodeBody(t, rr, &st)
	if !st.Authenticated || st.SetupRequired {
		t.Fatalf("status autenticado inesperado: %#v", st)
	}

	rr = localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "outracontrasinal"}, nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("setup dúas veces: código=%d", rr.Code)
	}
}

func TestAuthLoginRateLimitAndRemember(t *testing.T) {
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	s := testAuthServer(t, &now)
	h := s.Handler()

	rr := localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "contrasinal", "remember": true}, nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("login sen setup: código=%d", rr.Code)
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "contrasinal"}, nil); rr.Code != http.StatusCreated {
		t.Fatalf("setup: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	for i := 0; i < 5; i++ {
		rr = localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "errado", "remember": false}, nil)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("intento %d: código=%d corpo=%s", i+1, rr.Code, rr.Body.String())
		}
	}
	rr = localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "contrasinal", "remember": false}, nil)
	if rr.Code != http.StatusTooManyRequests || rr.Header().Get("Retry-After") == "" {
		t.Fatalf("rate limit: código=%d retry=%q corpo=%s", rr.Code, rr.Header().Get("Retry-After"), rr.Body.String())
	}
	now = now.Add(time.Minute + time.Second)
	rr = localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "contrasinal", "remember": false}, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("login tras bloqueo: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	sessionCookie := rr.Result().Cookies()[0]
	if sessionCookie.MaxAge != 0 {
		t.Fatalf("cookie de sesión debe ter MaxAge 0: %#v", sessionCookie)
	}
	if ok, err := s.store.SessionValid(hashToken(sessionCookie.Value), now.Add(13*time.Hour)); err != nil || ok {
		t.Fatalf("sesión de 12h válida despois de 13h: ok=%v err=%v", ok, err)
	}

	rr = localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "contrasinal", "remember": true}, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("login remember: código=%d", rr.Code)
	}
	rememberCookie := rr.Result().Cookies()[0]
	if rememberCookie.MaxAge != int((30*24*time.Hour)/time.Second) {
		t.Fatalf("MaxAge remember inesperado: %d", rememberCookie.MaxAge)
	}
	if ok, err := s.store.SessionValid(hashToken(rememberCookie.Value), now.Add(29*24*time.Hour)); err != nil || !ok {
		t.Fatalf("sesión remember non válida: ok=%v err=%v", ok, err)
	}
}

func TestAuthLogoutInvalidatesSession(t *testing.T) {
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	s := testAuthServer(t, &now)
	h := s.Handler()
	rr := localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "contrasinal"}, nil)
	cookie := rr.Result().Cookies()[0]
	rr = localJSON(t, h, http.MethodPost, "/api/auth/logout", nil, cookie)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("logout: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if c := rr.Result().Cookies()[0]; c.MaxAge != -1 {
		t.Fatalf("cookie non expirada: %#v", c)
	}
	if ok, err := s.store.SessionValid(hashToken(cookie.Value), now); err != nil || ok {
		t.Fatalf("sesión tras logout: ok=%v err=%v", ok, err)
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/logout", nil, nil); rr.Code != http.StatusNoContent {
		t.Fatalf("logout sen sesión: código=%d", rr.Code)
	}
}

func TestAuthChangePassword(t *testing.T) {
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	s := testAuthServer(t, &now)
	h := s.Handler()
	rr := localJSON(t, h, http.MethodPost, "/api/auth/setup", map[string]string{"password": "contrasinal"}, nil)
	current := rr.Result().Cookies()[0]
	rr = localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "contrasinal", "remember": true}, nil)
	other := rr.Result().Cookies()[0]

	if rr := localJSON(t, h, http.MethodPost, "/api/auth/password", map[string]string{"currentPassword": "mal", "newPassword": "novo-contrasinal"}, current); rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "actual") {
		t.Fatalf("current errado: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/password", map[string]string{"currentPassword": "contrasinal", "newPassword": ""}, current); rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "newPassword") {
		t.Fatalf("novo baleiro: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/password", map[string]string{"currentPassword": "contrasinal", "newPassword": "novo-contrasinal"}, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("sen sesión: código=%d", rr.Code)
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/password", map[string]string{"currentPassword": "contrasinal", "newPassword": "novo-contrasinal"}, current); rr.Code != http.StatusNoContent {
		t.Fatalf("cambio contrasinal: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
	if rr := localJSON(t, h, http.MethodGet, "/api/auth/status", nil, current); rr.Code != http.StatusOK {
		t.Fatalf("sesión actual inválida: código=%d", rr.Code)
	}
	var st struct{ Authenticated bool }
	decodeBody(t, localJSON(t, h, http.MethodGet, "/api/auth/status", nil, other), &st)
	if st.Authenticated {
		t.Fatal("a outra sesión debe quedar invalidada")
	}
	if rr := localJSON(t, h, http.MethodPost, "/api/auth/login", map[string]any{"password": "novo-contrasinal", "remember": false}, nil); rr.Code != http.StatusOK {
		t.Fatalf("login co novo contrasinal: código=%d corpo=%s", rr.Code, rr.Body.String())
	}
}

func TestMainResetPassword(t *testing.T) {
	db := filepath.Join(t.TempDir(), "test.db")
	st, err := store.Open(db)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	encoded, err := hashPassword("contrasinal", 1000)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if err := st.SetSetting(passwordSettingKey, encoded); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := st.CreateSession(hashToken("token"), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_ = st.Close()

	if code := Main([]string{"-reset-password", "-db", db}); code != 0 {
		t.Fatalf("Main reset: código=%d", code)
	}
	st, err = store.Open(db)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer st.Close()
	if _, ok, err := st.GetSetting(passwordSettingKey); err != nil || ok {
		t.Fatalf("password_hash segue presente: ok=%v err=%v", ok, err)
	}
	if ok, err := st.SessionValid(hashToken("token"), time.Now()); err != nil || ok {
		t.Fatalf("sesión segue presente: ok=%v err=%v", ok, err)
	}
}
