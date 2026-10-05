package web

import (
	"math"

	"invest-tracker/internal/charts"
	"invest-tracker/internal/closemonth"
	"invest-tracker/internal/domain"
	"invest-tracker/internal/repartoaporte"
	"invest-tracker/internal/viewassethistory"
	"invest-tracker/internal/viewcharts"
	"invest-tracker/internal/viewtotalhistory"
	"invest-tracker/internal/viewtypehistory"
	"invest-tracker/internal/viewtypereport"
)

type apiLineChart struct {
	Title     string             `json:"title"`
	Months    []domain.YearMonth `json:"months"`
	Series    []apiLineSeries    `json:"series"`
	HasAssets bool               `json:"hasAssets"`
}

type apiLineSeries struct {
	Label  string     `json:"label"`
	Values []*float64 `json:"values"`
}

func lineChartDTO(chart viewcharts.LineChart) apiLineChart {
	return apiLineChart{
		Title:     chart.Title,
		Months:    nonNil(chart.Months),
		Series:    lineSeriesDTO(chart.Series),
		HasAssets: chart.HasAssets,
	}
}

func lineSeriesDTO(series []charts.Series) []apiLineSeries {
	out := make([]apiLineSeries, 0, len(series))
	for _, s := range series {
		values := make([]*float64, len(s.Values))
		for i, raw := range s.Values {
			if math.IsNaN(raw) {
				continue
			}
			v := raw
			values[i] = &v
		}
		out = append(out, apiLineSeries{Label: s.Label, Values: values})
	}
	return out
}

func distributionDTO(dist viewcharts.Distribution) viewcharts.Distribution {
	dist.Items = nonNil(dist.Items)
	return dist
}

func typeReportDTO(report viewtypereport.Report) viewtypereport.Report {
	report.Assets = nonNil(report.Assets)
	report.Active = nonNil(report.Active)
	return report
}

func assetHistoryDTO(history viewassethistory.History) viewassethistory.History {
	history.Rows = nonNil(history.Rows)
	return history
}

func typeHistoryDTO(history viewtypehistory.History) viewtypehistory.History {
	history.Rows = nonNil(history.Rows)
	return history
}

func totalHistoryDTO(history viewtotalhistory.History) viewtotalhistory.History {
	history.Rows = nonNil(history.Rows)
	return history
}

func eligibleAssetsDTO(items []closemonth.EligibleAsset) []closemonth.EligibleAsset {
	return nonNil(items)
}

func allocationDTO(allocation repartoaporte.Allocation) repartoaporte.Allocation {
	allocation.Types = nonNil(allocation.Types)
	for i := range allocation.Types {
		allocation.Types[i].Assets = nonNil(allocation.Types[i].Assets)
	}
	return allocation
}
