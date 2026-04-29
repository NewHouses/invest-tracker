package domain

type Transaction struct {
	ID           int64
	InvestmentID int64
	AmountUSD    float64
	Month        int
	Year         int
}
