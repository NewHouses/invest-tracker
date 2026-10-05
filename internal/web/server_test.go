package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"invest-tracker/internal/store"
)

func testStaticFS() fs.FS {
	return fstest.MapFS{
		"index.html":          {Data: []byte("<html><body>inicio web</body></html>")},
		"assets":              {Mode: fs.ModeDir},
		"assets/app-123.js":   {Data: []byte("console.log('ola')")},
		"assets/style.css":    {Data: []byte("body{}")},
		"assets/nested":       {Mode: fs.ModeDir},
		"assets/nested/a.txt": {Data: []byte("a")},
	}
}

func testStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("non se puido abrir a base de datos de proba: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func testServer(t *testing.T) *Server {
	t.Helper()
	return New(testStore(t), testStaticFS(), Options{Logger: log.New(io.Discard, "", 0)})
}

func TestStaticHandler(t *testing.T) {
	s := testServer(t)
	h := s.Handler()

	t.Run("raíz", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "inicio web") {
			t.Fatalf("agardábase index.html, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("cache inesperada para index: %q", got)
		}
	})

	t.Run("activo inmutable", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/app-123.js", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "ola") {
			t.Fatalf("agardábase o activo, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("cache inesperada para activo: %q", got)
		}
	})

	t.Run("ruta de cliente", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/activos/5", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "inicio web") {
			t.Fatalf("agardábase index para a ruta SPA, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
	})

	t.Run("sen listado de directorio", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/assets/", nil))
		if rr.Code == http.StatusOK && strings.Contains(rr.Body.String(), "app-123.js") {
			t.Fatalf("non debe amosar listado de directorio: %q", rr.Body.String())
		}
	})

	// Un ficheiro inexistente non debe caer en index.html: o navegador
	// intentaría executar HTML coma JS.
	for _, p := range []string{"/assets/vello-999.js", "/favicon-inexistente.svg"} {
		t.Run("ficheiro inexistente "+p, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, p, nil))
			if rr.Code != http.StatusNotFound {
				t.Fatalf("código=%d, esperabamos 404", rr.Code)
			}
		})
	}
}

func TestNetworkURLs_LoopbackListenerHasNoLANURLs(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()
	if got := networkURLs(ln, listenerPort(ln)); len(got) != 0 {
		t.Fatalf("networkURLs en loopback = %v, esperabamos ningunha", got)
	}
}

func TestAPIRoutes(t *testing.T) {
	s := testServer(t)
	s.handleAPI("GET /api/privada", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"ok": "si"})
	})
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil))
	if rr.Code != http.StatusNotFound || !strings.Contains(rr.Body.String(), "recurso non atopado") {
		t.Fatalf("404 API inesperado: código=%d corpo=%q", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/privada", nil))
	if rr.Code != http.StatusUnauthorized || !strings.Contains(rr.Body.String(), "autenticación necesaria") {
		t.Fatalf("401 API inesperado: código=%d corpo=%q", rr.Code, rr.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	s := testServer(t)
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, name := range []string{"X-Content-Type-Options", "Referrer-Policy", "X-Frame-Options"} {
		if rr.Header().Get(name) == "" {
			t.Fatalf("falta a cabeceira %s", name)
		}
	}
	if rr.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("falta a política CSP nas respostas non API")
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil))
	if rr.Header().Get("X-Content-Type-Options") == "" {
		t.Fatal("faltan cabeceiras de seguridade na API")
	}
	if rr.Header().Get("Content-Security-Policy") != "" {
		t.Fatal("a API non debe levar CSP")
	}
}

func TestCrossOriginProtection(t *testing.T) {
	s := testServer(t)
	s.handlePublic("POST /api/publica", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNoContent, nil)
	})
	h := s.Handler()

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/publica", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("agardábase 403 para petición cross-site, obtido %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "http://example.com/api/publica", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code == http.StatusForbidden {
		t.Fatal("a petición same-origin non debe ser rexeitada")
	}
}

