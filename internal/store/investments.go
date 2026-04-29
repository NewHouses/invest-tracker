package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) InsertInvestment(inv domain.Investment) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO investments (type, name, amount_usd, month, year) VALUES (?, ?, ?, ?, ?)`,
		string(inv.Type), inv.Name, inv.AmountUSD, inv.Month, inv.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListInvestments() ([]domain.Investment, error) {
	rows, err := s.db.Query(
		`SELECT id, type, name, amount_usd, month, year FROM investments ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Investment
	for rows.Next() {
		var inv domain.Investment
		var t string
		if err := rows.Scan(&inv.ID, &t, &inv.Name, &inv.AmountUSD, &inv.Month, &inv.Year); err != nil {
			return nil, err
		}
		inv.Type = domain.InvestmentType(t)
		out = append(out, inv)
	}
	return out, rows.Err()
}
