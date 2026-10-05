package domain

type MonthlyResult struct {
	ID        int64   `json:"id"`
	AssetID   int64   `json:"assetId"`
	ResultUSD float64 `json:"resultUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}
