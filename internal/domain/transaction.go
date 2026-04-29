package domain

type Transaction struct {
	ID        int64
	AssetID   int64
	AmountUSD float64
	Month     int
	Year      int
}
