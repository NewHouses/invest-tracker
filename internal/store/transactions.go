package store

import (
	"database/sql"
	"fmt"

	"invest-tracker/internal/domain"
)

func (s *Store) InsertTransaction(t domain.Transaction) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO transactions (asset_id, amount_usd, month, year) VALUES (?, ?, ?, ?)`,
		t.AssetID, t.AmountUSD, t.Month, t.Year,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListTransactionsByAsset(assetID int64) ([]domain.Transaction, error) {
	rows, err := s.db.Query(
		`SELECT id, asset_id, amount_usd, month, year FROM transactions WHERE asset_id = ? ORDER BY id`,
		assetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.AssetID, &t.AmountUSD, &t.Month, &t.Year); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpdateTransaction actualiza amount_usd, month e year dunha transacción
// existente (o asset_id e o id NON cambian). Devolve sql.ErrNoRows envolto
// se o id non existe.
func (s *Store) UpdateTransaction(t domain.Transaction) error {
	res, err := s.db.Exec(
		`UPDATE transactions SET amount_usd = ?, month = ?, year = ? WHERE id = ?`,
		t.AmountUSD, t.Month, t.Year, t.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("transacción id=%d: %w", t.ID, sql.ErrNoRows)
	}
	return nil
}

func (s *Store) DeleteTransaction(id int64) error {
	res, err := s.db.Exec(`DELETE FROM transactions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("transacción id=%d: %w", id, sql.ErrNoRows)
	}
	return nil
}
