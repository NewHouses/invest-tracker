package domain

import "math"

// MinYear é o primeiro ano aceptado nas datas introducidas pola aplicación.
const MinYear = 1900

// MaxYear é o último ano aceptado nas datas introducidas pola aplicación.
const MaxYear = 2100

// ValidMonth indica se m representa un mes válido do calendario.
func ValidMonth(m int) bool {
	return m >= 1 && m <= MonthsPerYear
}

// ValidYear indica se y está dentro do rango de anos admitido.
func ValidYear(y int) bool {
	return y >= MinYear && y <= MaxYear
}

// ValidAmount indica se v é unha cantidade positiva e finita.
func ValidAmount(v float64) bool {
	return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0)
}

// ValidNonNegativeAmount indica se v é unha cantidade non negativa e finita.
func ValidNonNegativeAmount(v float64) bool {
	return v >= 0 && !math.IsNaN(v) && !math.IsInf(v, 0)
}
