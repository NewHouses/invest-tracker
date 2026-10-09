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

// uniqueResultsIndex garante un único resultado mensual por activo e mes.
const uniqueResultsIndex = "idx_monthly_results_asset_period"

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
	if err := ensureUniqueMonthlyResults(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// ensureUniqueMonthlyResults crea o índice único (activo, ano, mes) de
// monthly_results. As versións anteriores engadían unha fila nova cada vez que
// se corrixía un resultado e todos os cálculos usaban a última (id maior):
// antes de crear o índice consérvase esa e bórranse as substituídas. Só
// escribe na base de datos a primeira vez, mentres o índice non existe.
func ensureUniqueMonthlyResults(db *sql.DB) error {
	var exists int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?`,
		uniqueResultsIndex,
	).Scan(&exists)
	if err != nil || exists > 0 {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(
		`DELETE FROM monthly_results WHERE id NOT IN (
			SELECT MAX(id) FROM monthly_results GROUP BY asset_id, year, month
		)`,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + uniqueResultsIndex +
			` ON monthly_results(asset_id, year, month)`,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Close() error {
	return s.db.Close()
}
