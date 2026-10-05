package domain_test

import (
	"errors"
	"math"
	"testing"

	"invest-tracker/internal/domain"
)

// tol é a tolerancia usada nas comparacións de coma flotante.
const tol = 1e-9

func almostEqual(got, want float64) bool {
	return math.Abs(got-want) <= tol*math.Max(1, math.Abs(want))
}

func baseInput() domain.ProjectionInput {
	return domain.ProjectionInput{
		AnnualSalary:  40000,
		MonthlyReturn: 0.01, // 1% mensual
		Start:         domain.YearMonth{Year: 2026, Month: 1},
	}
}

func mustProject(t *testing.T, in domain.ProjectionInput) []domain.ProjectionMonth {
	t.Helper()
	months, err := domain.Project(in)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	return months
}

func TestProject_LengthAndDates(t *testing.T) {
	months := mustProject(t, baseInput())

	if len(months) != domain.ProjectionMonths {
		t.Fatalf("len = %d, esperabamos %d", len(months), domain.ProjectionMonths)
	}
	if domain.ProjectionMonths != 240 {
		t.Errorf("ProjectionMonths = %d, esperabamos 240 (20 anos)", domain.ProjectionMonths)
	}
	if got, want := months[0].Date, (domain.YearMonth{Year: 2026, Month: 1}); got != want {
		t.Errorf("primeiro mes = %+v, esperabamos %+v", got, want)
	}
	// 01/2026 + 239 meses = 12/2045.
	if got, want := months[239].Date, (domain.YearMonth{Year: 2045, Month: 12}); got != want {
		t.Errorf("último mes = %+v, esperabamos %+v", got, want)
	}
	for i, m := range months {
		if m.Index != i+1 {
			t.Fatalf("months[%d].Index = %d, esperabamos %d", i, m.Index, i+1)
		}
	}
}

func TestProject_SalaryRaisesEverySeptemberCumulative(t *testing.T) {
	months := mustProject(t, baseInput())

	cases := []struct {
		date domain.YearMonth
		want float64
	}{
		{domain.YearMonth{Year: 2026, Month: 1}, 40000}, // punto de partida
		{domain.YearMonth{Year: 2026, Month: 8}, 40000}, // véspera da subida
		{domain.YearMonth{Year: 2026, Month: 9}, 44000}, // +10%
		{domain.YearMonth{Year: 2027, Month: 8}, 44000}, // mantense todo o ano
		{domain.YearMonth{Year: 2027, Month: 9}, 48400}, // +10% acumulativo
		{domain.YearMonth{Year: 2028, Month: 9}, 53240}, // +10% acumulativo
	}
	for _, c := range cases {
		m := findMonth(t, months, c.date)
		if !almostEqual(m.AnnualSalary, c.want) {
			t.Errorf("salario anual en %02d/%d = %.4f, esperabamos %.4f",
				c.date.Month, c.date.Year, m.AnnualSalary, c.want)
		}
	}

	// 20 setembros na proxección → 40000 · 1.1^20.
	want := 40000 * math.Pow(1+domain.SalaryRaiseRate, 20)
	if got := months[len(months)-1].AnnualSalary; !almostEqual(got, want) {
		t.Errorf("salario anual final = %.4f, esperabamos %.4f", got, want)
	}
}

// O mes de inicio é o punto de partida: aínda que sexa setembro non leva suba.
func TestProject_StartMonthSeptemberIsNotRaised(t *testing.T) {
	in := baseInput()
	in.Start = domain.YearMonth{Year: 2026, Month: domain.SalaryRaiseMonth}
	months := mustProject(t, in)

	if got := months[0].AnnualSalary; !almostEqual(got, 40000) {
		t.Errorf("salario no mes de inicio = %.4f, esperabamos 40000", got)
	}
	// A primeira suba é o setembro seguinte (mes 13 da proxección).
	if got, want := months[12].Date, (domain.YearMonth{Year: 2027, Month: 9}); got != want {
		t.Fatalf("months[12].Date = %+v, esperabamos %+v", got, want)
	}
	if got := months[12].AnnualSalary; !almostEqual(got, 44000) {
		t.Errorf("salario tras a primeira suba = %.4f, esperabamos 44000", got)
	}
}

