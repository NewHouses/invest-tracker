package portfolio

import (
	"fmt"

	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/viewassethistory"
)

type Repo interface {
	ListAssets() ([]domain.Asset, error)
	MonthlySummary(assetID int64, year, month int) (domain.MonthlySummary, error)
	MonthsWithResultsForAsset(assetID int64) ([]domain.YearMonth, error)
}

type Row struct {
	Asset           domain.Asset      `json:"asset"`
	TotalInvested   float64           `json:"totalInvested"`
	CurrentValue    float64           `json:"currentValue"`
	HasCurrentValue bool              `json:"hasCurrentValue"`
	Gain            float64           `json:"gain"`
	GainPct         float64           `json:"gainPct"`
	HasGain         bool              `json:"hasGain"`
	HasGainPct      bool              `json:"hasGainPct"`
	LastResult      *domain.YearMonth `json:"lastResult"`
	Pending         bool              `json:"pending"`
}

type Portfolio struct {
	Period       domain.YearMonth `json:"period"`
	Rows         []Row            `json:"rows"`
	PendingCount int              `json:"pendingCount"`
}

func Build(repo Repo, now domain.YearMonth) (Portfolio, error) {
	assets, err := repo.ListAssets()
	if err != nil {
		return Portfolio{}, fmt.Errorf("listando activos: %w", err)
	}

	eligible, err := closemonth.Eligible(repo, now.Year, now.Month)
	if err != nil {
		return Portfolio{}, err
	}
	pendingByAsset := make(map[int64]bool, len(eligible))
	for _, item := range eligible {
		if !item.HasResult {
			pendingByAsset[item.Asset.ID] = true
		}
	}

	out := Portfolio{Period: now, Rows: make([]Row, 0, len(assets))}
	for _, asset := range assets {
		history, err := viewassethistory.Build(repo, asset)
		if err != nil {
			return Portfolio{}, fmt.Errorf("construíndo carteira de %s: %w", asset.Name, err)
		}

		row := Row{
			Asset:         asset,
			TotalInvested: history.TotalInvested,
			Gain:          history.TotalGain,
			HasGain:       history.HasTotalGain,
			Pending:       pendingByAsset[asset.ID],
		}
		if len(history.Rows) > 0 {
			last := history.Rows[len(history.Rows)-1]
			period := last.Period
			row.CurrentValue = history.CurrentValue
			row.HasCurrentValue = true
			row.LastResult = &period
		}
		if row.HasGain && row.TotalInvested > 0 {
			row.GainPct = row.Gain / row.TotalInvested * 100
			row.HasGainPct = true
		}
		if row.Pending {
			out.PendingCount++
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}
