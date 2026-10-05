package domain

type Transaction struct {
	ID        int64   `json:"id"`
	AssetID   int64   `json:"assetId"`
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}
