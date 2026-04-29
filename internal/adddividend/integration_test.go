package adddividend_test

import (
	"bufio"
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"invest-tracker/internal/adddividend"
	"invest-tracker/internal/store"
)

func TestRun_EndToEnd_PersistsDividend(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	input := "250.75\n5\n2026\n"
	r := bufio.NewReader(strings.NewReader(input))
	var out bytes.Buffer
	if err := adddividend.Run(r, &out, s); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var (
		gotID     int64
		gotAmount float64
		gotMonth  int
		gotYear   int
	)
	row := db.QueryRow(`SELECT id, amount_usd, month, year FROM dividends`)
	if err := row.Scan(&gotID, &gotAmount, &gotMonth, &gotYear); err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if gotID <= 0 {
		t.Errorf("id = %d, esperabamos > 0", gotID)
	}
	if gotAmount != 250.75 || gotMonth != 5 || gotYear != 2026 {
		t.Errorf("fila = (%d, %v, %d, %d); queremos (>0, 250.75, 5, 2026)",
			gotID, gotAmount, gotMonth, gotYear)
	}
}
