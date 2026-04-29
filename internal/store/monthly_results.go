package store

import (
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