func TestProject_MonthlySalaryAndInvestment(t *testing.T) {
	months := mustProject(t, baseInput())

	for _, m := range months {
		wantMonthly := m.AnnualSalary / domain.MonthsPerYear
		if !almostEqual(m.MonthlySalary, wantMonthly) {
			t.Fatalf("mes %d: salario mensual = %.6f, esperabamos %.6f",
				m.Index, m.MonthlySalary, wantMonthly)
		}
		wantInvest := wantMonthly * domain.InvestmentRate
		if !almostEqual(m.Investment, wantInvest) {
			t.Fatalf("mes %d: investimento = %.6f, esperabamos %.6f",
				m.Index, m.Investment, wantInvest)
		}
	}

	// 40000/12 · 0.188235 = 627.45 exactos.
	if got := months[0].Investment; !almostEqual(got, 627.45) {
		t.Errorf("investimento inicial = %.6f, esperabamos 627.45", got)
	}
	// Tras a suba de setembro: 44000/12 · 0.188235 = 690.195.
	sep := findMonth(t, months, domain.YearMonth{Year: 2026, Month: 9})
	if !almostEqual(sep.Investment, 690.195) {
		t.Errorf("investimento tras a suba = %.6f, esperabamos 690.195", sep.Investment)
	}
}

func TestProject_CompoundsCapitalMonthly(t *testing.T) {
	months := mustProject(t, baseInput())

	// Mes 1: (0 + 627.45) · 1.01 = 633.7245
	if got := months[0].Return; !almostEqual(got, 6.2745) {
		t.Errorf("rendemento do mes 1 = %.6f, esperabamos 6.2745", got)
	}
	if got := months[0].TotalCapital; !almostEqual(got, 633.7245) {
		t.Errorf("capital do mes 1 = %.6f, esperabamos 633.7245", got)
	}
	// Mes 2: (633.7245 + 627.45) · 1.01 = 1273.786245
	if got := months[1].TotalCapital; !almostEqual(got, 1273.786245) {
		t.Errorf("capital do mes 2 = %.6f, esperabamos 1273.786245", got)
	}
	if got := months[1].TotalInvested; !almostEqual(got, 2*627.45) {
		t.Errorf("investido acumulado do mes 2 = %.6f, esperabamos %.6f", got, 2*627.45)
	}
}

func TestProject_CapitalEqualsInvestedPlusGains(t *testing.T) {
	months := mustProject(t, baseInput())

	var invested float64
	for _, m := range months {
		invested += m.Investment
		if !almostEqual(m.TotalInvested, invested) {
			t.Fatalf("mes %d: investido acumulado = %.6f, esperabamos %.6f",
				m.Index, m.TotalInvested, invested)
		}
		if !almostEqual(m.TotalCapital, m.TotalInvested+m.TotalGains) {
			t.Fatalf("mes %d: capital (%.6f) != investido (%.6f) + ganancias (%.6f)",
				m.Index, m.TotalCapital, m.TotalInvested, m.TotalGains)
		}
	}
	if last := months[len(months)-1]; last.TotalCapital <= last.TotalInvested {
		t.Errorf("cun retorno positivo o capital (%.2f) debe superar o investido (%.2f)",
			last.TotalCapital, last.TotalInvested)
	}
}

func TestProject_ZeroReturn_CapitalEqualsInvested(t *testing.T) {
	in := baseInput()
	in.MonthlyReturn = 0
	months := mustProject(t, in)

	for _, m := range months {
		if !almostEqual(m.Return, 0) {
			t.Fatalf("mes %d: rendemento = %.6f, esperabamos 0", m.Index, m.Return)
		}
		if !almostEqual(m.TotalCapital, m.TotalInvested) {
			t.Fatalf("mes %d: capital = %.6f, esperabamos %.6f",
				m.Index, m.TotalCapital, m.TotalInvested)
		}
	}
}

