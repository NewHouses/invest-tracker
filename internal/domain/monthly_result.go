package domain

type MonthlyResult struct {
	ID        int64
	AssetID   int64
	ResultUSD float64
	Month     int
	Year      int
}
