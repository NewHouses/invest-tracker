import { describe, expect, it } from 'vitest'

import { buildProjectionRequest, describeRule, formRuleToRule, ruleToForm } from './rules'

describe('projection rules helpers', () => {
  it('converte regras entre API e formulario', () => {
    const apiRule = { kind: 'fixed' as const, value: 10, everyMonths: 12, from: { year: 2027, month: 1 } }
    const formRule = ruleToForm(apiRule)
    expect(formRule).toEqual({ kind: 'fixed', value: 10, frequencyUnit: 'years', frequencyCount: 1, from: { year: 2027, month: 1 } })
    expect(formRuleToRule(formRule)).toEqual(apiRule)
  })

  it('describe regras en galego', () => {
    expect(describeRule({ kind: 'fixed', value: 10, frequencyUnit: 'years', frequencyCount: 1, from: { year: 2027, month: 1 } })).toBe('+10,00 $ cada ano desde 01/2027')
    expect(describeRule({ kind: 'percent', value: 5, frequencyUnit: 'months', frequencyCount: 6, from: { year: 2026, month: 7 } })).toBe('+5 % cada 6 meses desde 07/2026')
    expect(describeRule({ kind: 'fixed', value: 50, frequencyUnit: 'once', frequencyCount: 1, from: { year: 2030, month: 1 } })).toBe('+50,00 $ unha vez en 01/2030')
  })

  // Regresión: unha regra recén engadida (sen valor) amosaba "+NaN $".
  it('describe regras incompletas ou negativas sen NaN', () => {
    expect(describeRule({ kind: 'fixed', value: '', frequencyUnit: 'years', frequencyCount: 1, from: { year: 2027, month: 1 } })).toBe('Indica o valor da regra (cada ano desde 01/2027)')
    expect(describeRule({ kind: 'fixed', value: 10, frequencyUnit: 'years', frequencyCount: 1, from: null })).toBe('Escolle desde que mes se aplica a regra')
    expect(describeRule({ kind: 'fixed', value: -25, frequencyUnit: 'once', frequencyCount: 1, from: { year: 2030, month: 1 } })).toBe('−25,00\u00a0$ unha vez en 01/2030')
    expect(describeRule({ kind: 'percent', value: -2.5, frequencyUnit: 'years', frequencyCount: 2, from: { year: 2028, month: 1 } })).toBe('−2,50 % cada 2 anos desde 01/2028')
  })

  it('constrúe un payload sen claves alleas', () => {
    expect(buildProjectionRequest({
      start: { year: 2026, month: 1 },
      years: '20',
      initialInvestment: '1000',
      monthlyReturnPct: '1,5',
      mode: 'contribution',
      monthlyContribution: '500',
      annualSalary: '120000',
      investmentRatePct: '18,8235',
      rules: [{ kind: 'fixed', value: '10', frequencyUnit: 'years', frequencyCount: 1, from: { year: 2027, month: 1 } }],
    })).toEqual({
      start: { year: 2026, month: 1 },
      years: 20,
      initialInvestment: 1000,
      monthlyReturnPct: 1.5,
      mode: 'contribution',
      monthlyContribution: 500,
      annualSalary: 0,
      investmentRatePct: 0,
      rules: [{ kind: 'fixed', value: 10, everyMonths: 12, from: { year: 2027, month: 1 } }],
    })
  })
})
