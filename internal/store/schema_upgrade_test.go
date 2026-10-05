package store_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"invest-tracker/internal/domain"
	"invest-tracker/internal/store"

	_ "modernc.org/sqlite"
)

func TestOpen_UpgradesOldSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("sql.Open antigo: %v", err)
	}
	oldSchema := `
CREATE TABLE assets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    type        TEXT    NOT NULL CHECK (type IN ('accion','indice','copy_trading','fondo')),
    name        TEXT    NOT NULL,
    amount_usd  REAL    NOT NULL,
    month       INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year        INTEGER NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE transactions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id   INTEGER NOT NULL REFERENCES assets(id),
    amount_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_transactions_asset ON transactions(asset_id);
CREATE TABLE monthly_results (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id   INTEGER NOT NULL REFERENCES assets(id),
    result_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_monthly_results_asset ON monthly_results(asset_id);
CREATE TABLE dividends (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    amount_usd REAL    NOT NULL,
    month      INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year       INTEGER NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);`
	if _, err := db.Exec(oldSchema); err != nil {
		_ = db.Close()
		t.Fatalf("crear esquema antigo: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO assets (type, name, amount_usd, month, year) VALUES (?, ?, ?, ?, ?)`,
		string(domain.Accion), "AAPL", 1000.0, 1, 2026,
	); err != nil {
		_ = db.Close()
		t.Fatalf("insert antigo: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("pechar DB antiga: %v", err)
	}

	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open upgrade: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	assets, err := s.ListAssets()
	if err != nil {
		t.Fatalf("ListAssets tras upgrade: %v", err)
	}
	if len(assets) != 1 || assets[0].Name != "AAPL" {
		t.Fatalf("activos tras upgrade = %+v, esperabamos conservar AAPL", assets)
	}
	if err := s.SetSetting("password_hash", "hash"); err != nil {
		t.Fatalf("SetSetting tras upgrade: %v", err)
	}
	value, ok, err := s.GetSetting("password_hash")
	if err != nil {
		t.Fatalf("GetSetting tras upgrade: %v", err)
	}
	if !ok || value != "hash" {
		t.Fatalf("setting tras upgrade = (%q, %v), queremos hash/true", value, ok)
	}
	if err := s.CreateSession("tok", time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("CreateSession tras upgrade: %v", err)
	}
}
