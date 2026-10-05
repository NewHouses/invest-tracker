package web

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginLimiter(t *testing.T) {
	lim := newLoginLimiter()
	now := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "192.168.1.5:5555"
	for i := 0; i < loginFailureLimit; i++ {
		if ok, _ := lim.allow(req, now); !ok {
			t.Fatalf("bloqueo antes do intento %d", i+1)
		}
		lim.failure(req, now)
	}
	if ok, retry := lim.allow(req, now); ok || retry != loginLockDuration {
		t.Fatalf("allow bloqueado = %v retry=%s", ok, retry)
	}
	if got := retryAfterSeconds(1500 * time.Millisecond); got != 2 {
		t.Fatalf("retryAfterSeconds = %d", got)
	}
	if ok, _ := lim.allow(req, now.Add(loginLockDuration+time.Second)); !ok {
		t.Fatal("debe permitir despois do bloqueo")
	}
	lim.success(req)
	if ok, _ := lim.allow(req, now); !ok {
		t.Fatal("success debe limpar os intentos")
	}
}
