package store

import (
	"database/sql"
	_ "embed"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// dsnParams activa as foreign keys e o busy timeout en cada conexión que abra
// o pool de database/sql; un PRAGMA con db.Exec só afectaría a unha.
const dsnParams = "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?"+dsnParams)
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
