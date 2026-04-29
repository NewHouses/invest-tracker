package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) InsertAsset(a domain.Asset) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO assets (type, name, amount_usd, month, year) VALUES (?, ?, ?, ?, ?)`,
		string(a.Type), a.Name, a.AmountUSD, a.Month, a.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListAssets() ([]domain.Asset, error) {
	rows, err := s.db.Query(
		`SELECT id, type, name, amount_usd, month, year FROM assets ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Asset
	for rows.Next() {
		var a domain.Asset
		var t string
		if err := rows.Scan(&a.ID, &t, &a.Name, &a.AmountUSD, &a.Month, &a.Year); err != nil {
			return nil, err
		}
		a.Type = domain.AssetType(t)
		out = append(out, a)
	}
	return out, rows.Err()
}
