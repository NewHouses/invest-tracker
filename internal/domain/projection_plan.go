package domain

import (
	"errors"
	"fmt"
	"math"
)

// ErrProjectionOverflow indica que os importes da proxección crecen tanto que
// deixan de ser números finitos (p.e. un retorno mensual desorbitado).
var ErrProjectionOverflow = errors.New("a proxección supera o rango numérico: reduce o retorno, os aportes ou as regras de crecemento")

type ProjectionMode string

type GrowthKind string

const (
	ProjectionModeContribution ProjectionMode = "contribution"
	ProjectionModeSalary       ProjectionMode = "salary"

	GrowthKindFixed   GrowthKind = "fixed"
	GrowthKindPercent GrowthKind = "percent"

	MaxProjectionPlanYears = 60
	MaxGrowthRules         = 20
	MaxRuleEveryMonths     = 1200
)

type GrowthRule struct {
	Kind        GrowthKind `json:"kind"`
	Value       float64    `json:"value"`
	EveryMonths int        `json:"everyMonths"`
	From        YearMonth  `json:"from"`
}

type PlanInput struct {
	Start               YearMonth
	Years               int
	InitialInvestment   float64
	MonthlyReturn       float64
	Mode                ProjectionMode
	MonthlyContribution float64
	AnnualSalary        float64
	InvestmentRate      float64
	Rules               []GrowthRule
}

type PlanMonth struct {
	Index         int       `json:"index"`
	Date          YearMonth `json:"date"`
	AnnualSalary  float64   `json:"annualSalary"`
	MonthlySalary float64   `json:"monthlySalary"`
	Contribution  float64   `json:"contribution"`
	TotalInvested float64   `json:"totalInvested"`
	Return        float64   `json:"return"`
	TotalGains    float64   `json:"totalGains"`
	TotalCapital  float64   `json:"totalCapital"`
}

func (in PlanInput) Validate() error {
	if !in.Start.Valid() {
		return fmt.Errorf("a data de inicio non é válida")
	}
	if in.Years < 1 || in.Years > MaxProjectionPlanYears {
		return fmt.Errorf("os anos deben estar entre 1 e %d", MaxProjectionPlanYears)
	}
	if !finite(in.InitialInvestment) || in.InitialInvestment < 0 {
		return fmt.Errorf("o investimento inicial debe ser unha cantidade non negativa")
	}
	if !finite(in.MonthlyReturn) || in.MonthlyReturn <= MinMonthlyReturn {
		return fmt.Errorf("o retorno mensual debe ser maior ca -100%%")
	}
	if in.Mode != ProjectionModeContribution && in.Mode != ProjectionModeSalary {
		return fmt.Errorf("o modo de proxección non é válido")
	}
	if in.Mode == ProjectionModeContribution {
		if !finite(in.MonthlyContribution) || in.MonthlyContribution < 0 {
			return fmt.Errorf("o aporte mensual debe ser unha cantidade non negativa")
		}
	} else {
		if !finite(in.AnnualSalary) || in.AnnualSalary < 0 {
			return fmt.Errorf("o salario anual debe ser unha cantidade non negativa")
		}
		if !finite(in.InvestmentRate) || in.InvestmentRate <= 0 || in.InvestmentRate > 1 {
			return fmt.Errorf("a porcentaxe de investimento debe estar entre 0 e 100")
		}
	}
	if len(in.Rules) > MaxGrowthRules {
		return fmt.Errorf("non se poden definir máis de %d regras", MaxGrowthRules)
	}
	for i, rule := range in.Rules {
		if rule.Kind != GrowthKindFixed && rule.Kind != GrowthKindPercent {
			return fmt.Errorf("a regra %d ten un tipo de crecemento non válido", i)
		}
		if !finite(rule.Value) || (rule.Kind == GrowthKindPercent && rule.Value <= -100) {
			return fmt.Errorf("a regra %d ten un valor non válido", i)
		}
		if rule.EveryMonths < 0 || rule.EveryMonths > MaxRuleEveryMonths {
			return fmt.Errorf("a regra %d ten unha periodicidade non válida", i)
		}
		if !rule.From.Valid() {
			return fmt.Errorf("a regra %d ten unha data de inicio non válida", i)
		}
	}
	return nil
}

func ProjectPlan(in PlanInput) ([]PlanMonth, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	months := make([]PlanMonth, 0, in.Years*MonthsPerYear)
	base := in.MonthlyContribution
	if in.Mode == ProjectionModeSalary {
		base = in.AnnualSalary
	}
	capital := in.InitialInvestment
	totalInvested := in.InitialInvestment

	for i := 0; i < in.Years*MonthsPerYear; i++ {
		date := in.Start.AddMonths(i)
		for _, rule := range in.Rules {
			if !ruleApplies(rule, date) {
				continue
			}
			switch rule.Kind {
			case GrowthKindFixed:
				base += rule.Value
			case GrowthKindPercent:
				base *= 1 + rule.Value/100
			}
			if base < 0 {
				base = 0
			}
		}

		annualSalary := 0.0
		monthlySalary := 0.0
		contribution := base
		if in.Mode == ProjectionModeSalary {
			annualSalary = base
			monthlySalary = base / MonthsPerYear
			contribution = monthlySalary * in.InvestmentRate
		}

		beforeReturn := capital + contribution
		monthReturn := beforeReturn * in.MonthlyReturn
		capital = beforeReturn + monthReturn
		totalInvested += contribution
		if !finite(base) || !finite(capital) || !finite(totalInvested) {
			return nil, ErrProjectionOverflow
		}

		months = append(months, PlanMonth{
			Index:         i + 1,
			Date:          date,
			AnnualSalary:  annualSalary,
			MonthlySalary: monthlySalary,
			Contribution:  contribution,
			TotalInvested: totalInvested,
			Return:        monthReturn,
			TotalGains:    capital - totalInvested,
			TotalCapital:  capital,
		})
	}
	return months, nil
}

func DefaultSalaryRaiseRule(start YearMonth) GrowthRule {
	year := start.Year
	if start.Month >= SalaryRaiseMonth {
		year++
	}
	return GrowthRule{
		Kind:        GrowthKindPercent,
		Value:       SalaryRaiseRate * 100,
		EveryMonths: MonthsPerYear,
		From:        YearMonth{Year: year, Month: SalaryRaiseMonth},
	}
}

func ruleApplies(rule GrowthRule, date YearMonth) bool {
	dateIndex := date.Index()
	fromIndex := rule.From.Index()
	if dateIndex < fromIndex {
		return false
	}
	if rule.EveryMonths == 0 {
		return dateIndex == fromIndex
	}
	return (dateIndex-fromIndex)%rule.EveryMonths == 0
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
