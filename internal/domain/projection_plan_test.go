package domain_test

import (
	"errors"
	"strings"
	"testing"

	"invest-tracker/internal/domain"
)

func planBase() domain.PlanInput {
	return domain.PlanInput{
		Start:               domain.YearMonth{Year: 2026, Month: 1},
		Years:               1,
		InitialInvestment:   0,
		MonthlyReturn:       0,
		Mode:                domain.ProjectionModeContribution,
		MonthlyContribution: 100,
	}
}

func mustProjectPlan(t *testing.T, in domain.PlanInput) []domain.PlanMonth {
	t.Helper()
	months, err := domain.ProjectPlan(in)
	if err != nil {
		t.Fatalf("ProjectPlan: %v", err)
	}
	return months
}

func TestProjectPlan_ContributionConstant(t *testing.T) {
	months := mustProjectPlan(t, planBase())
	if len(months) != 12 {
		t.Fatalf("len=%d, esperabamos 12", len(months))
	}
	for i, m := range months {
		if m.Index != i+1 || !almostEqual(m.Contribution, 100) || !almostEqual(m.TotalInvested, float64(i+1)*100) || !almostEqual(m.TotalCapital, m.TotalInvested) {
			t.Fatalf("mes %d inesperado: %#v", i, m)
		}
		if m.AnnualSalary != 0 || m.MonthlySalary != 0 {
			t.Fatalf("modo aporte non debe ter salario: %#v", m)
		}
	}
}

func TestProjectPlan_FixedYearlyIncrease(t *testing.T) {
	in := planBase()
	in.Years = 2
	in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, Value: 25, EveryMonths: 12, From: domain.YearMonth{Year: 2027, Month: 1}}}
	months := mustProjectPlan(t, in)
	if !almostEqual(months[11].Contribution, 100) || !almostEqual(months[12].Contribution, 125) || !almostEqual(months[23].TotalInvested, 12*100+12*125) {
		t.Fatalf("subida anual fixa inesperada: mes12=%#.2f mes13=%#.2f total=%#.2f", months[11].Contribution, months[12].Contribution, months[23].TotalInvested)
	}
}

func TestProjectPlan_PercentEverySixMonths(t *testing.T) {
	in := planBase()
	in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindPercent, Value: 10, EveryMonths: 6, From: domain.YearMonth{Year: 2026, Month: 7}}}
	months := mustProjectPlan(t, in)
	if !almostEqual(months[5].Contribution, 100) || !almostEqual(months[6].Contribution, 110) {
		t.Fatalf("primeira suba porcentual inesperada: %#v %#v", months[5], months[6])
	}
}

func TestProjectPlan_OnceOnlyAtDate(t *testing.T) {
	in := planBase()
	in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, Value: 50, EveryMonths: 0, From: domain.YearMonth{Year: 2026, Month: 3}}}
	months := mustProjectPlan(t, in)
	if !almostEqual(months[1].Contribution, 100) || !almostEqual(months[2].Contribution, 150) || !almostEqual(months[11].Contribution, 150) {
		t.Fatalf("regra única inesperada: feb=%#.2f mar=%#.2f dec=%#.2f", months[1].Contribution, months[2].Contribution, months[11].Contribution)
	}
}

func TestProjectPlan_TwoRulesSameMonthInOrder(t *testing.T) {
	in := planBase()
	in.Rules = []domain.GrowthRule{
		{Kind: domain.GrowthKindFixed, Value: 10, EveryMonths: 0, From: domain.YearMonth{Year: 2026, Month: 1}},
		{Kind: domain.GrowthKindPercent, Value: 10, EveryMonths: 0, From: domain.YearMonth{Year: 2026, Month: 1}},
	}
	months := mustProjectPlan(t, in)
	if !almostEqual(months[0].Contribution, 121) {
		t.Fatalf("orde de regras inesperada: %.6f", months[0].Contribution)
	}
}

func TestProjectPlan_NegativeValuesClampAtZero(t *testing.T) {
	in := planBase()
	in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, Value: -250, EveryMonths: 0, From: domain.YearMonth{Year: 2026, Month: 2}}}
	months := mustProjectPlan(t, in)
	if !almostEqual(months[0].Contribution, 100) || !almostEqual(months[1].Contribution, 0) || !almostEqual(months[11].Contribution, 0) {
		t.Fatalf("clamp inesperado: jan=%#.2f feb=%#.2f dec=%#.2f", months[0].Contribution, months[1].Contribution, months[11].Contribution)
	}
}

func TestProjectPlan_InitialInvestmentCompounds(t *testing.T) {
	in := planBase()
	in.InitialInvestment = 1000
	in.MonthlyContribution = 0
	in.MonthlyReturn = 0.10
	months := mustProjectPlan(t, in)
	if !almostEqual(months[0].Return, 100) || !almostEqual(months[0].TotalCapital, 1100) || !almostEqual(months[0].TotalGains, 100) || !almostEqual(months[0].TotalInvested, 1000) {
		t.Fatalf("capital inicial inesperado: %#v", months[0])
	}
}

