import type { GrowthRule, ProjectionMode, ProjectionRequest } from '@/api/tools'
import type { YearMonth } from '@/api/types'
import { formatNumber, formatUSD, formatYearMonth } from '@/lib/format'
import { addMonths } from '@/lib/yearMonth'
import { parseNumber } from '@/features/shared/formErrors'

export type RuleFrequencyUnit = 'once' | 'months' | 'years'

export type GrowthRuleFormValue = {
  kind: GrowthRule['kind']
  value: number | string
  frequencyUnit: RuleFrequencyUnit
  frequencyCount: number | string
  from: YearMonth | null
}

export type ProjectionFormValues = {
  start: YearMonth | null
  years: number | string
  initialInvestment: number | string
  monthlyReturnPct: number | string
  mode: ProjectionMode
  monthlyContribution: number | string
  annualSalary: number | string
  investmentRatePct: number | string
  rules: GrowthRuleFormValue[]
}

export function defaultProjectionFormValues(start: YearMonth): ProjectionFormValues {
  return {
    start,
    years: 20,
    initialInvestment: 0,
    monthlyReturnPct: 0,
    mode: 'contribution',
    monthlyContribution: 0,
    annualSalary: 0,
    investmentRatePct: 18.8235,
    rules: [],
  }
}

export function emptyRule(start: YearMonth | null, overrides: Partial<GrowthRuleFormValue> = {}): GrowthRuleFormValue {
  return {
    kind: 'fixed',
    value: '',
    frequencyUnit: 'years',
    frequencyCount: 1,
    from: start ? addMonths(start, 12) : null,
    ...overrides,
  }
}

export function ruleToForm(rule: GrowthRule): GrowthRuleFormValue {
  if (rule.everyMonths === 0) {
    return { kind: rule.kind, value: rule.value, frequencyUnit: 'once', frequencyCount: 1, from: rule.from }
  }
  if (rule.everyMonths % 12 === 0) {
    return { kind: rule.kind, value: rule.value, frequencyUnit: 'years', frequencyCount: rule.everyMonths / 12, from: rule.from }
  }
  return { kind: rule.kind, value: rule.value, frequencyUnit: 'months', frequencyCount: rule.everyMonths, from: rule.from }
}

export function formRuleToRule(rule: GrowthRuleFormValue): GrowthRule {
  const frequencyCount = Math.max(1, Math.trunc(parseNumber(rule.frequencyCount)))
  const everyMonths = rule.frequencyUnit === 'once' ? 0 : rule.frequencyUnit === 'years' ? frequencyCount * 12 : frequencyCount
  return {
    kind: rule.kind,
    value: parseNumber(rule.value),
    everyMonths,
    from: rule.from ?? { year: Number.NaN, month: Number.NaN },
  }
}

function recurrenceText(rule: GrowthRule) {
  if (rule.everyMonths === 0) return `unha vez en ${formatYearMonth(rule.from)}`
  if (rule.everyMonths === 12) return `cada ano desde ${formatYearMonth(rule.from)}`
  if (rule.everyMonths % 12 === 0) {
    const years = rule.everyMonths / 12
    return `cada ${formatNumber(years, 0)} anos desde ${formatYearMonth(rule.from)}`
  }
  return `cada ${formatNumber(rule.everyMonths, 0)} meses desde ${formatYearMonth(rule.from)}`
}

export function describeRule(rule: GrowthRuleFormValue) {
  if (!rule.from) return 'Escolle desde que mes se aplica a regra'
  const apiRule = formRuleToRule(rule)
  if (!Number.isFinite(apiRule.value)) return `Indica o valor da regra (${recurrenceText(apiRule)})`
  const sign = apiRule.value < 0 ? '−' : '+'
  const abs = Math.abs(apiRule.value)
  const value = apiRule.kind === 'fixed' ? formatUSD(abs) : `${formatNumber(abs, Number.isInteger(abs) ? 0 : 2)} %`
  return `${sign}${value} ${recurrenceText(apiRule)}`
}

export function buildProjectionRequest(values: ProjectionFormValues): ProjectionRequest {
  return {
    start: values.start ?? { year: Number.NaN, month: Number.NaN },
    years: Math.trunc(parseNumber(values.years)),
    initialInvestment: parseNumber(values.initialInvestment),
    monthlyReturnPct: parseNumber(values.monthlyReturnPct),
    mode: values.mode,
    monthlyContribution: values.mode === 'contribution' ? parseNumber(values.monthlyContribution) : 0,
    annualSalary: values.mode === 'salary' ? parseNumber(values.annualSalary) : 0,
    investmentRatePct: values.mode === 'salary' ? parseNumber(values.investmentRatePct) : 0,
    rules: values.rules.map(formRuleToRule),
  }
}
