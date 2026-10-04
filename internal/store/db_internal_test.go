package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// Test interno porque o pool de conexións non se ve dende a API pública: as
// foreign keys deben estar activas en tódalas conexións que abra
// database/sql, non só na primeira.
func TestOpen_ForeignKeysOnEveryConnection(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "fk.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	ctx := context.Background()
	// Mentres a primeira conexión estea collida, o pool ten que abrir outra.
	conns := make([]*sql.Conn, 2)
	for i := range conns {
		c, err := s.db.Conn(ctx)
		if err != nil {
			t.Fatalf("Conn[%d]: %v", i, err)
		}
		t.Cleanup(func() { _ = c.Close() })
		conns[i] = c
	}

	for i, c := range conns {
		var on int
		if err := c.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&on); err != nil {
			t.Fatalf("PRAGMA foreign_keys na conexión %d: %v", i, err)
		}
		if on != 1 {
			t.Errorf("conexión %d: foreign_keys = %d, esperabamos 1", i, on)
		}
	}
}
