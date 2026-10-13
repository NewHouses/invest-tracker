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

	"invest-tracker/internal/i18n"
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
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "inicio web") {
			t.Fatalf("agardábase index.html, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Cache-Control"); got != "no-cache" {
			t.Fatalf("cache inesperada para index: %q", got)
		}
	})

	t.Run("activo inmutable", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/assets/app-123.js", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "ola") {
			t.Fatalf("agardábase o activo, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
		if got := rr.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("cache inesperada para activo: %q", got)
		}
	})

	t.Run("ruta de cliente", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/activos/5", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "inicio web") {
			t.Fatalf("agardábase index para a ruta SPA, código=%d corpo=%q", rr.Code, rr.Body.String())
		}
	})

	t.Run("sen listado de directorio", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/assets/", nil))
		if rr.Code == http.StatusOK && strings.Contains(rr.Body.String(), "app-123.js") {
			t.Fatalf("non debe amosar listado de directorio: %q", rr.Body.String())
		}
	})

	// Un ficheiro inexistente non debe caer en index.html: o navegador
	// intentaría executar HTML coma JS.
	for _, p := range []string{"/assets/vello-999.js", "/favicon-inexistente.svg"} {
		t.Run("ficheiro inexistente "+p, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost"+p, nil))
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
	s.handleAPI("GET /api/proba", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"ok": "si"})
	})
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/api/does-not-exist", nil))
	if rr.Code != http.StatusNotFound || !strings.Contains(rr.Body.String(), "recurso non atopado") {
		t.Fatalf("404 API inesperado: código=%d corpo=%q", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/api/proba", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"ok":"si"`) {
		t.Fatalf("API sen contrasinal: código=%d corpo=%q", rr.Code, rr.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	s := testServer(t)
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/", nil))
	for _, name := range []string{"X-Content-Type-Options", "Referrer-Policy", "X-Frame-Options"} {
		if rr.Header().Get(name) == "" {
			t.Fatalf("falta a cabeceira %s", name)
		}
	}
	if rr.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("falta a política CSP nas respostas non API")
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/api/does-not-exist", nil))
	if rr.Header().Get("X-Content-Type-Options") == "" {
		t.Fatal("faltan cabeceiras de seguridade na API")
	}
	if rr.Header().Get("Content-Security-Policy") != "" {
		t.Fatal("a API non debe levar CSP")
	}
}

func TestCrossOriginProtection(t *testing.T) {
	s := testServer(t)
	s.handleAPI("POST /api/publica", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNoContent, nil)
	})
	h := s.Handler()

	req := httptest.NewRequest(http.MethodPost, "http://localhost/api/publica", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("agardábase 403 para petición cross-site, obtido %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "http://localhost/api/publica", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code == http.StatusForbidden {
		t.Fatal("a petición same-origin non debe ser rexeitada")
	}
}

func TestRecoverAPI(t *testing.T) {
	s := testServer(t)
	s.handleAPI("GET /api/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("proba")
	})
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "http://localhost/api/panic", nil))
	if rr.Code != http.StatusInternalServerError || !strings.Contains(rr.Body.String(), "erro interno") {
		t.Fatalf("recuperación inesperada: código=%d corpo=%q", rr.Code, rr.Body.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	t.Run("campo descoñecido", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader(`{"outro":1}`)), &dst)
		if err == nil || !strings.Contains(err.Error(), "campo descoñecido") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})

	t.Run("datos sobrantes", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader(`{"nome":"a"} {}`)), &dst)
		if err == nil || !strings.Contains(err.Error(), "único valor") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})

	t.Run("demasiado grande", func(t *testing.T) {
		var dst struct {
			Nome string `json:"nome"`
		}
		body := `{"nome":"` + strings.Repeat("a", 1<<20) + `"}`
		err := decodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader(body)), &dst)
		if err == nil || !strings.Contains(err.Error(), "1 MiB") {
			t.Fatalf("erro inesperado: %v", err)
		}
	})
}

