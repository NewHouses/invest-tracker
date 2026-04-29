package store

import (
	"database/sql"
	"fmt"

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

func (s *Store) ListDividends() ([]domain.Dividend, error) {
	rows, err := s.db.Query(
		`SELECT id, amount_usd, month, year FROM dividends ORDER BY year, month, id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Dividend
	for rows.Next() {
		var d domain.Dividend
		if err := rows.Scan(&d.ID, &d.AmountUSD, &d.Month, &d.Year); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) DeleteDividend(id int64) error {
	res, err := s.db.Exec(`DELETE FROM dividends WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("dividendo id=%d: %w", id, sql.ErrNoRows)
	}
	return nil
}
