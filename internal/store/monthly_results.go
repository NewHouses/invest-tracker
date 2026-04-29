package store

import (
	"database/sql"
	"fmt"

	"invest-tracker/internal/domain"
)

func (s *Store) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO monthly_results (asset_id, result_usd, month, year) VALUES (?, ?, ?, ?)`,
		m.AssetID, m.ResultUSD, m.Month, m.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) TotalInvested(assetID int64) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		`SELECT
			(SELECT amount_usd FROM assets WHERE id = ?) +
			COALESCE((SELECT SUM(amount_usd) FROM transactions WHERE asset_id = ?), 0)`,
		assetID, assetID,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
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
