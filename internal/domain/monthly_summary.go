package domain

type MonthlySummary struct {
	TotalInvestedUpTo float64 `json:"totalInvestedUpTo"`
	InvestedInMonth   float64 `json:"investedInMonth"`
	Result            float64 `json:"result"`
	HasResult         bool    `json:"hasResult"`

	// EstimatedHolding é o valor que se estima ter no activo ANTES da
	// variación de mercado deste mes. Cálculo:
	//   - se hai un resultado mensual rexistrado nun mes anterior:
	//     prev_result + InvestedInMonth
	//   - se non hai resultado anterior: TotalInvestedUpTo (cost basis acumulado)
	EstimatedHolding float64 `json:"estimatedHolding"`

	// HasPrevResult indica se existe un monthly_result anterior a este mes.
	HasPrevResult bool `json:"hasPrevResult"`
}
