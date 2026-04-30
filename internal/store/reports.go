package store

import (
	"database/sql"
	"errors"

	"invest-tracker/internal/domain"
)

func (s *Store) MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error) {
	var summary domain.MonthlySummary

	// As transaccións filtran por (year*12+month) >= (year*12+month do asset)
	// para ignorar txs en meses anteriores á data do activo (poderían xurdir
	// tras editar a data do activo a un mes posterior).
	err := s.db.QueryRow(
		`SELECT
			COALESCE(
				(SELECT amount_usd FROM assets
				 WHERE id = ?1 AND (year * 12 + month) <= (?2 * 12 + ?3)),
				0
			) +
			COALESCE(
				(SELECT SUM(amount_usd) FROM transactions
				 WHERE asset_id = ?1
				   AND (year * 12 + month) <= (?2 * 12 + ?3)
				   AND (year * 12 + month) >= (SELECT year * 12 + month FROM assets WHERE id = ?1)),
				0
			)`,
		assetID, year, month,
	).Scan(&summary.TotalInvestedUpTo)
	if err != nil {
		return summary, err
	}

	err = s.db.QueryRow(
		`SELECT
			COALESCE(
				(SELECT amount_usd FROM assets
				 WHERE id = ?1 AND year = ?2 AND month = ?3),
				0
			) +
			COALESCE(
				(SELECT SUM(amount_usd) FROM transactions
				 WHERE asset_id = ?1
				   AND year = ?2 AND month = ?3
				   AND (year * 12 + month) >= (SELECT year * 12 + month FROM assets WHERE id = ?1)),
				0
			)`,
		assetID, year, month,
	).Scan(&summary.InvestedInMonth)
	if err != nil {
		return summary, err
	}

	var prevResult sql.NullFloat64
	err = s.db.QueryRow(
		`SELECT result_usd FROM monthly_results
		 WHERE asset_id = ? AND (year * 12 + month) < (? * 12 + ?)
		 ORDER BY year DESC, month DESC, id DESC LIMIT 1`,
		assetID, year, month,
	).Scan(&prevResult)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return summary, err
	}
	if prevResult.Valid {
		summary.HasPrevResult = true
		summary.EstimatedHolding = prevResult.Float64 + summary.InvestedInMonth
	} else {
		summary.EstimatedHolding = summary.TotalInvestedUpTo
	}

	err = s.db.QueryRow(
		`SELECT result_usd FROM monthly_results
		 WHERE asset_id = ? AND year = ? AND month = ?
		 ORDER BY id DESC LIMIT 1`,
		assetID, year, month,
	).Scan(&summary.Result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			summary.Result = 0
			summary.HasResult = false
			return summary, nil
		}
		return summary, err
	}
	summary.HasResult = true
	return summary, nil
}
