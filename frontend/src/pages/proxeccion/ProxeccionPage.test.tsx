import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ProxeccionPage } from '@/pages/proxeccion/ProxeccionPage'
import { installFetch, jsonResponse } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

const defaults = {
  start: { year: 2026, month: 1 },
  startFromAssets: true,
  years: 20,
  investmentRatePct: 18.8235,
  salaryRules: [{ kind: 'percent', value: 10, everyMonths: 12, from: { year: 2026, month: 9 } }],
}

function projectionResponse(mode: 'contribution' | 'salary' = 'contribution') {
  return {
    input: {
      start: { year: 2026, month: 1 },
      years: 20,
      initialInvestment: 1000,
      monthlyReturnPct: 1.5,
      mode,
      monthlyContribution: mode === 'contribution' ? 500 : 0,
      annualSalary: mode === 'salary' ? 120000 : 0,
      investmentRatePct: mode === 'salary' ? 18.8235 : 0,
      rules: mode === 'salary' ? defaults.salaryRules : [{ kind: 'fixed', value: 10, everyMonths: 12, from: { year: 2027, month: 1 } }],
    },
    months: [{ index: 1, date: { year: 2026, month: 1 }, annualSalary: mode === 'salary' ? 120000 : 0, monthlySalary: mode === 'salary' ? 10000 : 0, contribution: mode === 'salary' ? 1882.35 : 500, totalInvested: 1500, return: 22.5, totalGains: 22.5, totalCapital: 1522.5 }],
    summary: { initialInvestment: 1000, firstContribution: mode === 'salary' ? 1882.35 : 500, finalContribution: mode === 'salary' ? 1882.35 : 500, finalAnnualSalary: mode === 'salary' ? 120000 : 0, totalContributions: mode === 'salary' ? 1882.35 : 500, totalInvested: 1500, totalGains: 22.5, finalCapital: 1522.5, months: 1 },
  }
}

beforeEach(() => vi.restoreAllMocks())

describe('ProxeccionPage', () => {
  it('prefire os defaults de data e anos', async () => {
    installFetch({ '/api/tools/projection/defaults': defaults })
    renderWithProviders(<ProxeccionPage />)
    expect(await screen.findByText('(ativo máis antigo)')).toBeInTheDocument()
    expect(screen.getByLabelText('Anos a proxectar')).toHaveValue('20')
  })

  it('envía contribution mode con regra e renderiza resumo', async () => {
    let body: unknown = null
    installFetch({
      '/api/tools/projection/defaults': defaults,
      '/api/tools/projection': (init: RequestInit) => { body = JSON.parse(String(init.body)); return projectionResponse('contribution') },
    })
    const user = userEvent.setup()
    renderWithProviders(<ProxeccionPage />)
    await screen.findByText('(ativo máis antigo)')
    await user.type(screen.getByLabelText('Investimento inicial (USD)'), '1000')
    await user.type(screen.getByLabelText('Retorno mensual esperado (%)'), '1,5')
    await user.type(screen.getByLabelText('Aporte mensual (USD)'), '500')
    await user.click(screen.getByRole('button', { name: 'Engadir regra' }))
    await user.type(screen.getByLabelText('Valor'), '10')
    await user.click(screen.getByRole('button', { name: 'Calcular proxección' }))
    await waitFor(() => expect(body).not.toBeNull())
    expect(body).toEqual({
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
    expect(await screen.findByText('Capital final')).toBeInTheDocument()
    expect(screen.getAllByText(/1\.522,50/).length).toBeGreaterThan(0)
  })

  it('envía salary mode con regra salarial predefinida', async () => {
    let body: unknown = null
    installFetch({
      '/api/tools/projection/defaults': defaults,
      '/api/tools/projection': (init: RequestInit) => { body = JSON.parse(String(init.body)); return projectionResponse('salary') },
    })
    const user = userEvent.setup()
    renderWithProviders(<ProxeccionPage />)
    await user.click(await screen.findByText('A partir do salario'))
    expect(screen.getByText('+10 % cada ano desde 09/2026')).toBeInTheDocument()
    await user.type(screen.getByLabelText('Investimento inicial (USD)'), '1000')
    await user.type(screen.getByLabelText('Retorno mensual esperado (%)'), '1,5')
    await user.type(screen.getByLabelText('Salario anual neto inicial (USD)'), '120000')
    await user.click(screen.getByRole('button', { name: 'Calcular proxección' }))
    await waitFor(() => expect(body).not.toBeNull())
    expect(body).toEqual({
      start: { year: 2026, month: 1 },
      years: 20,
      initialInvestment: 1000,
      monthlyReturnPct: 1.5,
      mode: 'salary',
      monthlyContribution: 0,
      annualSalary: 120000,
      investmentRatePct: 18.8235,
      rules: defaults.salaryRules,
    })
  })

  it('mapea erros de campo das regras', async () => {
    vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL, init: RequestInit = {}) => {
      const url = String(input)
      if (url === '/api/tools/projection/defaults') return Promise.resolve(jsonResponse(defaults))
      if (url === '/api/tools/projection') {
        expect(init.body).toBeTruthy()
        return Promise.resolve(jsonResponse({ error: 'datos non válidos', fields: { 'rules[0].value': 'valor inválido' } }, 400))
      }
      return Promise.resolve(jsonResponse({ error: `sen ruta ${url}` }, 404))
    }))
    const user = userEvent.setup()
    renderWithProviders(<ProxeccionPage />)
    await screen.findByText('(ativo máis antigo)')
    await user.type(screen.getByLabelText('Aporte mensual (USD)'), '500')
    await user.click(screen.getByRole('button', { name: 'Engadir regra' }))
    await user.click(screen.getByRole('button', { name: 'Calcular proxección' }))
    expect(await screen.findByText('valor inválido')).toBeInTheDocument()
  })
})
