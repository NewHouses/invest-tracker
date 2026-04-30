package store

import (
	"invest-tracker/internal/domain"
)

// MonthsWithResults devolve todos os pares (year, month) distintos en
// monthly_results, ordenados cronoloxicamente.
func (s *Store) MonthsWithResults() ([]domain.YearMonth, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT year, month FROM monthly_results ORDER BY year, month`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.YearMonth
	for rows.Next() {
		var ym domain.YearMonth
		if err := rows.Scan(&ym.Year, &ym.Month); err != nil {
			return nil, err
		}
		out = append(out, ym)
	}
	return out, rows.Err()
}

// MonthsWithResultsUpTo devolve os pares (year, month) distintos en
// monthly_results cuxo (year*12+month) <= (year*12+month) do argumento,
// ordenados cronoloxicamente.
func (s *Store) MonthsWithResultsUpTo(year, month int) ([]domain.YearMonth, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT year, month FROM monthly_results
		 WHERE (year * 12 + month) <= (? * 12 + ?)
		 ORDER BY year, month`,
		year, month,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.YearMonth
	for rows.Next() {
		var ym domain.YearMonth
		if err := rows.Scan(&ym.Year, &ym.Month); err != nil {
			return nil, err
		}
		out = append(out, ym)
	}
	return out, rows.Err()
}

// SumDividends devolve a suma de dividends.amount_usd para (year, month).
// Cero se non hai filas (non é erro).
func (s *Store) SumDividends(year, month int) (float64, error) {
	var sum float64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(amount_usd), 0) FROM dividends WHERE year = ? AND month = ?`,
		year, month,
	).Scan(&sum)
	if err != nil {
		return 0, err
	}
	return sum, nil
}