func TestProject_NegativeReturn_ShrinksCapital(t *testing.T) {
	in := baseInput()
	in.MonthlyReturn = -0.5
	months := mustProject(t, in)

	last := months[len(months)-1]
	if last.TotalGains >= 0 {
		t.Errorf("cun retorno negativo as ganancias deben ser negativas, got %.6f", last.TotalGains)
	}
	if last.TotalCapital < 0 {
		t.Errorf("o capital non debería ser negativo con retorno > -100%%, got %.6f", last.TotalCapital)
	}
}

func TestProject_ZeroSalary_IsAllowed(t *testing.T) {
	in := baseInput()
	in.AnnualSalary = 0
	months := mustProject(t, in)

	last := months[len(months)-1]
	if !almostEqual(last.TotalInvested, 0) || !almostEqual(last.TotalCapital, 0) {
		t.Errorf("con salario 0 todo debe ser 0, got investido=%.6f capital=%.6f",
			last.TotalInvested, last.TotalCapital)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := []struct {
		name  string
		mutar func(*domain.ProjectionInput)
		want  error
	}{
		{"salario negativo", func(in *domain.ProjectionInput) { in.AnnualSalary = -1 }, domain.ErrNegativeSalary},
		{"retorno -100%", func(in *domain.ProjectionInput) { in.MonthlyReturn = -1 }, domain.ErrReturnTooLow},
		{"retorno < -100%", func(in *domain.ProjectionInput) { in.MonthlyReturn = -1.5 }, domain.ErrReturnTooLow},
		{"mes 0", func(in *domain.ProjectionInput) { in.Start.Month = 0 }, domain.ErrInvalidStart},
		{"mes 13", func(in *domain.ProjectionInput) { in.Start.Month = 13 }, domain.ErrInvalidStart},
		{"ano 0", func(in *domain.ProjectionInput) { in.Start.Year = 0 }, domain.ErrInvalidStart},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := baseInput()
			c.mutar(&in)
			if err := in.Validate(); !errors.Is(err, c.want) {
				t.Errorf("Validate() = %v, esperabamos %v", err, c.want)
			}
			if _, err := domain.Project(in); !errors.Is(err, c.want) {
				t.Errorf("Project() = %v, esperabamos %v", err, c.want)
			}
		})
	}
}

func TestValidate_AcceptsValidInput(t *testing.T) {
	if err := baseInput().Validate(); err != nil {
		t.Errorf("Validate() erro inesperado: %v", err)
	}
}

func TestProject_DelegatesWithOldAlgorithmValues(t *testing.T) {
	in := domain.ProjectionInput{
		AnnualSalary:  120000,
		MonthlyReturn: 0,
		Start:         domain.YearMonth{Year: 2026, Month: 8},
	}
	months := mustProject(t, in)

	cases := []struct {
		idx        int
		salary     float64
		investment float64
		invested   float64
	}{
		{0, 120000, 1882.35, 1882.35},
		{1, 132000, 2070.585, 3952.935},
		{12, 132000, 2070.585, 26729.37},
		{13, 145200, 2277.6435, 29007.0135},
	}
	for _, c := range cases {
		m := months[c.idx]
		if !almostEqual(m.AnnualSalary, c.salary) || !almostEqual(m.Investment, c.investment) || !almostEqual(m.TotalInvested, c.invested) || !almostEqual(m.TotalCapital, c.invested) {
			t.Fatalf("mes %d inesperado: %#v", c.idx+1, m)
		}
	}
}

func findMonth(t *testing.T, months []domain.ProjectionMonth, date domain.YearMonth) domain.ProjectionMonth {
	t.Helper()
	for _, m := range months {
		if m.Date == date {
			return m
		}
	}
	t.Fatalf("non se atopou o mes %02d/%d na proxección", date.Month, date.Year)
	return domain.ProjectionMonth{}
}
