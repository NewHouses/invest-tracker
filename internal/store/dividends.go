package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) InsertDividend(d domain.Dividend) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO dividends (amount_usd, month, year) VALUES (?, ?, ?)`,
		d.AmountUSD, d.Month, d.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