func TestRecoverAPI(t *testing.T) {
	s := testServer(t)
	s.handlePublic("GET /api/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("proba")
	})
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/panic", nil))
	if rr.Code != http.StatusInternalServerError || !strings.Contains(rr.Body.String(), "erro interno") {
		t.Fatalf("recuperación inesperada: código=%d corpo=%q", rr.Code, rr.Body.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	t.Run("campo descoñecido", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"outro":1}`)), &dst)
		if err == nil || !strings.Contains(err.Error(), "campo descoñecido") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})

	t.Run("datos sobrantes", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"nome":"a"} {}`)), &dst)
		if err == nil || !strings.Contains(err.Error(), "único valor") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})

	t.Run("demasiado grande", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		body := `{"nome":"` + strings.Repeat("a", 1<<20) + `"}`
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), &dst)
		if err == nil || !strings.Contains(err.Error(), "1 MiB") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})
}

func TestPathIDAndQueryYearMonth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/items/abc", nil)
	req.SetPathValue("id", "abc")
	if _, err := pathID(req, "id"); err == nil {
		t.Fatal("agardábase erro para id non numérico")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/items/42", nil)
	req.SetPathValue("id", "42")
	if id, err := pathID(req, "id"); err != nil || id != 42 {
		t.Fatalf("id inesperado: %d, %v", id, err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api?year=2200&month=13", nil)
	_, fields := queryYearMonth(req)
	if fields["year"] == "" || fields["month"] == "" {
		t.Fatalf("faltan erros de campos: %#v", fields)
	}

	req = httptest.NewRequest(http.MethodGet, "/api?year=2026&month=10", nil)
	ym, fields := queryYearMonth(req)
	if len(fields) != 0 || ym.Year != 2026 || ym.Month != 10 {
		t.Fatalf("YearMonth inesperado: %#v campos=%#v", ym, fields)
	}
}

func TestRun(t *testing.T) {
	st := testStore(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("non se puido abrir listener: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var out bytes.Buffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, st, ln, Config{NoOpen: true}, &out)
	}()

	url := "http://" + ln.Addr().String() + "/"
	var resp *http.Response
	for i := 0; i < 20; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if err != nil {
		cancel()
		t.Fatalf("non se puido facer GET ao servidor: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("código HTTP inesperado: %d", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("apagado inesperado: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("o servidor non apagou a tempo")
	}
	if !strings.Contains(out.String(), "Local:      http://localhost:") {
		t.Fatalf("banner inesperado: %q", out.String())
	}
}

func TestIsPrivateIPv4(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.10", true},
		{"10.0.0.5", true},
		{"172.16.4.2", true},
		{"8.8.8.8", false},
		{"127.0.0.1", false},
		{"::1", false},
	}
	for _, tc := range cases {
		if got := isPrivateIPv4(net.ParseIP(tc.ip)); got != tc.want {
			t.Fatalf("isPrivateIPv4(%s)=%v, want %v", tc.ip, got, tc.want)
		}
	}
}

func TestMainBadFlag(t *testing.T) {
	if code := Main([]string{"-non-existe"}); code != 2 {
		t.Fatalf("código inesperado para flag inválida: %d", code)
	}
}

func TestWriteHelpers(t *testing.T) {
	rr := httptest.NewRecorder()
	writeFieldErrors(rr, map[string]string{"year": "mal"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("código inesperado: %d", rr.Code)
	}
	var got apiError
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if got.Error != "datos non válidos" || got.Fields["year"] == "" {
		t.Fatalf("resposta inesperada: %#v", got)
	}

	rr = httptest.NewRecorder()
	writeJSON(rr, http.StatusNoContent, map[string]string{"non": "vai"})
	if rr.Body.Len() != 0 {
		t.Fatalf("204 non debe ter corpo: %q", rr.Body.String())
	}
}

func TestQueryYearMonthMissingValues(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	_, fields := queryYearMonth(req)
	if fields["year"] == "" || fields["month"] == "" {
		t.Fatalf("agardábanse erros para parámetros ausentes: %#v", fields)
	}
}
