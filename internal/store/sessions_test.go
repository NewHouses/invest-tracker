package store_test

import (
	"testing"
	"time"

	"invest-tracker/internal/store"
)

func TestStore_Sessions(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	if err := s.CreateSession("tok1", now.Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession tok1: %v", err)
	}
	valid, err := s.SessionValid("tok1", now)
	if err != nil {
		t.Fatalf("SessionValid antes de caducar: %v", err)
	}
	if !valid {
		t.Fatal("tok1 debería ser válida antes da caducidade")
	}
	valid, err = s.SessionValid("tok1", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("SessionValid no límite: %v", err)
	}
	if valid {
		t.Fatal("tok1 debería ser inválida na caducidade exacta")
	}

	if err := s.DeleteSession("tok1"); err != nil {
		t.Fatalf("DeleteSession existente: %v", err)
	}
	if err := s.DeleteSession("tok1"); err != nil {
		t.Fatalf("DeleteSession idempotente: %v", err)
	}
	valid, err = s.SessionValid("tok1", now)
	if err != nil {
		t.Fatalf("SessionValid borrada: %v", err)
	}
	if valid {
		t.Fatal("tok1 non debería ser válida tras borrar")
	}
}

func TestStore_DeleteAllSessions(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for _, token := range []string{"tok1", "tok2"} {
		if err := s.CreateSession(token, now.Add(time.Hour)); err != nil {
			t.Fatalf("CreateSession %s: %v", token, err)
		}
	}
	if err := s.DeleteAllSessions(); err != nil {
		t.Fatalf("DeleteAllSessions: %v", err)
	}
	for _, token := range []string{"tok1", "tok2"} {
		valid, err := s.SessionValid(token, now)
		if err != nil {
			t.Fatalf("SessionValid %s: %v", token, err)
		}
		if valid {
			t.Fatalf("%s debería estar borrada", token)
		}
	}
}

func TestStore_DeleteSessionsExcept(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	for _, token := range []string{"keep", "drop1", "drop2"} {
		if err := s.CreateSession(token, now.Add(time.Hour)); err != nil {
			t.Fatalf("CreateSession %s: %v", token, err)
		}
	}
	if err := s.DeleteSessionsExcept("keep"); err != nil {
		t.Fatalf("DeleteSessionsExcept: %v", err)
	}
	for _, tc := range []struct {
		token string
		want  bool
	}{
		{"keep", true},
		{"drop1", false},
		{"drop2", false},
	} {
		valid, err := s.SessionValid(tc.token, now)
		if err != nil {
			t.Fatalf("SessionValid %s: %v", tc.token, err)
		}
		if valid != tc.want {
			t.Fatalf("SessionValid(%s) = %v, queremos %v", tc.token, valid, tc.want)
		}
	}
}

func TestStore_DeleteExpiredSessions(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		token string
		exp   time.Time
	}{
		{"old", now.Add(-time.Second)},
		{"limit", now},
		{"new", now.Add(time.Second)},
	}
	for _, tc := range cases {
		if err := s.CreateSession(tc.token, tc.exp); err != nil {
			t.Fatalf("CreateSession %s: %v", tc.token, err)
		}
	}
	n, err := s.DeleteExpiredSessions(now)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	if n != 2 {
		t.Fatalf("DeleteExpiredSessions borrou %d, esperabamos 2", n)
	}
	valid, err := s.SessionValid("new", now)
	if err != nil {
		t.Fatalf("SessionValid new: %v", err)
	}
	if !valid {
		t.Fatal("new debería seguir válida")
	}
}
