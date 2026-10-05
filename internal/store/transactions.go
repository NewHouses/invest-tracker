package store

import (
	"database/sql"
	"errors"
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

func (s *Store) GetTransaction(id int64) (domain.Transaction, error) {
	var t domain.Transaction
	err := s.db.QueryRow(
		`SELECT id, asset_id, amount_usd, month, year FROM transactions WHERE id = ?`,
		id,
	).Scan(&t.ID, &t.AssetID, &t.AmountUSD, &t.Month, &t.Year)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Transaction{}, fmt.Errorf("transacción id=%d: %w", id, sql.ErrNoRows)
		}
		return domain.Transaction{}, err
	}
	return t, nil
}

func (s *Store) InsertTransactions(txs []domain.Transaction) ([]int64, error) {
	if len(txs) == 0 {
		return nil, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ids := make([]int64, len(txs))
	for i, t := range txs {
		res, err := tx.Exec(
			`INSERT INTO transactions (asset_id, amount_usd, month, year) VALUES (?, ?, ?, ?)`,
			t.AssetID, t.AmountUSD, t.Month, t.Year,
		)
		if err != nil {
			return nil, err
		}
		ids[i], err = res.LastInsertId()
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
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