func TestPathIDAndQueryYearMonth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost/api/items/abc", nil)
	req.SetPathValue("id", "abc")
	if _, err := pathID(req, "id"); err == nil {
		t.Fatal("agardábase erro para id non numérico")
	}

	req = httptest.NewRequest(http.MethodGet, "http://localhost/api/items/42", nil)
	req.SetPathValue("id", "42")
	if id, err := pathID(req, "id"); err != nil || id != 42 {
		t.Fatalf("id inesperado: %d, %v", id, err)
	}

	req = httptest.NewRequest(http.MethodGet, "http://localhost/api?year=2200&month=13", nil)
	_, fields := queryYearMonth(req)
	if fields["year"] == "" || fields["month"] == "" {
		t.Fatalf("faltan erros de campos: %#v", fields)
	}

	req = httptest.NewRequest(http.MethodGet, "http://localhost/api?year=2026&month=10", nil)
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
		errCh <- Run(ctx, st, ln, Config{NoOpen: true, DBPath: "investimentos.db"}, &out)
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
	// A ruta completa da base de datos axuda a ver que datos se están a usar
	// (en macOS, abrir o binario desde o Finder usaría o cartafol persoal).
	if abs, _ := filepath.Abs("investimentos.db"); !strings.Contains(out.String(), "Datos:      "+abs) {
		t.Fatalf("o banner non mostra a ruta da base de datos: %q", out.String())
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
	for lang, want := range map[i18n.Lang]string{i18n.Galician: "datos non válidos", i18n.English: "invalid data"} {
		rr := httptest.NewRecorder()
		writeFieldErrors(rr, lang, map[string]string{"year": "mal"})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("código inesperado: %d", rr.Code)
		}
		var got apiError
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("JSON inválido: %v", err)
		}
		if got.Error != want || got.Fields["year"] == "" {
			t.Fatalf("resposta inesperada en %s: %#v", lang, got)
		}
	}

	rr := httptest.NewRecorder()
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

// Volver abrir a aplicación cando xa está aberta (p.e. dobre clic outra vez no
// acceso directo) só abre o navegador, sen erro de porto ocupado.
func TestMainWhenAlreadyRunningOpensBrowser(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("non se puido abrir listener: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, testStore(t), ln, Config{NoOpen: true}, io.Discard) }()
	t.Cleanup(func() { cancel(); <-done })
	port := listenerPort(ln)
	waitForHTTP(t, "http://127.0.0.1:"+port+"/api/meta")

	var opened []string
	orig := openBrowser
	openBrowser = func(url string) error { opened = append(opened, url); return nil }
	t.Cleanup(func() { openBrowser = orig })

	if code := Main([]string{"-addr", "127.0.0.1:" + port, "-db", filepath.Join(t.TempDir(), "outra.db")}); code != 0 {
		t.Fatalf("código=%d, esperabamos 0", code)
	}
	if len(opened) != 1 || opened[0] != "http://localhost:"+port {
		t.Fatalf("navegador abriu %v, esperabamos http://localhost:%s", opened, port)
	}
}

// Se o porto o ocupa outro programa, segue sendo un erro.
func TestMainPortUsedByAnotherProgram(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("non se puido abrir listener: %v", err)
	}
	srv := &http.Server{Handler: http.NotFoundHandler()}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	waitForHTTP(t, "http://"+ln.Addr().String()+"/")

	orig := openBrowser
	openBrowser = func(string) error { t.Fatal("non debería abrir o navegador"); return nil }
	t.Cleanup(func() { openBrowser = orig })

	if code := Main([]string{"-addr", ln.Addr().String(), "-db", filepath.Join(t.TempDir(), "outra.db")}); code != 1 {
		t.Fatalf("código=%d, esperabamos 1", code)
	}
}

// waitForHTTP agarda a que url responda (con calquera código HTTP).
func waitForHTTP(t *testing.T, url string) {
	t.Helper()
	for i := 0; i < 40; i++ {
		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("%s non respondeu a tempo", url)
}
