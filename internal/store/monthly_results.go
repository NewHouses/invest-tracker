package store

import (
	"database/sql"
	"fmt"

	"invest-tracker/internal/domain"
)

// upsertMonthlyResultSQL garda o resultado dun activo nun mes. Só pode haber
// un por activo e mes: se xa existía, substitúese o seu valor.
const upsertMonthlyResultSQL = `INSERT INTO monthly_results (asset_id, result_usd, month, year) VALUES (?, ?, ?, ?)
	ON CONFLICT (asset_id, year, month) DO UPDATE SET result_usd = excluded.result_usd
	RETURNING id`

// InsertMonthlyResult garda o resultado mensual dun activo, substituíndo o
// que houbese nese mes. Devolve o id da fila gardada.
func (s *Store) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	var id int64
	err := s.db.QueryRow(upsertMonthlyResultSQL, m.AssetID, m.ResultUSD, m.Month, m.Year).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// InsertMonthlyResults garda varios resultados nunha única transacción, coa
// mesma semántica de substitución ca InsertMonthlyResult.
func (s *Store) InsertMonthlyResults(rs []domain.MonthlyResult) ([]int64, error) {
	if len(rs) == 0 {
		return nil, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ids := make([]int64, len(rs))
	for i, r := range rs {
		if err := tx.QueryRow(upsertMonthlyResultSQL, r.AssetID, r.ResultUSD, r.Month, r.Year).Scan(&ids[i]); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Store) ListMonthlyResultsByAsset(assetID int64) ([]domain.MonthlyResult, error) {
	rows, err := s.db.Query(
		`SELECT id, asset_id, result_usd, month, year FROM monthly_results
		 WHERE asset_id = ? ORDER BY year, month, id`,
		assetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.MonthlyResult
	for rows.Next() {
		var m domain.MonthlyResult
		if err := rows.Scan(&m.ID, &m.AssetID, &m.ResultUSD, &m.Month, &m.Year); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) DeleteMonthlyResult(id int64) error {
	res, err := s.db.Exec(`DELETE FROM monthly_results WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("resultado mensual id=%d: %w", id, sql.ErrNoRows)
	}
	return nil
}

// DeleteMonthlyResultsByMonth borra todos os monthly_results dun (mes, ano).
// Devolve o número de filas afectadas. Cero non é erro: a operación é idempotente.
func (s *Store) DeleteMonthlyResultsByMonth(year, month int) (int64, error) {
	res, err := s.db.Exec(
		`DELETE FROM monthly_results WHERE year = ? AND month = ?`,
		year, month,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
