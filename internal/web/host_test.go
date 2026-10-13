package web

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllowedHost(t *testing.T) {
	cases := map[string]bool{
		"localhost:8080":       true,
		"LOCALHOST":            true,
		"127.0.0.1:8080":       true,
		"192.168.1.20:8080":    true,
		"[::1]:8080":           true,
		"[fe80::1%eth0]:8080":  true,
		"meu-pc:8080":          true,
		"MEU-PC.local:8080":    true,
		"evil.example:8080":    false,
		"localhost.evil.com":   false,
		"meu-pc.evil.com:8080": false,
		"":                     false,
	}
	for host, want := range cases {
		if got := allowedHost(host, "Meu-PC"); got != want {
			t.Errorf("allowedHost(%q) = %v, esperabamos %v", host, got, want)
		}
	}
	if allowedHost("meu-pc", "") {
		t.Error("sen nome de equipo non se debe aceptar ningún nome")
	}
}

// Sen contrasinal, unha web maliciosa que fixese resolver o seu dominio a este
// equipo (DNS rebinding) podería ler e cambiar os datos: só se atenden
// peticións dirixidas a unha IP, a localhost ou ao nome do equipo.
func TestCheckHostRejectsForeignHosts(t *testing.T) {
	s := New(testStore(t), testStaticFS(), Options{Logger: log.New(io.Discard, "", 0), Hostname: "meu-pc"})
	h := s.Handler()

	for _, target := range []string{"http://evil.example/api/meta", "http://evil.example/"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, target, nil))
		if rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "enderezo non permitido") {
			t.Fatalf("%s: código=%d corpo=%q, esperabamos 403", target, rr.Code, rr.Body.String())
		}
	}

	for _, host := range []string{"localhost:8080", "192.168.1.20:8080", "meu-pc:8080"} {
		req := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
		req.Host = host
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Host %s: código=%d corpo=%q, esperabamos 200", host, rr.Code, rr.Body.String())
		}
	}
}
