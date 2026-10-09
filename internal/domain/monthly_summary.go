package domain

// MinHolding é o capital mínimo (medio céntimo) para considerar que un activo
// segue tendo diñeiro investido. Por debaixo só quedan restos de redondeo, como
// os 0,004 USD que deixa unha retirada total calculada a partir dun resultado
// con máis decimais, e eses restos non deben manter o activo pendente de pechar
// cada mes.
const MinHolding = 0.005

// HasHolding indica se un capital estimado é significativo (>= MinHolding).
func HasHolding(v float64) bool {
	return v >= MinHolding
}

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

// CurrentHolding devolve EstimatedHolding, ou 0 se o activo xa non ten
// capital significativo (vendido por completo ou só restos de redondeo).
// Sobre o resumo "de por vida" (ano 9999) é o valor neto actual: o último
// resultado máis os aportes e vendas posteriores.
func (s MonthlySummary) CurrentHolding() float64 {
	if !HasHolding(s.EstimatedHolding) {
		return 0
	}
	return s.EstimatedHolding
}
