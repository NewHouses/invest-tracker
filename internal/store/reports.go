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
	var prevYear, prevMonth int
	err = s.db.QueryRow(
		`SELECT result_usd, year, month FROM monthly_results
		 WHERE asset_id = ? AND (year * 12 + month) < (? * 12 + ?)
		 ORDER BY year DESC, month DESC, id DESC LIMIT 1`,
		assetID, year, month,
	).Scan(&prevResult, &prevYear, &prevMonth)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return summary, err
	}
	if prevResult.Valid {
		summary.HasPrevResult = true
		// EstimatedHolding = prev_result + transaccións estritamente posteriores
		// ao mes do prev_result e ata (incluído) o mes target. Isto cobre o caso
		// de transaccións feitas en meses intermedios sen resultado rexistrado
		// (p.e. unha venda total en 05 cando o último resultado é de 04).
		var txGap float64
		err = s.db.QueryRow(
			`SELECT COALESCE(SUM(amount_usd), 0) FROM transactions
			 WHERE asset_id = ?1
			   AND (year * 12 + month) > (?2 * 12 + ?3)
			   AND (year * 12 + month) <= (?4 * 12 + ?5)`,
			assetID, prevYear, prevMonth, year, month,
		).Scan(&txGap)
		if err != nil {
			return summary, err
		}
		summary.EstimatedHolding = prevResult.Float64 + txGap
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
