package domain

type MonthlyResult struct {
	ID           int64
	InvestmentID int64
	ResultUSD    float64
	Month        int
	Year         int
}
