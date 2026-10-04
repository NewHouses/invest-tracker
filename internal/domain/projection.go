package domain

import "errors"

// Parámetros de negocio da proxección de investimentos a longo prazo.
const (
	// ProjectionYears é o horizonte temporal da proxección.
	ProjectionYears = 20

	// ProjectionMonths é o número total de meses simulados (20 anos).
	ProjectionMonths = ProjectionYears * MonthsPerYear

	// SalaryRaiseMonth é o mes no que se aplica a subida salarial (setembro).
	SalaryRaiseMonth = 9

	// SalaryRaiseRate é a subida salarial que se aplica cada setembro (10%),
	// de forma acumulativa sobre o salario anual vixente.
	SalaryRaiseRate = 0.10

	// InvestmentRate é a fracción do salario mensual que se destina a
	// investimento (18.8235%).
	InvestmentRate = 0.188235

	// MinMonthlyReturn é o retorno mensual mínimo (exclusivo) admisible,
	// expresado como fracción: -1.0 equivale a perder o 100% cada mes.
	MinMonthlyReturn = -1.0

	// MinMonthlyReturnPct é MinMonthlyReturn expresado en porcentaxe, pensado
	// para os prompts e as mensaxes de validación.
	MinMonthlyReturnPct = MinMonthlyReturn * 100
)

// Erros de validación de ProjectionInput.
var (
	ErrNegativeSalary = errors.New("o salario anual non pode ser negativo")
	ErrReturnTooLow   = errors.New("o retorno mensual debe ser maior ca -100%")
	ErrInvalidStart   = errors.New("a data de inicio non é válida")
)

// ProjectionInput son os parámetros de entrada da proxección.
type ProjectionInput struct {
	// AnnualSalary é o salario anual inicial en USD (punto de partida).
	AnnualSalary float64

	// MonthlyReturn é o retorno mensual esperado como fracción: 0.01 = 1%.
	MonthlyReturn float64

	// Start é a data de inicio dos investimentos.
	Start YearMonth
}

// Validate comproba que todos os campos obrigatorios son coherentes.
func (in ProjectionInput) Validate() error {
	if in.AnnualSalary < 0 {
		return ErrNegativeSalary
	}
	if in.MonthlyReturn <= MinMonthlyReturn {
		return ErrReturnTooLow
	}
	if in.Start.Month < 1 || in.Start.Month > MonthsPerYear || in.Start.Year < 1 {
		return ErrInvalidStart
	}
	return nil
}

// ProjectionMonth é a foto da carteira ao remate dun mes da proxección.
type ProjectionMonth struct {
	// Index é o número de mes dentro da proxección (1..ProjectionMonths).
	Index int

	// Date é o ano/mes natural ao que corresponde este período.
	Date YearMonth

	// AnnualSalary e MonthlySalary son os salarios vixentes neste mes.
	AnnualSalary  float64
	MonthlySalary float64

	// Investment é o aporte feito neste mes; TotalInvested o acumulado.
	Investment    float64
	TotalInvested float64

	// Return é o rendemento xerado neste mes; TotalGains o acumulado.
	Return     float64
	TotalGains float64

	// TotalCapital é o capital total tras aportar e aplicar o rendemento.
	TotalCapital float64
}

// Project simula a evolución do patrimonio durante ProjectionMonths meses
// desde in.Start.
//
// Cada mes:
//  1. Se é setembro (agás o propio mes de inicio, que é o punto de partida),
//     o salario anual sobe SalaryRaiseRate de forma acumulativa.
//  2. Apórtase InvestmentRate do salario mensual.
//  3. Aplícase MonthlyReturn sobre o capital acumulado xa co aporte incluído,
//     é dicir capitalización composta.
func Project(in ProjectionInput) ([]ProjectionMonth, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	months := make([]ProjectionMonth, 0, ProjectionMonths)
	annualSalary := in.AnnualSalary
	var capital, totalInvested float64

	for i := 0; i < ProjectionMonths; i++ {
		date := in.Start.AddMonths(i)
		// O salario inicial rexe no mes de inicio, así que a primeira subida é
		// a do primeiro setembro posterior a ese mes.
		if i > 0 && date.Month == SalaryRaiseMonth {
			annualSalary *= 1 + SalaryRaiseRate
		}

		monthlySalary := annualSalary / MonthsPerYear
		investment := monthlySalary * InvestmentRate

		// Capitalización composta: o aporte do mes tamén renta este mes.
		base := capital + investment
		monthReturn := base * in.MonthlyReturn
		capital = base + monthReturn
		totalInvested += investment

		months = append(months, ProjectionMonth{
			Index:         i + 1,
			Date:          date,
			AnnualSalary:  annualSalary,
			MonthlySalary: monthlySalary,
			Investment:    investment,
			TotalInvested: totalInvested,
			Return:        monthReturn,
			// Derívanse do capital para que sempre se cumpra
			// TotalCapital = TotalInvested + TotalGains.
			TotalGains:   capital - totalInvested,
			TotalCapital: capital,
		})
	}
	return months, nil
}
