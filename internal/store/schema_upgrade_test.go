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

// oldSchemaSQL é o esquema das primeiras versións: sen settings, sessions nin
// o índice único de monthly_results.
const oldSchemaSQL = `
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

func TestOpen_UpgradesOldSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("sql.Open antigo: %v", err)
	}
	if _, err := db.Exec(oldSchemaSQL); err != nil {
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

// As versións anteriores gardaban varias filas para o mesmo activo e mes ao
// corrixir un resultado. Ao abrir a base de datos consérvase a última (a que
// usaban todos os cálculos) e créase o índice único.
func TestOpen_DeduplicatesMonthlyResults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dupes.db")
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("sql.Open antigo: %v", err)
	}
	if _, err := db.Exec(oldSchemaSQL); err != nil {
		_ = db.Close()
		t.Fatalf("crear esquema antigo: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO assets (type, name, amount_usd, month, year) VALUES ('accion', 'AAPL', 1000, 1, 2026);
INSERT INTO monthly_results (asset_id, result_usd, month, year) VALUES
    (1, 1100, 1, 2026), (1, 1090, 1, 2026), (1, 1105, 1, 2026), (1, 1200, 2, 2026);`); err != nil {
		_ = db.Close()
		t.Fatalf("datos antigos: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("pechar DB antiga: %v", err)
	}

	for i := 0; i < 2; i++ { // a segunda apertura non debe tocar nada
		s, err := store.Open(path)
		if err != nil {
			t.Fatalf("Open #%d: %v", i+1, err)
		}
		got, err := s.ListMonthlyResultsByAsset(1)
		_ = s.Close()
		if err != nil {
			t.Fatalf("ListMonthlyResultsByAsset: %v", err)
		}
		if len(got) != 2 || got[0].ID != 3 || got[0].ResultUSD != 1105 || got[1].ResultUSD != 1200 {
			t.Fatalf("apertura #%d: resultados = %+v, esperabamos o último de 01/2026 (#3, 1105) e o de 02/2026", i+1, got)
		}
	}

	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if _, err := s.InsertMonthlyResult(domain.MonthlyResult{AssetID: 1, ResultUSD: 1110, Month: 1, Year: 2026}); err != nil {
		t.Fatalf("InsertMonthlyResult: %v", err)
	}
	got, err := s.ListMonthlyResultsByAsset(1)
	if err != nil {
		t.Fatalf("ListMonthlyResultsByAsset: %v", err)
	}
	if len(got) != 2 || got[0].ResultUSD != 1110 {
		t.Fatalf("tras corrixir 01/2026 = %+v, esperabamos un só resultado con 1110", got)
	}
}
