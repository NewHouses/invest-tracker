package domain

type AssetReportRow struct {
	Year              int
	Month             int
	InvestedInMonth   float64
	TotalInvestedUpTo float64
	Result            float64
	Gain              float64
	GainPct           float64
	HasGainPct        bool
}

type AssetReport struct {
	Rows               []AssetReportRow
	TotalInvested      float64
	TotalGain          float64
	HasTotalGain       bool
	AvgMonthlyIndexPct float64
	HasAvgIndex        bool
}