func TestProjectPlan_SalaryModeWithRate(t *testing.T) {
	in := planBase()
	in.Mode = domain.ProjectionModeSalary
	in.MonthlyContribution = 0
	in.AnnualSalary = 120000
	in.InvestmentRate = 0.25
	months := mustProjectPlan(t, in)
	if !almostEqual(months[0].AnnualSalary, 120000) || !almostEqual(months[0].MonthlySalary, 10000) || !almostEqual(months[0].Contribution, 2500) {
		t.Fatalf("modo salario inesperado: %#v", months[0])
	}
}

func TestProjectPlan_YearsLength(t *testing.T) {
	in := planBase()
	in.Years = 3
	months := mustProjectPlan(t, in)
	if len(months) != 36 || months[35].Date != (domain.YearMonth{Year: 2028, Month: 12}) {
		t.Fatalf("lonxitude ou data final inesperada: len=%d last=%#v", len(months), months[35].Date)
	}
}

// Valores que desbordan float64 producirían Inf/NaN: a API devolvía un 200
// baleiro e a gráfica da CLI entraba en pánico.
func TestProjectPlan_OverflowReturnsError(t *testing.T) {
	cases := map[string]func(*domain.PlanInput){
		"retorno desorbitado": func(in *domain.PlanInput) {
			in.Years = 60
			in.InitialInvestment = 1000
			in.MonthlyReturn = 10 // +1000 % ao mes
		},
		"regra fixa enorme": func(in *domain.PlanInput) {
			in.MonthlyContribution = 1e308
			in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, Value: 1e308, EveryMonths: 1, From: in.Start}}
		},
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			in := planBase()
			mut(&in)
			months, err := domain.ProjectPlan(in)
			if !errors.Is(err, domain.ErrProjectionOverflow) {
				t.Fatalf("err=%v, esperabamos ErrProjectionOverflow", err)
			}
			if months != nil {
				t.Fatalf("non debería devolver meses: %d", len(months))
			}
		})
	}
}

func TestPlanInputValidate_Errors(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*domain.PlanInput)
		want string
	}{
		{"start", func(in *domain.PlanInput) { in.Start.Month = 13 }, "inicio"},
		{"years baixo", func(in *domain.PlanInput) { in.Years = 0 }, "anos"},
		{"years alto", func(in *domain.PlanInput) { in.Years = 61 }, "anos"},
		{"initial", func(in *domain.PlanInput) { in.InitialInvestment = -1 }, "investimento inicial"},
		{"return", func(in *domain.PlanInput) { in.MonthlyReturn = -1 }, "retorno mensual"},
		{"mode", func(in *domain.PlanInput) { in.Mode = "malo" }, "modo"},
		{"contribution", func(in *domain.PlanInput) { in.MonthlyContribution = -1 }, "aporte mensual"},
		{"salary", func(in *domain.PlanInput) {
			in.Mode = domain.ProjectionModeSalary
			in.AnnualSalary = -1
			in.InvestmentRate = 0.5
		}, "salario anual"},
		{"rate", func(in *domain.PlanInput) { in.Mode = domain.ProjectionModeSalary; in.InvestmentRate = 0 }, "porcentaxe"},
		{"rules max", func(in *domain.PlanInput) { in.Rules = make([]domain.GrowthRule, domain.MaxGrowthRules+1) }, "regras"},
		{"rule kind", func(in *domain.PlanInput) { in.Rules = []domain.GrowthRule{{Kind: "mala", From: in.Start}} }, "tipo"},
		{"rule percent", func(in *domain.PlanInput) {
			in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindPercent, Value: -100, From: in.Start}}
		}, "valor"},
		{"rule every", func(in *domain.PlanInput) {
			in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, EveryMonths: -1, From: in.Start}}
		}, "periodicidade"},
		{"rule from", func(in *domain.PlanInput) {
			in.Rules = []domain.GrowthRule{{Kind: domain.GrowthKindFixed, From: domain.YearMonth{Year: 0, Month: 1}}}
		}, "data"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := planBase()
			c.mut(&in)
			if err := in.Validate(); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("erro=%v, esperabamos conter %q", err, c.want)
			}
		})
	}
}

func TestDefaultSalaryRaiseRule(t *testing.T) {
	rule := domain.DefaultSalaryRaiseRule(domain.YearMonth{Year: 2026, Month: 8})
	if rule.From != (domain.YearMonth{Year: 2026, Month: 9}) || rule.Kind != domain.GrowthKindPercent || rule.Value != 10 || rule.EveryMonths != 12 {
		t.Fatalf("regra por defecto inesperada: %#v", rule)
	}
	rule = domain.DefaultSalaryRaiseRule(domain.YearMonth{Year: 2026, Month: 9})
	if rule.From != (domain.YearMonth{Year: 2027, Month: 9}) {
		t.Fatalf("setembro inicial debe saltar ao ano seguinte: %#v", rule)
	}
}
