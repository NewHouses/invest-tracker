package domain

// MonthsPerYear é o número de meses dun ano, usado nos cálculos de datas e
// na conversión de salario anual a mensual.
const MonthsPerYear = 12

type YearMonth struct {
	Year  int
	Month int
}

// AddMonths devolve o YearMonth resultante de sumar n meses (n pode ser
// negativo), normalizando o mes ao rango 1-12 e axustando o ano.
func (ym YearMonth) AddMonths(n int) YearMonth {
	total := ym.Year*MonthsPerYear + (ym.Month - 1) + n
	year := total / MonthsPerYear
	month := total % MonthsPerYear
	if month < 0 {
		month += MonthsPerYear
		year--
	}
	return YearMonth{Year: year, Month: month + 1}
}
