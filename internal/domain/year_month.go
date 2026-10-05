package domain

// MonthsPerYear é o número de meses dun ano, usado nos cálculos de datas e
// na conversión de salario anual a mensual.
const MonthsPerYear = 12

type YearMonth struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

// Valid indica se o ano e o mes están dentro dos rangos admitidos.
func (ym YearMonth) Valid() bool {
	return ValidYear(ym.Year) && ValidMonth(ym.Month)
}

// Index devolve un índice mensual absoluto útil para comparar datas.
func (ym YearMonth) Index() int {
	return ym.Year*MonthsPerYear + (ym.Month - 1)
}

// Before indica se ym é anterior a o.
func (ym YearMonth) Before(o YearMonth) bool {
	return ym.Index() < o.Index()
}

// AddMonths devolve o YearMonth resultante de sumar n meses (n pode ser
// negativo), normalizando o mes ao rango 1-12 e axustando o ano.
func (ym YearMonth) AddMonths(n int) YearMonth {
	total := ym.Index() + n
	year := total / MonthsPerYear
	month := total % MonthsPerYear
	if month < 0 {
		month += MonthsPerYear
		year--
	}
	return YearMonth{Year: year, Month: month + 1}
}
