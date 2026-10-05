package store

import (
	"database/sql"
	"errors"
	"fmt"

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

func (s *Store) GetAsset(id int64) (domain.Asset, error) {
	var a domain.Asset
	var t string
	err := s.db.QueryRow(
		`SELECT id, type, name, amount_usd, month, year FROM assets WHERE id = ?`,
		id,
	).Scan(&a.ID, &t, &a.Name, &a.AmountUSD, &a.Month, &a.Year)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Asset{}, fmt.Errorf("activo id=%d: %w", id, sql.ErrNoRows)
		}
		return domain.Asset{}, err
	}
	a.Type = domain.AssetType(t)
	return a, nil
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

// FirstInvestmentMonth devolve a data de inicio dos investimentos, é dicir o
// mes do activo máis antigo rexistrado. ok=false se aínda non hai activos.
func (s *Store) FirstInvestmentMonth() (domain.YearMonth, bool, error) {
	var ym domain.YearMonth
	err := s.db.QueryRow(
		`SELECT year, month FROM assets ORDER BY year, month, id LIMIT 1`,
	).Scan(&ym.Year, &ym.Month)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.YearMonth{}, false, nil
		}
		return domain.YearMonth{}, false, err
	}
	return ym, true, nil
}

// UpdateAsset actualiza nome, cantidade e data dun activo existente.
// Non cambia o tipo. Devolve sql.ErrNoRows envolto se o id non existe.
func (s *Store) UpdateAsset(a domain.Asset) error {
	res, err := s.db.Exec(
		`UPDATE assets SET name = ?, amount_usd = ?, month = ?, year = ? WHERE id = ?`,
		a.Name, a.AmountUSD, a.Month, a.Year, a.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("activo id=%d: %w", a.ID, sql.ErrNoRows)
	}
	return nil
}

// DeleteAsset borra o activo e todas as súas transaccións e resultados mensuais.
// Faino nunha única transacción para que sexa atómico.
func (s *Store) DeleteAsset(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM monthly_results WHERE asset_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM transactions WHERE asset_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM assets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("activo id=%d: %w", id, sql.ErrNoRows)
	}
	return tx.Commit()
}
