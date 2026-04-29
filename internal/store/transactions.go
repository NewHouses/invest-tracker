package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) InsertTransaction(t domain.Transaction) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO transactions (investment_id, amount_usd, month, year) VALUES (?, ?, ?, ?)`,
		t.InvestmentID, t.AmountUSD, t.Month, t.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListTransactionsByInvestment(investmentID int64) ([]domain.Transaction, error) {
	rows, err := s.db.Query(
		`SELECT id, investment_id, amount_usd, month, year FROM transactions WHERE investment_id = ? ORDER BY id`,
		investmentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.InvestmentID, &t.AmountUSD, &t.Month, &t.Year); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
