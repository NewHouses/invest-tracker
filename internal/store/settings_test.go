package store_test

import (
	"testing"

	"invest-tracker/internal/store"
)

func TestStore_Settings(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	if value, ok, err := s.GetSetting("password_hash"); err != nil {
		t.Fatalf("GetSetting inexistente: %v", err)
	} else if ok || value != "" {
		t.Fatalf("GetSetting inexistente = (%q, %v), esperabamos baleiro/false", value, ok)
	}

	if err := s.SetSetting("password_hash", "v1"); err != nil {
		t.Fatalf("SetSetting insert: %v", err)
	}
	value, ok, err := s.GetSetting("password_hash")
	if err != nil {
		t.Fatalf("GetSetting insertado: %v", err)
	}
	if !ok || value != "v1" {
		t.Fatalf("GetSetting insertado = (%q, %v), queremos v1/true", value, ok)
	}

	if err := s.SetSetting("password_hash", "v2"); err != nil {
		t.Fatalf("SetSetting upsert: %v", err)
	}
	value, ok, err = s.GetSetting("password_hash")
	if err != nil {
		t.Fatalf("GetSetting actualizado: %v", err)
	}
	if !ok || value != "v2" {
		t.Fatalf("GetSetting actualizado = (%q, %v), queremos v2/true", value, ok)
	}

	if err := s.DeleteSetting("password_hash"); err != nil {
		t.Fatalf("DeleteSetting existente: %v", err)
	}
	if err := s.DeleteSetting("password_hash"); err != nil {
		t.Fatalf("DeleteSetting idempotente: %v", err)
	}
	if _, ok, err := s.GetSetting("password_hash"); err != nil {
		t.Fatalf("GetSetting borrado: %v", err)
	} else if ok {
		t.Fatal("GetSetting tras borrar devolveu ok=true")
	}
}
