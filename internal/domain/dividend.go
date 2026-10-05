package domain

type Dividend struct {
	ID        int64   `json:"id"`
	AmountUSD float64 `json:"amountUsd"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}
