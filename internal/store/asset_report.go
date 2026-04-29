package store

import (
	"invest-tracker/internal/domain"
)

func (s *Store) AssetReport(assetID int64) (domain.AssetReport, error) {
	var report domain.AssetReport

	total, err := s.TotalInvested(assetID)
	if err != nil {
		return report, err
	}
	report.TotalInvested = total

	rows, err := s.db.Query(
		`SELECT mr.year, mr.month, mr.result_usd,
			COALESCE(
				(SELECT amount_usd FROM assets WHERE id = mr.asset_id AND year = mr.year AND month = mr.month),
				0
			) +
			COALESCE(
				(SELECT SUM(amount_usd) FROM transactions WHERE asset_id = mr.asset_id AND year = mr.year AND month = mr.month),
				0
			) AS invested_in_month,
			COALESCE(
				(SELECT amount_usd FROM assets WHERE id = mr.asset_id AND (year * 12 + month) <= (mr.year * 12 + mr.month)),
				0
			) +
			COALESCE(
				(SELECT SUM(amount_usd) FROM transactions WHERE asset_id = mr.asset_id AND (year * 12 + month) <= (mr.year * 12 + mr.month)),
				0
			) AS invested_up_to
		FROM monthly_results mr
		WHERE mr.asset_id = ?
		  AND mr.id = (
			SELECT MAX(id) FROM monthly_results
			WHERE asset_id = mr.asset_id AND year = mr.year AND month = mr.month
		  )
		ORDER BY mr.year, mr.month`,
		assetID,
	)
	if err != nil {
		return report, err
	}
	defer rows.Close()

	var sumPct float64
	var pctCount int
	for rows.Next() {
		var row domain.AssetReportRow
		if err := rows.Scan(&row.Year, &row.Month, &row.Result, &row.InvestedInMonth, &row.TotalInvestedUpTo); err != nil {
			return report, err
		}
		row.Gain = row.Result - row.TotalInvestedUpTo
		if row.TotalInvestedUpTo > 0 {
			row.GainPct = row.Gain / row.TotalInvestedUpTo * 100
			row.HasGainPct = true
			sumPct += row.GainPct
			pctCount++
		}
		report.Rows = append(report.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return report, err
	}

	if len(report.Rows) > 0 {
		last := report.Rows[len(report.Rows)-1]
		report.TotalGain = last.Result - report.TotalInvested
		report.HasTotalGain = true
	}
	if pctCount > 0 {
		report.AvgMonthlyIndexPct = sumPct / float64(pctCount)
		report.HasAvgIndex = true
	}
	return report, nil
}
