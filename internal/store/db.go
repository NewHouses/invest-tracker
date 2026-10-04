package store

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// fkPragma activa as foreign keys en cada conexión que abra o pool de
// database/sql; un "PRAGMA foreign_keys = ON" con db.Exec só afectaría a unha.
const fkPragma = "_pragma=foreign_keys(1)"

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?"+fkPragma)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
