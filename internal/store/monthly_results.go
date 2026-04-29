package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) InsertMonthlyResult(m domain.MonthlyResult) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO monthly_results (investment_id, result_usd, month, year) VALUES (?, ?, ?, ?)`,
		m.InvestmentID, m.ResultUSD, m.Month, m.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) TotalInvested(investmentID int64) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		`SELECT
			(SELECT amount_usd FROM investments WHERE id = ?) +
			COALESCE((SELECT SUM(amount_usd) FROM transactions WHERE investment_id = ?), 0)`,
		investmentID, investmentID,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
